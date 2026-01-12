package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

// --- НАЛАШТУВАННЯ ---

// Список твоїх гаманців. Скрипт буде обирати випадковий для кожного нового блоку.
var wallets = []string{
	"044cdf0c77956a8672e1e993a19d3c894252ebabb31a51fc6a898492c03d9d66201c7735759111eb2dd25377ac4f89460e9ed01dedb455372b64b7543316fb23a1",
	"04861e89edbe971a4bcbdd113fd6432fc3f0079c82185dec8a640c77e0f8d117d0adce59bc15bd1d5beb85ff7c89a823c8adbd1227ee1ac3008ff2ff7f46f10942",
	"048470a886afeae24bed50e4d103e951e2294db2277e21ceddd3224fd99a99b52a5ef6a8a674ecae52bdc86b93f41fa9f11f1286b2c9311993058068aa8df55f36",
	"042ce761607dd12837e8f3608099578d5119680db732af4aadd072f8b9e8019989d2875e5b5ada9b59ad52d38b258ece06827bd5e0738de8c4e52d14502bce9d1d",
	"0418b493fe77252ca3b455f2af416cced53122e237ffeb6ff1b20d14adcad87099bd1604724ec125be57c4d8c2646bd15143f062542be7b9a60fc2e9c071983f92",
}

const (
	baseURL    = "https://s-hryvnia-1.onrender.com"
	difficulty = "00000" // Кількість нулів
)

// --- СТРУКТУРИ ДАНИХ ---

type Transaction struct {
	From   *string `json:"from"`
	To     string  `json:"to"`
	Amount int     `json:"amount"`
}

type BlockPayload struct {
	PrevHash     string        `json:"prevHash"`
	Transactions []Transaction `json:"transactions"`
	Nonce        int           `json:"nonce"`
	Miner        string        `json:"miner"`
	Reward       int           `json:"reward"`
	Timestamp    int64         `json:"timestamp"`
	Hash         string        `json:"hash"`
}

// Глобальні змінні для статистики
var (
	hashCount uint64
	found     int32
)

func main() {
	fmt.Println("=========================================")
	fmt.Println("🚀 STUDENT HRYVNIA GO-MINER ULTIMATE")
	fmt.Println("⚡️ CPU Cores:", runtime.NumCPU())
	fmt.Println("💼 Wallets loaded:", len(wallets))
	fmt.Println("=========================================")

	rand.Seed(time.Now().UnixNano())

	// Запуск монітора швидкості (Hashrate)
	go speedMonitor()

	for {
		// 1. Отримуємо актуальний хеш
		prevHash := getChainLastHash()
		if prevHash == "" {
			fmt.Println("⚠️ Не вдалося отримати ланцюг. Повторна спроба через 2 сек...")
			time.Sleep(2 * time.Second)
			continue
		}

		// 2. Вибираємо гаманець
		currentWallet := wallets[rand.Intn(len(wallets))]

		// 3. Запускаємо майнінг
		mineBlock(prevHash, currentWallet)

		// Невелика пауза, щоб не заспамити сервер, якщо блоки знаходяться дуже швидко
		// time.Sleep(100 * time.Millisecond)
	}
}

func mineBlock(prevHash string, wallet string) {
	fmt.Printf("\n🔨 Mining new block...\nParent: %s...\nWallet: ...%s\n", prevHash[:15], wallet[len(wallet)-10:])

	atomic.StoreInt32(&found, 0)             // Скидаємо прапорець знайденого блоку
	timestamp := time.Now().UnixNano() / 1e6 // JS Date.now()

	// Підготовка JSON транзакцій (жорстко заданий формат, щоб збігався з JS)
	// В JS: JSON.stringify([{from:null,to:wallet,amount:1}])
	txsJSON := fmt.Sprintf(`[{"from":null,"to":"%s","amount":1}]`, wallet)

	cores := runtime.NumCPU()
	doneChan := make(chan bool)

	// Запуск воркерів
	for i := 0; i < cores; i++ {
		go func(workerID int) {
			// Кожен воркер починає зі свого зміщення і стрибає на крок = кількості ядер
			nonce := workerID

			for atomic.LoadInt32(&found) == 0 {
				// Логіка хешування точно як в Angular:
				// prevHash + transactionsStr + nonce + miner + reward + timestamp
				data := prevHash + txsJSON + strconv.Itoa(nonce) + wallet + "1" + strconv.FormatInt(timestamp, 10)

				hash := sha256.Sum256([]byte(data))
				hashStr := hex.EncodeToString(hash[:])

				atomic.AddUint64(&hashCount, 1)

				// Перевірка складності
				if hashStr[:5] == difficulty {
					if atomic.CompareAndSwapInt32(&found, 0, 1) {
						fmt.Printf("\n✅ \033[32mBLOCK FOUND!\033[0m\nNonce: %d\nHash: %s\n", nonce, hashStr)

						// Формуємо об'єкт для відправки
						txs := []Transaction{{From: nil, To: wallet, Amount: 1}}
						submitBlock(prevHash, txs, nonce, wallet, timestamp, hashStr)
						doneChan <- true
					}
					return
				}

				nonce += cores
			}
		}(i)
	}

	<-doneChan // Чекаємо, поки хтось знайде блок
}

// --- МЕРЕЖЕВІ ФУНКЦІЇ ---

func getChainLastHash() string {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", baseURL+"/chain", nil)

	// Додаємо заголовки, щоб виглядати як браузер
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// Швидкий парсинг JSON без створення структур для всього ланцюга
	var chain []map[string]interface{}
	if err := json.Unmarshal(body, &chain); err != nil {
		return ""
	}

	if len(chain) > 0 {
		return chain[len(chain)-1]["hash"].(string)
	}
	return ""
}

func submitBlock(prev string, txs []Transaction, nonce int, miner string, ts int64, hash string) {
	payload := map[string]interface{}{
		"block": BlockPayload{
			PrevHash:     prev,
			Transactions: txs,
			Nonce:        nonce,
			Miner:        miner,
			Reward:       1,
			Timestamp:    ts,
			Hash:         hash,
		},
	}

	jsonData, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("POST", baseURL+"/submit-block", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (StudentMiner)")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Помилка мережі при відправці: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		fmt.Println("💰 \033[33mСервер прийняв блок! Монети нараховано.\033[0m")
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("❌ Сервер відхилив: %s (Status: %s)\n", string(body), resp.Status)
	}
}

// --- ДОДАТКОВІ ФУНКЦІЇ ---

func speedMonitor() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		count := atomic.SwapUint64(&hashCount, 0)
		if count > 0 {
			// Виводимо хешрейт у тисячах (kH/s) або мільйонах (MH/s)
			if count > 1000000 {
				fmt.Printf("\r🚀 Speed: %.2f MH/s", float64(count)/1000000.0)
			} else {
				fmt.Printf("\r🚀 Speed: %.2f kH/s", float64(count)/1000.0)
			}
		}
	}
}
