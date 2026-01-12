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
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

var wallets = []string{
	"044cdf0c77956a8672e1e993a19d3c894252ebabb31a51fc6a898492c03d9d66201c7735759111eb2dd25377ac4f89460e9ed01dedb455372b64b7543316fb23a1",
	"04861e89edbe971a4bcbdd113fd6432fc3f0079c82185dec8a640c77e0f8d117d0adce59bc15bd1d5beb85ff7c89a823c8adbd1227ee1ac3008ff2ff7f46f10942",
	"048470a886afeae24bed50e4d103e951e2294db2277e21ceddd3224fd99a99b52a5ef6a8a674ecae52bdc86b93f41fa9f11f1286b2c9311993058068aa8df55f36",
	"042ce761607dd12837e8f3608099578d5119680db732af4aadd072f8b9e8019989d2875e5b5ada9b59ad52d38b258ece06827bd5e0738de8c4e52d14502bce9d1d",
	"0418b493fe77252ca3b455f2af416cced53122e237ffeb6ff1b20d14adcad87099bd1604724ec125be57c4d8c2646bd15143f062542be7b9a60fc2e9c071983f92",
}

const (
	baseURL    = "https://s-hryvnia-1.onrender.com"
	difficulty = "00000"
)

// --- СТРУКТУРИ ---

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

type BalanceResponse struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

// --- ГЛОБАЛЬНІ ЗМІННІ ---

var (
	hashCount         uint64
	found             int32
	startTime         time.Time
	sessionRewards    map[string]int // Скільки заробив кожен гаманець за сесію
	totalMinedSession int            // Всього блоків за сесію
)

func main() {
	startTime = time.Now()
	sessionRewards = make(map[string]int)
	rand.Seed(time.Now().UnixNano())

	for _, w := range wallets {
		sessionRewards[w] = 0
	}

	printDashboard()

	// Запуск монітора швидкості (ломає табличку)
	//go speedMonitor()

	for {
		prevHash := getChainLastHash()
		if prevHash == "" {
			fmt.Println("⚠️ Не вийшло отримати блокчейн. Чекаємо...")
			time.Sleep(2 * time.Second)
			continue
		}

		currentWallet := wallets[rand.Intn(len(wallets))]

		success := mineBlock(prevHash, currentWallet)

		if success {
			totalMinedSession++
			sessionRewards[currentWallet]++

			printDashboard()
		}

		time.Sleep(500 * time.Millisecond)
	}
}
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// --- ТАБЛИЧКА ---
func printDashboard() {
	clearScreen()
	uptime := time.Since(startTime).Truncate(time.Second)

	fmt.Println("================================================================")
	fmt.Println("🚀 Майнер студентської гривні | v2.0")
	fmt.Println("================================================================")
	fmt.Printf("⚡️ CPU Ядра: %d | ⏱ Час роботи: %s\n", runtime.NumCPU(), uptime)
	fmt.Printf("⛏  Добуто блоків за сесію: \033[32m%d\033[0m (Total Earned: %d S-UAH)\n", totalMinedSession, totalMinedSession*2)
	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("%-10s | %-15s | %-12s | %-10s\n", "ГАМАНЕЦЬ", "АДРЕСА (Кінець)", "СЕСІЯ", "БАЛАНС")
	fmt.Println("----------------------------------------------------------------")

	totalServerBalance := 0

	for i, w := range wallets {
		shortAddr := "..." + w[len(w)-10:]
		minedNow := sessionRewards[w]

		serverBal := getBalance(w)
		totalServerBalance += serverBal

		fmt.Printf("Wallet #%d  | %s | +%-11d | %d S-UAH\n", i+1, shortAddr, minedNow, serverBal)
	}
	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("💰 В ЗАГАЛЬНОМУ НА БАЛАНСІ: \033[33m%d S-UAH\033[0m\n", totalServerBalance)
	fmt.Println("================================================================")
	fmt.Println("\n🔨 Майниться...")
}

func mineBlock(prevHash string, wallet string) bool {
	atomic.StoreInt32(&found, 0)
	timestamp := time.Now().UnixNano() / 1e6

	txPart := []byte(`[{"from":null,"to":"` + wallet + `","amount":1}]`)

	minerPart := []byte(wallet)

	rewardPart := []byte("1")

	tsPart := []byte(strconv.FormatInt(timestamp, 10))

	baseLen := len(prevHash) + len(txPart) + len(minerPart) + len(rewardPart) + len(tsPart)

	cores := runtime.NumCPU()
	doneChan := make(chan bool)
	var isSuccess bool = false

	for i := 0; i < cores; i++ {
		go func(workerID int) {
			buffer := make([]byte, 0, baseLen+20)
			nonce := workerID

			for atomic.LoadInt32(&found) == 0 {
				buffer = buffer[:0]

				buffer = append(buffer, prevHash...)
				buffer = append(buffer, txPart...)
				buffer = strconv.AppendInt(buffer, int64(nonce), 10)
				buffer = append(buffer, minerPart...)
				buffer = append(buffer, rewardPart...)
				buffer = append(buffer, tsPart...)

				hashArr := sha256.Sum256(buffer)

				if hashArr[0] == 0 && hashArr[1] == 0 && hashArr[2] < 16 {
					if atomic.CompareAndSwapInt32(&found, 0, 1) {
						fullHash := hex.EncodeToString(hashArr[:])

						txs := []Transaction{{From: nil, To: wallet, Amount: 1}}
						if submitBlock(prevHash, txs, nonce, wallet, timestamp, fullHash) {
							isSuccess = true
						}
						doneChan <- true
					}
					return
				}

				nonce += cores
			}
		}(i)
	}

	<-doneChan
	return isSuccess
}

// --- API ЗАПИТИ ---

func getBalance(address string) int {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/balance/" + address)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var data BalanceResponse
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &data)

	return data.Balance
}

func getChainLastHash() string {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", baseURL+"/chain", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var chain []map[string]interface{}
	if err := json.Unmarshal(body, &chain); err != nil {
		return ""
	}
	if len(chain) > 0 {
		return chain[len(chain)-1]["hash"].(string)
	}
	return ""
}

func submitBlock(prev string, txs []Transaction, nonce int, miner string, ts int64, hash string) bool {
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

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200 || resp.StatusCode == 201
}

// func speedMonitor() {
// 	ticker := time.NewTicker(1 * time.Second)
// 	for range ticker.C {
// 		count := atomic.SwapUint64(&hashCount, 0)
// 		if count > 0 {
// 			fmt.Printf("\rSpeed: %.2f MH/s", float64(count)/1000000.0)
// 		}
// 	}
// }
