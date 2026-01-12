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
	"sync"
	"sync/atomic"
	"time"
)

// --- КОНФІГУРАЦІЯ ---

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
	serverPort = ":8090"
)

// --- СТРУКТУРИ ---

type DashboardData struct {
	Hashrate     float64       `json:"hashrate"`
	TotalBlocks  int           `json:"total_blocks"`
	Uptime       string        `json:"uptime"`
	TotalBalance int           `json:"total_balance"`
	Wallets      []WalletStats `json:"wallets"`
	NewLogs      []LogEntry    `json:"new_logs"`
}

type WalletStats struct {
	Address       string `json:"address"`
	Short         string `json:"short"`
	SessionMined  int    `json:"session_mined"`
	ServerBalance int    `json:"server_balance"`
}

type LogEntry struct {
	ID      int64  `json:"id"`
	Time    string `json:"time"`
	Message string `json:"message"`
	Type    string `json:"type"` // "info", "success", "error"
}

// --- ГЛОБАЛЬНИЙ СТАН ---

var (
	hashCount uint64
	found     int32
	startTime time.Time

	// Статистика
	sessionMined  int
	walletDataMap map[string]*WalletStats
	dataMutex     sync.RWMutex

	// Логи
	logsBuffer []LogEntry
	logsMutex  sync.Mutex
	lastLogID  int64
)

