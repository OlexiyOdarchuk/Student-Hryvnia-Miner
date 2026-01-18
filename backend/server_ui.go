package backend

import (
	"encoding/json"
	"net/http"
)

const WebUI = `<!DOCTYPE html>
<html lang="uk">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>S-UAH Моніторинг</title>
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA1MTIgNTEyIj48ZGVmcz48bGluZWFyR3JhZGllbnQgaWQ9ImdyYWQiIHgxPSIwJSIgeTE9IjAlIiB4Mj0iMTAwJSIgeTI9IjEwMCUiPjxzdG9wIG9mZnNldD0iMCUiIHN0eWxlPSJzdG9wLWNvbG9yOiM4MThjZjg7c3RvcC1vcGFjaXR5OjEiIC8+PHN0b3Agb2ZZnNldD0iMTAwJSIgc3R5bGU9InN0b3AtY29sb3I6IzRmNDZlNTtzdG9wLW9wYWNpdHk6MSIgLz48L2xpbmVhckdyYWRpZW50PjwvZGVmcz48cGF0aCBmaWxsPSJ1cmwoI2dyYWQpIiBkPSJNMjU2IDBMNTUgMTEwdjI5MmwyMDEgMTEwIDIwMS0xMTBWMTEwTDI1NiAwem0wIDQ2MGwtMTYwLTg4VjE0MmwxNjAgODh2MjMwem0xNjAtODhsLTE2MCA4OFYyMzBsMTYwLTg4djIzMHoiLz48L3N2Zz4=">
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        :root {
            --bg-core: #0a0e1a;
            --bg-glass: rgba(15, 23, 42, 0.75);
            --primary: #818cf8;
            --primary-dark: #4f46e5;
            --accent: #f472b6;
            --neon-cyan: #06b6d4;
            --success: #34d399;
            --warning: #fbbf24;
            --danger: #ef4444;
            --glass-bg: rgba(30, 41, 59, 0.65);
            --glass-border: 1px solid rgba(255, 255, 255, 0.1);
            --glass-shine: 1px solid rgba(255, 255, 255, 0.15);
            --blur: blur(24px);
            --radius: 20px;
            --shadow: 0 10px 40px -10px rgba(0,0,0,0.5);
            --font-main: 'Outfit', sans-serif;
            --font-mono: 'JetBrains Mono', monospace;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; -webkit-tap-highlight-color: transparent; }

        body {
            background: var(--bg-core);
            background-image: 
                radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.15) 0px, transparent 50%),
                radial-gradient(at 100% 100%, rgba(236, 72, 153, 0.15) 0px, transparent 50%),
                radial-gradient(at 50% 50%, rgba(6, 182, 212, 0.1) 0px, transparent 50%);
            color: white;
            font-family: var(--font-main);
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            overflow-x: hidden;
        }

        body::before {
            content: ''; position: fixed; inset: 0;
            background: repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(129, 140, 248, 0.03) 2px, rgba(129, 140, 248, 0.03) 4px);
            pointer-events: none; z-index: 0;
        }

        .header {
            padding: 20px 30px; display: flex; justify-content: space-between; align-items: center;
            background: rgba(15, 23, 42, 0.8); backdrop-filter: blur(20px); border-bottom: var(--glass-border);
            z-index: 50; position: sticky; top: 0;
        }
        .brand {
            font-size: 1.5rem; font-weight: 800; display: flex; align-items: center; gap: 12px;
            background: linear-gradient(135deg, #fff 0%, #a5b4fc 100%);
            -webkit-background-clip: text; -webkit-text-fill-color: transparent;
        }
        .nav-btn {
            padding: 8px 16px; border-radius: 12px; cursor: pointer; color: #94a3b8; font-weight: 600;
            display: flex; gap: 8px; align-items: center; transition: 0.3s; font-size: 0.9rem;
        }
        .nav-btn.active { background: rgba(129, 140, 248, 0.15); color: var(--primary); }

        .main { flex: 1; padding: 30px; display: flex; flex-direction: column; gap: 30px; max-width: 1600px; margin: 0 auto; width: 100%; }
        
        .grid-4 { display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 24px; }
        
        .glass-card {
            background: var(--glass-bg); backdrop-filter: var(--blur);
            border: var(--glass-border); border-top: var(--glass-shine);
            border-radius: var(--radius); padding: 24px;
            box-shadow: var(--shadow); position: relative; overflow: hidden;
        }

        .stat-label { font-size: 0.85rem; color: #94a3b8; font-weight: 600; text-transform: uppercase; letter-spacing: 1px; }
        .stat-value { font-size: 2.2rem; font-weight: 700; font-family: var(--font-mono); margin: 10px 0; }
        .stat-sub { font-size: 0.85rem; color: #64748b; }
        .stat-icon-bg { position: absolute; right: -20px; bottom: -20px; font-size: 6rem; opacity: 0.05; transform: rotate(-15deg); }

        .text-success { color: var(--success); }
        .text-warning { color: var(--warning); }
        .text-cyan { color: var(--neon-cyan); }
        .text-accent { color: var(--accent); }
        .text-primary { color: var(--primary); }

        .chart-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 30px; min-height: 350px; }
        .chart-container { position: relative; width: 100%; height: 300px; }

        /* Table */
        .table-wrap { overflow-x: auto; }
        table { width: 100%; border-collapse: separate; border-spacing: 0 8px; min-width: 700px; }
        th { text-align: left; color: #64748b; font-size: 0.8rem; text-transform: uppercase; padding: 0 20px; }
        td { background: rgba(255,255,255,0.03); padding: 16px 20px; vertical-align: middle; }
        tr td:first-child { border-top-left-radius: 12px; border-bottom-left-radius: 12px; }
        tr td:last-child { border-top-right-radius: 12px; border-bottom-right-radius: 12px; }
        
        .badge { padding: 4px 10px; border-radius: 6px; font-size: 0.7rem; font-weight: 700; letter-spacing: 0.5px; }
        .badge.active { background: rgba(16, 185, 129, 0.15); color: var(--success); }
        .badge.paused { background: rgba(251, 191, 36, 0.15); color: var(--warning); }

        /* Focus Mode */
        #focus-layer {
            position: fixed; inset: 0; z-index: 100;
            background: linear-gradient(135deg, rgba(10, 14, 26, 0.98) 0%, rgba(15, 23, 42, 0.98) 100%);
            display: none; flex-direction: column; align-items: center; justify-content: center;
            overflow-y: auto; padding: 20px;
        }
        #focus-layer.active { display: flex; }

        .zen-ring {
            width: min(40vh, 400px); height: min(40vh, 400px); border-radius: 50%;
            background: radial-gradient(circle, rgba(99,102,241,0.15) 0%, rgba(99,102,241,0.05) 50%, transparent 70%);
            border: 2px solid rgba(129, 140, 248, 0.2);
            display: flex; flex-direction: column; align-items: center; justify-content: center;
            box-shadow: 0 0 100px rgba(99, 102, 241, 0.3);
            margin-bottom: 5vh; animation: pulse 3s infinite ease-in-out;
        }
        @keyframes pulse { 0%,100%{transform:scale(1);} 50%{transform:scale(1.02);} }
        .zen-val { font-size: min(8vh, 4rem); font-weight: 800; font-family: var(--font-mono); 
            background: linear-gradient(135deg, #fff 0%, #a5b4fc 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
        .btn-close { position: absolute; top: 20px; right: 20px; background: rgba(255,255,255,0.1); border: none; color: white; width: 44px; height: 44px; border-radius: 12px; cursor: pointer; font-size: 1.2rem; }

        @media (max-width: 900px) {
            .chart-grid { grid-template-columns: 1fr; }
            .grid-4 { grid-template-columns: 1fr 1fr; }
        }
        @media (max-width: 600px) {
            .grid-4 { grid-template-columns: 1fr; }
            .header { padding: 15px; } .main { padding: 15px; }
            .stat-value { font-size: 2rem; }
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="brand"><i class="fas fa-cube" style="color: var(--primary)"></i> <span>S-UAH WEB</span></div>
        <div style="display:flex; gap:10px;">
            <div class="nav-btn active" onclick="showView('stats')"><i class="fas fa-chart-bar"></i> Статистика</div>
            <div class="nav-btn" onclick="showView('focus')"><i class="fas fa-expand"></i> Фокус</div>
        </div>
    </div>

    <!-- STATISTICS VIEW -->
    <div id="stats-view" class="main">
        <!-- Top Cards -->
        <div class="grid-4">
            <div class="glass-card">
                <i class="fas fa-server stat-icon-bg"></i>
                <div class="stat-label">Активні гаманці</div>
                <div class="stat-value text-primary" id="active-wallets">0</div>
                <div class="stat-sub">Зараз майнять</div>
            </div>
            <div class="glass-card" style="border-top-color: var(--warning)">
                <i class="fas fa-cube stat-icon-bg"></i>
                <div class="stat-label">Сесійні блоки</div>
                <div class="stat-value text-warning" id="sess-blocks">0</div>
                <div class="stat-sub">За цей запуск</div>
            </div>
            <div class="glass-card" style="border-top-color: var(--accent)">
                <i class="fas fa-layer-group stat-icon-bg"></i>
                <div class="stat-label">Загальні блоки</div>
                <div class="stat-value text-accent" id="life-blocks">0</div>
                <div class="stat-sub">За весь час</div>
            </div>
            <div class="glass-card" style="border-top-color: var(--neon-cyan)">
                <i class="fas fa-clock stat-icon-bg"></i>
                <div class="stat-label">Час роботи</div>
                <div class="stat-value text-cyan" id="uptime" style="font-size: 1.8rem">00:00:00</div>
                <div class="stat-sub">Аптайм ноди</div>
            </div>
        </div>

        <!-- Charts -->
        <div class="chart-grid">
            <div class="glass-card">
                <div class="stat-label" style="margin-bottom: 15px;">Розподіл балансу</div>
                <div class="chart-container">
                    <canvas id="balanceChart"></canvas>
                </div>
            </div>
            <div class="glass-card">
                <div class="stat-label" style="margin-bottom: 15px;">Блоки (Загальні)</div>
                <div class="chart-container">
                    <canvas id="blocksChart"></canvas>
                </div>
            </div>
        </div>

        <!-- Table -->
        <div class="glass-card">
            <div class="stat-label" style="margin-bottom: 20px;">Продуктивність гаманців</div>
            <div class="table-wrap">
                <table id="wallet-table"></table>
            </div>
        </div>
    </div>

    <!-- FOCUS VIEW -->
    <div id="focus-layer">
        <button class="btn-close" onclick="closeFocus()"><i class="fas fa-times"></i></button>
        <div class="zen-ring">
            <div class="zen-val" id="f-hash">0.00</div>
            <div style="letter-spacing: 4px; color: #94a3b8; margin-top: 10px; font-weight: 600;">MH/s ШВИДКІСТЬ</div>
        </div>
        <div class="grid-4" style="width: 100%; max-width: 800px; text-align: center; gap: 20px;">
            <div class="glass-card">
                <div class="stat-label">Баланс</div>
                <div class="stat-value text-success" id="f-balance" style="font-size: 2rem">0.00</div>
            </div>
            <div class="glass-card">
                <div class="stat-label">Сесійні блоки</div>
                <div class="stat-value text-warning" id="f-s-blocks" style="font-size: 2rem">0</div>
            </div>
        </div>
    </div>

    <script>
        // Charts
        let bChart, blChart;
        
        function initCharts() {
            const ctx1 = document.getElementById('balanceChart').getContext('2d');
            bChart = new Chart(ctx1, {
                type: 'doughnut',
                data: { labels: [], datasets: [{ data: [], backgroundColor: ['#818cf8', '#34d399', '#f472b6', '#fbbf24', '#06b6d4'], borderWidth: 0 }] },
                options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'right', labels: { color: '#94a3b8' } } } }
            });

            const ctx2 = document.getElementById('blocksChart').getContext('2d');
            blChart = new Chart(ctx2, {
                type: 'bar',
                data: { labels: [], datasets: [{ label: 'Блоки', data: [], backgroundColor: '#818cf8', borderRadius: 4 }] },
                options: { responsive: true, maintainAspectRatio: false, scales: { y: { grid: { color: 'rgba(255,255,255,0.05)' }, ticks: { color: '#64748b' } }, x: { display: false } }, plugins: { legend: { display: false } } }
            });
        }

        function showView(view) {
            if (view === 'focus') {
                document.getElementById('focus-layer').classList.add('active');
                document.documentElement.requestFullscreen().catch(e=>{});
            } else {
                document.getElementById('focus-layer').classList.remove('active');
                if (document.fullscreenElement) document.exitFullscreen().catch(e=>{});
            }
        }
        function closeFocus() { showView('stats'); }

        async function update() {
            try {
                const res = await fetch('/api/stats');
                const data = await res.json();
                
                // Calc stats
                let sessionBlocks = 0, lifetimeBlocks = 0, active = 0;
                let wNames = [], wBalances = [], wBlocks = [];
                let html = '<thead><tr><th>Гаманець</th><th>Адреса</th><th>Статус</th><th>С. Блоки</th><th>З. Блоки</th><th>Баланс</th></tr></thead><tbody>';
                
                (data.wallets || []).forEach(w => {
                    if (w.working) active++;
                    sessionBlocks += w.session_mined;
                    lifetimeBlocks += w.total_mined;
                    wNames.push(w.name);
                    wBalances.push(w.server_balance);
                    wBlocks.push(w.total_mined);
                    
                    const status = w.working ? '<span class="badge active">АКТИВНИЙ</span>' : '<span class="badge paused">ПАУЗА</span>';
                    html += '<tr><td style="font-weight:700">'+w.name+'</td><td style="font-family:monospace; color:#94a3b8">'+w.address.substring(0,8)+'...</td><td>'+status+'</td><td>'+w.session_mined+'</td><td>'+w.total_mined+'</td><td class="text-success font-mono">'+w.server_balance.toFixed(2)+' S-UAH</td></tr>';
                });
                
                if((data.wallets||[]).length === 0) html += '<tr><td colspan="6" style="text-align:center; padding:20px; color:#64748b">Немає гаманців</td></tr>';
                document.getElementById('wallet-table').innerHTML = html + '</tbody>';

                // Update Text
                document.getElementById('active-wallets').innerText = active;
                document.getElementById('sess-blocks').innerText = sessionBlocks;
                document.getElementById('life-blocks').innerText = lifetimeBlocks;
                document.getElementById('uptime').innerText = data.uptime;
                
                document.getElementById('f-hash').innerText = data.hashrate.toFixed(2);
                document.getElementById('f-balance').innerText = data.total_balance.toFixed(2);
                document.getElementById('f-s-blocks').innerText = sessionBlocks;

                // Update Charts
                if (bChart) {
                    bChart.data.labels = wNames;
                    bChart.data.datasets[0].data = wBalances;
                    bChart.update('none');
                }
                if (blChart) {
                    blChart.data.labels = wNames;
                    blChart.data.datasets[0].data = wBlocks;
                    blChart.update('none');
                }

            } catch(e) { console.error(e); }
        }
        
        window.onload = () => { initCharts(); setInterval(update, 1000); update(); };
    </script>
</body>
</html>`

func StartWebServer() {
    http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Access-Control-Allow-Origin", "*")
        
        fullData := GetDashboardData()
        response := map[string]interface{}{
            "hashrate": fullData.Hashrate,
            "total_balance": fullData.TotalBalance,
            "uptime": fullData.Uptime,
            "wallets": fullData.Wallets,
        }
        
        json.NewEncoder(w).Encode(response)
    })

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        w.Write([]byte(WebUI))
    })

    go func() {
        if err := http.ListenAndServe(Config.ServerPort, nil); err != nil {
            PushLog("❌ Web server error: "+err.Error(), "error")
        }
    }()
    PushLog("🌐 API Server running at http://localhost"+Config.ServerPort, "info")
}