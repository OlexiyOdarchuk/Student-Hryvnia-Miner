package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func startWebServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(htmlPage))
	})

	http.HandleFunc("/api/hashrate-history", func(w http.ResponseWriter, r *http.Request) {
		hashrateHistMutex.Lock()
		history := make([]float64, len(hashrateHistory))
		copy(history, hashrateHistory[:])
		hashrateHistMutex.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"history":  history,
			"position": hashrateHistPos,
		})
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		clientLogCursor := int64(0)
		ctx := r.Context()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			dataMutex.RLock()

			var walletsExport []WalletStats
			totalBal := 0
			for _, wAddr := range getWallets() {
				stats := walletDataMap[wAddr]
				totalBal += stats.ServerBalance
				walletsExport = append(walletsExport, *stats)
			}

			var newLogs []LogEntry
			logRing.mu.Lock()
			for i := 0; i < logRing.pos && i < 100; i++ {
				idx := i
				if i < logRing.pos-100 {
					continue
				}
				log := logRing.data[idx%100]
				if log.ID > clientLogCursor {
					newLogs = append(newLogs, log)
					clientLogCursor = log.ID
				}
			}
			logRing.mu.Unlock()

			hashRate := float64(0)
			if val := globalHashrate.Load(); val != nil {
				hashRate = val.(float64)
			}

			response := DashboardData{
				Hashrate:     hashRate,
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

			select {
			case <-time.After(200 * time.Millisecond):
			case <-ctx.Done():
				return
			}
		}
	})

	pushLog("🌐 Вебсервер запущено на http://localhost"+Config.ServerPort, "info")
	if err := http.ListenAndServe(Config.ServerPort, nil); err != nil {
		pushLog("❌ Помилка вебсервера: "+err.Error(), "error")
	}
}

