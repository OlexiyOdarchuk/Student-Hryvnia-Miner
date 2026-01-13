package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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
			for _, wAddr := range getWallets() {
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

	pushLog("🌐 Вебсервер запущено на http://localhost"+Config.ServerPort, "info")
	if err := http.ListenAndServe(Config.ServerPort, nil); err != nil {
		pushLog("❌ Помилка вебсервера: "+err.Error(), "error")
	}
}

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