func main() {
	startTime = time.Now()
	walletDataMap = make(map[string]*WalletStats)

	for _, w := range wallets {
		walletDataMap[w] = &WalletStats{
			Address:       w,
			Short:         "..." + w[len(w)-8:],
			SessionMined:  0,
			ServerBalance: 0,
		}
	}

	go startWebServer()

	go speedMonitor()

	go balanceUpdater()

	fmt.Println("==================================================")
	fmt.Printf("🌐 ВЕБІНТФЕЙС: http://localhost%s\n", serverPort)
	fmt.Println("🔨 МАЙНЕР ЗАПУЩЕНО...")
	fmt.Println("==================================================")

	rand.Seed(time.Now().UnixNano())

	for {
		prevHash := getChainLastHash()
		if prevHash == "" {
			pushLog("⚠️ Немає зв'язку з сервером. Рестарт...", "error")
			time.Sleep(2 * time.Second)
			continue
		}

		currentWallet := wallets[rand.Intn(len(wallets))]

		success := mineBlock(prevHash, currentWallet)

		if success {
			dataMutex.Lock()
			sessionMined++
			walletDataMap[currentWallet].SessionMined++
			dataMutex.Unlock()

			go updateSingleBalance(currentWallet)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// --- МАЙНИНГ ---

func mineBlock(prevHash string, wallet string) bool {
	atomic.StoreInt32(&found, 0)
	timestamp := time.Now().UnixNano() / 1e6

	txPart := []byte(`[{"from":null,"to":"` + wallet + `","amount":1}]`)
	minerPart := []byte(wallet)
	rewardPart := []byte("1")
	tsPart := []byte(strconv.FormatInt(timestamp, 10))

	cores := runtime.NumCPU()
	doneChan := make(chan bool)
	isSuccess := false

	for i := 0; i < cores; i++ {
		go func(workerID int) {
			buffer := make([]byte, 0, 512)
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
				atomic.AddUint64(&hashCount, 1)

				if hashArr[0] == 0 && hashArr[1] == 0 && hashArr[2] < 16 {
					if atomic.CompareAndSwapInt32(&found, 0, 1) {
						hashStr := hex.EncodeToString(hashArr[:])
						pushLog(fmt.Sprintf("🔨 Found nonce: %d", nonce), "info")

						if submitBlock(prevHash, wallet, nonce, timestamp, hashStr) {
							isSuccess = true
							pushLog(fmt.Sprintln("💰 Блок зараховано! (+2 S-UAH)"), "success")
						} else {
							pushLog("❌ Сервер відхилив блок", "error")
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

// --- ВЕБ СЕРВЕР ---

func startWebServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(htmlPage))
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		clientLogCursor := int64(0)

		for {
			dataMutex.RLock()

			var walletsExport []WalletStats
			totalBal := 0
			for _, wAddr := range wallets {
				stats := walletDataMap[wAddr]
				totalBal += stats.ServerBalance
				walletsExport = append(walletsExport, *stats)
			}

			var newLogs []LogEntry
			logsMutex.Lock()
			for _, log := range logsBuffer {
				if log.ID > clientLogCursor {
					newLogs = append(newLogs, log)
					clientLogCursor = log.ID
				}
			}

			if len(logsBuffer) > 50 {
				logsBuffer = logsBuffer[len(logsBuffer)-20:]
			}
			logsMutex.Unlock()

			response := DashboardData{
				Hashrate:     globalHashrate,
				TotalBlocks:  sessionMined,
				Uptime:       time.Since(startTime).Round(time.Second).String(),
				TotalBalance: totalBal,
				Wallets:      walletsExport,
				NewLogs:      newLogs,
			}
			dataMutex.RUnlock()

			jsonData, _ := json.Marshal(response)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			w.(http.Flusher).Flush()

			time.Sleep(200 * time.Millisecond)
		}
	})

	http.ListenAndServe(serverPort, nil)
}

// --- АПДЕЙТЕРИ ---

var globalHashrate float64

func speedMonitor() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		c := atomic.SwapUint64(&hashCount, 0)
		dataMutex.Lock()
		globalHashrate = float64(c) / 1000000.0
		dataMutex.Unlock()
	}
}

func pushLog(msg string, lType string) {
	logsMutex.Lock()
	defer logsMutex.Unlock()

	lastLogID++
	entry := LogEntry{
		ID:      lastLogID,
		Time:    time.Now().Format("15:04:05"),
		Message: msg,
		Type:    lType,
	}
	logsBuffer = append(logsBuffer, entry)
}

func balanceUpdater() {
	for {
		for _, w := range wallets {
			updateSingleBalance(w)
			time.Sleep(1 * time.Second) // Пауза між запитами, щоб не банили
		}
		time.Sleep(5 * time.Second)
	}
}

func updateSingleBalance(wallet string) {
	bal := getBalance(wallet)
	dataMutex.Lock()
	if val, ok := walletDataMap[wallet]; ok {
		val.ServerBalance = bal
	}
	dataMutex.Unlock()
}

// --- АПІШКА ---

func submitBlock(prev, wallet string, nonce int, ts int64, hash string) bool {
	payload := map[string]interface{}{
		"block": map[string]interface{}{
			"prevHash": prev, "transactions": []map[string]interface{}{{"from": nil, "to": wallet, "amount": 1}},
			"nonce": nonce, "miner": wallet, "reward": 1, "timestamp": ts, "hash": hash,
		},
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("POST", baseURL+"/submit-block", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200 || resp.StatusCode == 201
}

func getChainLastHash() string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/chain")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var chain []map[string]interface{}
	json.Unmarshal(body, &chain)
	if len(chain) > 0 {
		return chain[len(chain)-1]["hash"].(string)
	}
	return ""
}

func getBalance(addr string) int {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/balance/" + addr)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var data struct {
		Balance int `json:"balance"`
	}
	json.Unmarshal(body, &data)
	return data.Balance
}

// --- FRONTEND ---

const htmlPage = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>S-UAH Miner Pro</title>
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap" rel="stylesheet">
    <style>
        :root { --bg: #0d1117; --card: #161b22; --border: #30363d; --text: #c9d1d9; --accent: #58a6ff; --green: #2ea043; --gold: #e3b341; }
        html, body { height: 100%; margin: 0; padding: 0; background: var(--bg); color: var(--text); font-family: 'JetBrains Mono', monospace; }
        body { display: flex; justify-content: center; padding: 20px; box-sizing: border-box; }
        .container { max-width: 900px; width: 100%; }
        header { display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--border); padding-bottom: 20px; margin-bottom: 20px; }
        h1 { margin: 0; font-size: 1.5rem; display: flex; align-items: center; gap: 10px; }
        .badge { background: var(--border); padding: 4px 8px; border-radius: 4px; font-size: 0.8rem; color: var(--accent); }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 15px; margin-bottom: 30px; }
        .card { background: var(--card); border: 1px solid var(--border); padding: 20px; border-radius: 8px; position: relative; overflow: hidden; }
        .card h3 { margin: 0 0 10px 0; font-size: 0.9rem; color: #8b949e; text-transform: uppercase; }
        .card .value { font-size: 1.8rem; font-weight: bold; }
        .card.glow { box-shadow: 0 0 15px rgba(88, 166, 255, 0.1); border-color: var(--accent); }
        .table-container { background: var(--card); border: 1px solid var(--border); border-radius: 8px; overflow: hidden; margin-bottom: 30px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px 15px; text-align: left; border-bottom: 1px solid var(--border); }
        th { background: #21262d; font-size: 0.85rem; color: #8b949e; }
        tr:last-child td { border-bottom: none; }
        .w-addr { color: var(--accent); }
        .w-bal { color: var(--gold); font-weight: bold; }
        .w-new { animation: flash 1s ease; }
        .terminal { background: #090c10; border: 1px solid var(--border); border-radius: 8px; padding: 15px; height: 300px; overflow-y: auto; font-size: 0.85rem; }
        .log-row { margin-bottom: 4px; display: flex; gap: 10px; opacity: 0; animation: fadeIn 0.3s forwards; }
        .log-time { color: #8b949e; min-width: 70px; }
        .log-msg { flex-grow: 1; }
        .type-info { color: var(--text); }
        .type-success { color: var(--green); }
        .type-error { color: #da3633; }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
        @keyframes flash { 0% { background-color: rgba(46, 160, 67, 0.2); } 100% { background-color: transparent; } }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>⚡ S-UAH Miner PRO<span class="badge">by iShawyha</span></h1>
            <div id="uptime">0s</div>
        </header>
        <div class="grid">
            <div class="card glow"><h3>ШВИДКІСТЬ МАЙНУ:</h3><div class="value" id="hashrate">0.00 MH/s</div></div>
            <div class="card"><h3>БЛОКІВ ЗА СЕСІЮ:</h3><div class="value" id="blocks" style="color: var(--green)">0</div></div>
            <div class="card"><h3>ЗАГАЛЬНИЙ БАЛАНС:</h3><div class="value w-bal" id="balance">0 S-UAH</div></div>
        </div>
        <div class="table-container">
            <table><thead><tr><th>Гаманці</th><th>Блоків за сесію</th><th>Баланс</th></tr></thead><tbody id="wallet-list"></tbody></table>
        </div>
        <div class="terminal" id="terminal"><div class="log-row"><span class="log-time">System</span><span class="log-msg">Waiting for connection...</span></div></div>
    </div>
    <script>
        const es = new EventSource("/events");
        const walletList = document.getElementById('wallet-list');
        const terminal = document.getElementById('terminal');
        es.onmessage = (e) => {
            const data = JSON.parse(e.data);
            document.getElementById('hashrate').innerText = data.hashrate.toFixed(2) + " MH/s";
            document.getElementById('blocks').innerText = data.total_blocks;
            document.getElementById('balance').innerText = data.total_balance + " S-UAH";
            document.getElementById('uptime').innerText = data.uptime;
            data.wallets.forEach(w => {
                let row = document.getElementById('w-' + w.address);
                if (!row) {
                    row = document.createElement('tr'); row.id = 'w-' + w.address;
                    row.innerHTML = '<td class="w-addr">' + w.short + '</td><td id="sess-' + w.address + '">0</td><td id="bal-' + w.address + '" class="w-bal">0</td>';
                    walletList.appendChild(row);
                }
                const sessCell = document.getElementById('sess-' + w.address);
                const balCell = document.getElementById('bal-' + w.address);
                if (sessCell.innerText != w.session_mined) {
                    sessCell.innerText = w.session_mined; sessCell.classList.remove('w-new'); void sessCell.offsetWidth; sessCell.classList.add('w-new');
                }
                if (balCell.innerText != w.server_balance + " S-UAH") { balCell.innerText = w.server_balance + " S-UAH"; }
            });
            if (data.new_logs && data.new_logs.length > 0) {
                data.new_logs.forEach(log => {
                    const div = document.createElement('div'); div.className = 'log-row';
                    div.innerHTML = '<span class="log-time">' + log.time + '</span><span class="log-msg type-' + log.type + '">' + log.message + '</span>';
                    terminal.prepend(div); 
                });
                while(terminal.children.length > 50) terminal.removeChild(terminal.lastChild);
            }
        };
    </script>
</body>
</html>
`