const htmlPage = `
<!DOCTYPE html>
<html lang="uk">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>⚡ S-UAH Miner PRO</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap" rel="stylesheet">
    <style>
        :root { --dark-bg: #0d1117; --dark-card: #161b22; --dark-border: #30363d; --dark-text: #c9d1d9; --accent: #58a6ff; --green: #2ea043; --gold: #e3b341; --red: #da3633; --bg: var(--dark-bg); --card: var(--dark-card); --border: var(--dark-border); --text: var(--dark-text); }
        html, body { height: 100%; margin: 0; padding: 0; background: var(--bg); color: var(--text); font-family: 'JetBrains Mono', monospace; }
        body { padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        header { display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--border); padding-bottom: 20px; margin-bottom: 20px; }
        h1 { margin: 0; font-size: 1.8rem; display: flex; align-items: center; gap: 10px; }
        .badge { background: var(--border); padding: 6px 12px; border-radius: 6px; font-size: 0.8rem; color: var(--accent); }
        .top-info { display: flex; gap: 20px; align-items: center; }
        .status { display: flex; gap: 5px; align-items: center; font-size: 0.9rem; }
        .status-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--green); animation: pulse 2s infinite; }
        .status-dot.offline { background: var(--red); animation: none; }
        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin-bottom: 30px; }
        .card { background: var(--card); border: 1px solid var(--border); padding: 20px; border-radius: 8px; }
        .card h3 { margin: 0 0 10px 0; font-size: 0.85rem; color: #8b949e; text-transform: uppercase; letter-spacing: 1px; }
        .card .value { font-size: 2rem; font-weight: bold; margin: 5px 0; }
        .card .subtext { font-size: 0.85rem; color: #8b949e; }
        .card.glow { box-shadow: 0 0 20px rgba(88, 166, 255, 0.15); border-color: var(--accent); }
        .section { margin-bottom: 30px; }
        .section h2 { font-size: 1rem; color: #8b949e; text-transform: uppercase; letter-spacing: 1px; margin: 20px 0 15px 0; padding-bottom: 10px; border-bottom: 1px solid var(--border); }
        .chart-container { background: var(--card); border: 1px solid var(--border); padding: 20px; border-radius: 8px; margin-bottom: 20px; height: 250px; position: relative; }
        .table-container { background: var(--card); border: 1px solid var(--border); border-radius: 8px; overflow: hidden; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px 15px; text-align: left; border-bottom: 1px solid var(--border); }
        th { background: #21262d; font-size: 0.8rem; color: #8b949e; font-weight: 600; text-transform: uppercase; }
        tr:last-child td { border-bottom: none; }
        .w-addr { color: var(--accent); font-weight: 600; }
        .w-bal { color: var(--gold); font-weight: bold; }
        .w-anim { animation: flash 0.8s ease; }
        .terminal { background: #090c10; border: 1px solid var(--border); border-radius: 8px; padding: 15px; height: 350px; overflow-y: auto; font-size: 0.8rem; }
        .log-row { margin-bottom: 5px; display: flex; gap: 10px; opacity: 0; animation: slideIn 0.4s ease forwards; font-size: 0.85rem; }
        .log-time { color: #8b949e; min-width: 75px; flex-shrink: 0; }
        .log-msg { flex-grow: 1; white-space: pre-wrap; word-break: break-word; }
        .type-info { color: var(--text); }
        .type-success { color: var(--green); font-weight: bold; }
        .type-error { color: var(--red); font-weight: bold; }
        @keyframes slideIn { from { opacity: 0; transform: translateX(-10px); } to { opacity: 1; transform: translateX(0); } }
        @keyframes flash { 0% { background-color: rgba(46, 160, 67, 0.3); } 100% { background-color: transparent; } }
        .footer { text-align: center; margin-top: 40px; padding-top: 20px; border-top: 1px solid var(--border); color: #8b949e; font-size: 0.8rem; }
        @media (max-width: 768px) { .grid { grid-template-columns: 1fr; } h1 { font-size: 1.3rem; } .top-info { flex-direction: column; align-items: flex-start; } }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>⚡ S-UAH Miner PRO<span class="badge">by iShawyha</span></h1>
            <div class="top-info">
                <div class="status"><div class="status-dot" id="status-dot"></div><span id="status">Connecting...</span></div>
                <div id="uptime" style="color: #8b949e; font-size: 0.9rem;">Uptime: 0s</div>
            </div>
        </header>

        <div class="grid">
            <div class="card glow"><h3>📊 Швидкість</h3><div class="value" id="hashrate">0.00</div><div class="subtext">MH/s</div></div>
            <div class="card"><h3>🔗 Блоків сесії</h3><div class="value" style="color: var(--green);" id="blocks">0</div><div class="subtext">з момента запуску</div></div>
            <div class="card"><h3>💰 Баланс</h3><div class="value" style="color: var(--gold);" id="balance">0</div><div class="subtext">S-UAH</div></div>
            <div class="card"><h3>⏱️ Частота</h3><div class="value" id="frequency">0.00</div><div class="subtext">block/час</div></div>
        </div>

        <div class="section">
            <h2>📈 Графік хешрейту (останні 60 сек)</h2>
            <div class="chart-container">
                <canvas id="hashChart" style="height: 100%; position: relative;"></canvas>
            </div>
        </div>

        <div class="section">
            <h2>💼 Статистика гаманців</h2>
            <div class="table-container">
                <table><thead><tr><th>Адреса</th><th>Блоків сесії</th><th>Баланс</th><th>Статус</th></tr></thead><tbody id="wallet-list"></tbody></table>
            </div>
        </div>

        <div class="section">
            <h2>📜 Лог подій</h2>
            <div class="terminal" id="terminal"></div>
        </div>

        <div class="footer">🚀 S-Hryvnia Miner PRO | Real-time Blockchain Mining Dashboard</div>
    </div>

    <script>
        let chart = null;
        const es = new EventSource("/events");
        const walletList = document.getElementById('wallet-list');
        const terminal = document.getElementById('terminal');
        let connectedTime = Date.now();

        function initChart() {
            const ctx = document.getElementById('hashChart').getContext('2d');
            if (chart) chart.destroy();
            chart = new Chart(ctx, {
                type: 'line',
                data: { labels: Array(60).fill(''), datasets: [{ label: 'MH/s', data: Array(60).fill(0), borderColor: '#58a6ff', backgroundColor: 'rgba(88, 166, 255, 0.1)', borderWidth: 2, fill: true, tension: 0.1, pointRadius: 0, pointHoverRadius: 3 }] },
                options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false } }, scales: { y: { beginAtZero: true, grid: { color: '#30363d' }, ticks: { color: '#8b949e' } }, x: { grid: { display: false }, ticks: { display: false } } } }
            });
        }

        es.onmessage = (e) => {
            try {
                const data = JSON.parse(e.data);
                document.getElementById('hashrate').innerText = data.hashrate.toFixed(2);
                document.getElementById('blocks').innerText = data.total_blocks;
                document.getElementById('balance').innerText = data.total_balance;
                document.getElementById('uptime').innerText = 'Uptime: ' + data.uptime;
                document.getElementById('status-dot').classList.remove('offline');
                document.getElementById('status').innerText = '🟢 Online';

                const blockRate = (data.total_blocks / Math.max(1, (Date.now() - connectedTime) / 3600000)).toFixed(2);
                document.getElementById('frequency').innerText = blockRate;

                if (!chart) initChart();
                fetch('/api/hashrate-history').then(r => r.json()).then(d => {
                    if (chart && d.history) {
                        chart.data.datasets[0].data = d.history.slice(-60);
                        chart.update('none');
                    }
                });

                data.wallets && data.wallets.forEach(w => {
                    let row = document.getElementById('w-' + w.address);
                    if (!row) {
                        row = document.createElement('tr');
                        row.id = 'w-' + w.address;
                        row.innerHTML = '<td class="w-addr">' + (w.address || w.short || '?').substring(0, 10) + '...</td><td id="sess-' + w.address + '">0</td><td id="bal-' + w.address + '" class="w-bal">0</td><td id="stat-' + w.address + '">✓</td>';
                        walletList.appendChild(row);
                    }
                    const s = document.getElementById('sess-' + w.address);
                    const b = document.getElementById('bal-' + w.address);
                    if (s && s.innerText != w.session_mined) { s.innerText = w.session_mined; s.classList.add('w-anim'); setTimeout(() => s.classList.remove('w-anim'), 800); }
                    if (b && b.innerText != w.server_balance) { b.innerText = w.server_balance; b.classList.add('w-anim'); setTimeout(() => b.classList.remove('w-anim'), 800); }
                });

                if (data.new_logs && data.new_logs.length > 0) {
                    data.new_logs.forEach(log => {
                        const div = document.createElement('div');
                        div.className = 'log-row';
                        div.innerHTML = '<span class="log-time">' + log.time + '</span><span class="log-msg type-' + log.type + '">' + log.message + '</span>';
                        terminal.insertBefore(div, terminal.firstChild);
                        if (terminal.children.length > 100) terminal.removeChild(terminal.lastChild);
                    });
                }
            } catch (e) { console.error('Parse error:', e); }
        };

        es.onerror = () => {
            document.getElementById('status-dot').classList.add('offline');
            document.getElementById('status').innerText = '🔴 Offline';
        };

        initChart();
    </script>
</body>
</html>
`
