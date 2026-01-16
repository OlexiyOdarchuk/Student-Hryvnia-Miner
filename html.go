package main

const htmlPage = `<!DOCTYPE html>
<html lang="uk">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>⚡ S-UAH MINER</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <link href="https://fonts.googleapis.com/css2?family=Courier+Prime:wght@400;700&display=swap" rel="stylesheet">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        html, body { width: 100%; height: 100%; font-family: 'Courier Prime', monospace; background: linear-gradient(135deg, #0a0e27 0%, #1a1f3a 30%, #2a0845 60%, #0f1425 100%); color: #ffffff; overflow-x: hidden; }
        body { padding: 15px; }
        .container { max-width: 1400px; margin: 0 auto; }
        
        header { text-align: center; margin-bottom: 30px; padding: 20px 15px; background: linear-gradient(135deg, rgba(26, 31, 58, 0.95) 0%, rgba(42, 8, 69, 0.8) 100%); border: 2px solid rgba(88, 166, 255, 0.4); border-radius: 10px; box-shadow: 0 0 40px rgba(88, 166, 255, 0.15), 0 0 20px rgba(138, 43, 226, 0.1); position: relative; }
        header::before { content: ''; position: absolute; top: 0; left: 0; right: 0; height: 2px; background: linear-gradient(90deg, transparent, #58a6ff, #8a2be2, #58a6ff, transparent); }
        
        h1 { font-size: clamp(1.8em, 5vw, 3em); background: linear-gradient(135deg, #58a6ff 0%, #79c0ff 25%, #8a2be2 50%, #ff6b9d 75%, #58a6ff 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; margin: 10px 0; letter-spacing: 2px; font-weight: 700; }
        .subtitle { font-size: clamp(0.75em, 2vw, 0.95em); color: #a8d8ff; letter-spacing: 1px; margin-top: 8px; }
        
        .header-controls { display: flex; gap: 10px; justify-content: center; margin-top: 15px; flex-wrap: wrap; }
        .btn { padding: 8px 16px; background: linear-gradient(135deg, #8a2be2 0%, #5c2d91 100%); border: 1px solid rgba(138, 43, 226, 0.6); border-radius: 6px; color: #ffffff; cursor: pointer; font-weight: 600; transition: all 0.3s ease; font-family: 'Courier Prime', monospace; }
        .btn:hover { border-color: #8a2be2; box-shadow: 0 0 15px rgba(138, 43, 226, 0.5); transform: translateY(-2px); }
        
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 15px; margin: 20px 0; }
        
        .stat-card { background: linear-gradient(135deg, rgba(88, 166, 255, 0.1) 0%, rgba(138, 43, 226, 0.08) 100%); border: 2px solid transparent; border-image: linear-gradient(135deg, #58a6ff, #8a2be2, #ff6b9d) 1; border-radius: 10px; padding: 20px 15px; position: relative; overflow: hidden; transition: all 0.3s ease; }
        .stat-card::before { content: ''; position: absolute; top: -50%; left: -50%; width: 200%; height: 200%; background: radial-gradient(circle, rgba(88, 166, 255, 0.08) 0%, transparent 70%); animation: glow 4s ease-in-out infinite; }
        @keyframes glow { 0%, 100% { transform: translate(0, 0) scale(1); } 50% { transform: translate(8px, 8px) scale(1.05); } }
        .stat-card:hover { box-shadow: 0 0 25px rgba(88, 166, 255, 0.25), 0 0 15px rgba(138, 43, 226, 0.15); transform: translateY(-3px); }
        
        .stat-label { font-size: clamp(0.7em, 1.5vw, 0.8em); color: #ffb3d9; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 8px; font-weight: 600; }
        .stat-value { font-size: clamp(1.8em, 4vw, 2.3em); font-weight: 700; background: linear-gradient(135deg, #58a6ff, #79c0ff, #8a2be2); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; margin: 8px 0; }
        .stat-unit { font-size: clamp(0.75em, 1.8vw, 0.9em); color: #9db4c7; margin-top: 5px; }
        
        .section { background: linear-gradient(135deg, rgba(26, 31, 58, 0.8) 0%, rgba(42, 8, 69, 0.5) 100%); border: 2px solid rgba(88, 166, 255, 0.25); border-radius: 10px; padding: 20px 15px; margin: 20px 0; box-shadow: 0 0 20px rgba(138, 43, 226, 0.1); }
        .section h2 { font-size: clamp(1.3em, 4vw, 1.8em); color: #ffffff; margin-bottom: 15px; position: relative; padding-bottom: 10px; font-weight: 700; letter-spacing: 0.5px; }
        .section h2::after { content: ''; position: absolute; bottom: 0; left: 0; width: 60px; height: 3px; background: linear-gradient(90deg, #58a6ff, #8a2be2, transparent); }
        
        .chart-container { background: rgba(10, 14, 39, 0.9); border: 2px solid rgba(88, 166, 255, 0.2); border-radius: 8px; padding: 15px; height: clamp(250px, 50vw, 320px); margin: 15px 0; position: relative; }
        
        .wallets-table { width: 100%; border-collapse: collapse; }
        .wallets-table thead { background: linear-gradient(90deg, rgba(88, 166, 255, 0.25), rgba(138, 43, 226, 0.2)); }
        .wallets-table th { padding: 12px 10px; text-align: left; color: #ffffff; font-weight: 700; border-bottom: 2px solid rgba(138, 43, 226, 0.4); font-size: clamp(0.75em, 2vw, 0.95em); }
        .wallets-table td { padding: 10px; border-bottom: 1px solid rgba(88, 166, 255, 0.15); color: #e0e8ff; font-size: clamp(0.75em, 1.8vw, 0.9em); }
        .wallets-table tr:hover { background: rgba(88, 166, 255, 0.08); }
        .wallet-addr { color: #ff6b9d; font-weight: 700; word-break: break-all; }
        .wallet-balance { color: #7ee787; font-weight: 700; }
        .wallet-pulse { animation: pulse-val 0.6s ease; }
        @keyframes pulse-val { 0% { color: #ffa500; transform: scale(1.1); } 100% { color: #7ee787; transform: scale(1); } }
        
        .wallet-actions { display: flex; gap: 6px; }
        .wallet-btn { padding: 4px 8px; font-size: 0.75em; border: 1px solid rgba(138, 43, 226, 0.4); background: transparent; color: #a8d8ff; border-radius: 4px; cursor: pointer; transition: all 0.2s; }
        .wallet-btn:hover { background: rgba(138, 43, 226, 0.2); border-color: #8a2be2; }
        .wallet-btn-delete { color: #ff8787; border-color: rgba(255, 135, 135, 0.3); }
        .wallet-btn-delete:hover { background: rgba(255, 135, 135, 0.15); border-color: #ff8787; }
        
        .terminal { background: rgba(0, 0, 0, 0.95); border: 2px solid rgba(88, 166, 255, 0.25); border-radius: 8px; padding: 12px; height: clamp(250px, 40vh, 350px); overflow-y: auto; font-size: clamp(0.75em, 1.5vw, 0.9em); }
        .log-entry { margin: 4px 0; padding: 4px; border-left: 3px solid transparent; padding-left: 8px; animation: logSlide 0.4s ease; word-break: break-word; }
        @keyframes logSlide { from { opacity: 0; transform: translateX(-15px); } to { opacity: 1; transform: translateX(0); } }
        .log-info { border-left-color: #58a6ff; color: #a8d8ff; }
        .log-success { border-left-color: #7ee787; color: #7ee787; }
        .log-error { border-left-color: #ff6b6b; color: #ff8787; }
        .log-time { color: #8b949e; margin-right: 8px; }
        
        .footer { text-align: center; margin-top: 30px; padding-top: 15px; border-top: 2px solid rgba(88, 166, 255, 0.2); color: #8b949e; font-size: clamp(0.8em, 1.5vw, 0.9em); }
        
        .modal { display: none; position: fixed; z-index: 1000; left: 0; top: 0; width: 100%; height: 100%; background-color: rgba(0, 0, 0, 0.7); animation: fadeIn 0.3s ease; }
        @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
        .modal-content { background: linear-gradient(135deg, rgba(26, 31, 58, 0.98) 0%, rgba(42, 8, 69, 0.95) 100%); margin: 5% auto; padding: 25px; border: 2px solid rgba(138, 43, 226, 0.5); border-radius: 12px; width: clamp(280px, 90%, 450px); box-shadow: 0 0 50px rgba(138, 43, 226, 0.3); animation: slideIn 0.4s ease; }
        @keyframes slideIn { from { transform: translateY(-50px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
        .modal-header { font-size: 1.5em; color: #ffffff; font-weight: 700; margin-bottom: 20px; }
        .modal-close { color: #ff6b9d; float: right; font-size: 1.5em; cursor: pointer; transition: all 0.2s; }
        .modal-close:hover { color: #ff8fb3; transform: scale(1.2); }
        .form-group { margin-bottom: 15px; }
        .form-label { color: #a8d8ff; font-weight: 600; margin-bottom: 5px; display: block; }
        .form-input { width: 100%; padding: 10px; background: rgba(10, 14, 39, 0.8); border: 2px solid rgba(88, 166, 255, 0.3); border-radius: 6px; color: #ffffff; font-family: 'Courier Prime', monospace; transition: all 0.2s; }
        .form-input:focus { outline: none; border-color: #8a2be2; box-shadow: 0 0 10px rgba(138, 43, 226, 0.4); }
        .modal-buttons { display: flex; gap: 10px; margin-top: 20px; }
        .modal-btn { flex: 1; padding: 10px; border: none; border-radius: 6px; cursor: pointer; font-weight: 600; font-family: 'Courier Prime', monospace; transition: all 0.3s; }
        .modal-btn-submit { background: linear-gradient(135deg, #8a2be2, #5c2d91); color: white; }
        .modal-btn-submit:hover { box-shadow: 0 0 15px rgba(138, 43, 226, 0.5); }
        .modal-btn-cancel { background: rgba(88, 166, 255, 0.15); color: #a8d8ff; border: 1px solid rgba(88, 166, 255, 0.3); }
        .modal-btn-cancel:hover { background: rgba(88, 166, 255, 0.25); }
        
        ::-webkit-scrollbar { width: 8px; }
        ::-webkit-scrollbar-track { background: rgba(88, 166, 255, 0.1); }
        ::-webkit-scrollbar-thumb { background: rgba(138, 43, 226, 0.5); border-radius: 4px; }
        ::-webkit-scrollbar-thumb:hover { background: rgba(138, 43, 226, 0.8); }
        
        @media (max-width: 768px) {
            body { padding: 10px; }
            header { padding: 15px 10px; margin-bottom: 20px; }
            h1 { letter-spacing: 1px; }
            .header-controls { gap: 8px; }
            .btn { padding: 6px 12px; font-size: 0.85em; }
            .stats-grid { grid-template-columns: repeat(2, 1fr); gap: 12px; margin: 15px 0; }
            .stat-card { padding: 15px 12px; }
            .stat-label { margin-bottom: 6px; }
            .stat-value { margin: 6px 0; }
            .section { padding: 15px 12px; margin: 15px 0; }
            .section h2 { margin-bottom: 12px; }
            .chart-container { padding: 12px; height: 220px; }
            .wallets-table th, .wallets-table td { padding: 8px 6px; }
            .terminal { padding: 10px; height: 220px; font-size: 0.8em; }
            .footer { margin-top: 20px; padding-top: 12px; }
            .modal-content { width: 85%; margin: 30% auto; }
        }
        
        @media (max-width: 480px) {
            body { padding: 8px; }
            header { padding: 12px 8px; margin-bottom: 15px; }
            h1 { margin: 5px 0; }
            .subtitle { margin-top: 5px; }
            .header-controls { gap: 6px; margin-top: 10px; }
            .btn { padding: 6px 12px; font-size: 0.8em; }
            .stats-grid { grid-template-columns: 1fr; gap: 10px; margin: 10px 0; }
            .stat-card { padding: 12px 10px; }
            .stat-label { font-size: 0.65em; }
            .stat-value { font-size: 1.5em; margin: 4px 0; }
            .stat-unit { font-size: 0.7em; }
            .section { padding: 12px 10px; margin: 12px 0; }
            .section h2 { font-size: 1.2em; margin-bottom: 10px; }
            .chart-container { padding: 10px; height: 180px; }
            .wallets-table th, .wallets-table td { padding: 6px 4px; font-size: 0.7em; }
            .wallet-addr { word-break: break-all; }
            .terminal { padding: 8px; height: 180px; font-size: 0.7em; }
            .footer { margin-top: 15px; padding-top: 10px; font-size: 0.8em; }
            .modal-content { width: 90%; margin: 40% auto; padding: 20px; }
            .modal-header { font-size: 1.2em; }
            .modal-buttons { flex-direction: column; }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>⚡ S-UAH MINER</h1>
            <div class="subtitle">Real-time Blockchain Mining Dashboard</div>
            <div class="header-controls">
                <button class="btn" onclick="openAddWalletModal()">➕ Додати гаманець</button>
            </div>
        </header>
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-label">📊 Хешрейт</div>
                <div class="stat-value" id="hashrate">0.00</div>
                <div class="stat-unit">MH/s</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">🔗 Блоків сесії</div>
                <div class="stat-value" id="blocks">0</div>
                <div class="stat-unit">шт</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">💰 Баланс</div>
                <div class="stat-value" id="balance">0</div>
                <div class="stat-unit">S-UAH</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">⏱️ Статус</div>
                <div class="stat-value" id="status" style="font-size: 1.5em;">🟢 ONLINE</div>
                <div class="stat-unit" id="uptime">0s</div>
            </div>
        </div>
        <div class="section">
            <h2>📈 Історія хешрейту (60 сек)</h2>
            <div class="chart-container">
                <canvas id="hashChart"></canvas>
            </div>
        </div>
        <div class="section">
            <h2>💼 Гаманці</h2>
            <table class="wallets-table">
                <thead>
                    <tr><th>Назва</th><th>Адреса</th><th>Блоків</th><th>Баланс</th><th>Дія</th></tr>
                </thead>
                <tbody id="wallet-list"></tbody>
            </table>
        </div>
        <div class="section">
            <h2>📜 Лог подій</h2>
            <div class="terminal" id="terminal"></div>
        </div>
        <div class="footer">🚀 S-Hryvnia Miner | Real-time</div>
    </div>

    <div id="addWalletModal" class="modal">
        <div class="modal-content">
            <span class="modal-close" onclick="closeAddWalletModal()">&times;</span>
            <div class="modal-header">➕ Додати новий гаманець</div>
            <form onsubmit="submitAddWallet(event)">
                <div class="form-group">
                    <label class="form-label">Адреса гаманця</label>
                    <input type="text" id="walletAddress" class="form-input" placeholder="Введіть адресу гаманця..." required minlength="20">
                </div>
                <div class="form-group">
                    <label class="form-label">Пароль</label>
                    <input type="password" id="walletPassword" class="form-input" placeholder="Введіть пароль..." required minlength="6">
                </div>
                <div class="modal-buttons">
                    <button type="submit" class="modal-btn modal-btn-submit">Додати</button>
                    <button type="button" class="modal-btn modal-btn-cancel" onclick="closeAddWalletModal()">Скасувати</button>
                </div>
            </form>
        </div>
    </div>

    <div id="editWalletModal" class="modal">
        <div class="modal-content">
            <span class="modal-close" onclick="closeEditWalletModal()">&times;</span>
            <div class="modal-header">✏️ Назва гаманця</div>
            <form onsubmit="submitEditWallet(event)">
                <div class="form-group">
                    <label class="form-label">Назва</label>
                    <input type="text" id="editWalletName" class="form-input" placeholder="Назва гаманця..." required>
                </div>
                <div class="modal-buttons">
                    <button type="submit" class="modal-btn modal-btn-submit">Зберегти</button>
                    <button type="button" class="modal-btn modal-btn-cancel" onclick="closeEditWalletModal()">Скасувати</button>
                </div>
            </form>
        </div>
    </div>

    <div id="deleteWalletModal" class="modal">
        <div class="modal-content">
            <span class="modal-close" onclick="closeDeleteWalletModal()">&times;</span>
            <div class="modal-header">⚠️ Видалити гаманець</div>
            <p style="color: #ffb3d9; margin-bottom: 20px;">Ви впевнені? Ця дія не може бути скасована.</p>
            <div class="form-group">
                <label class="form-label">Пароль для підтвердження</label>
                <input type="password" id="deleteWalletPassword" class="form-input" placeholder="Введіть пароль..." required>
            </div>
            <div class="modal-buttons">
                <button onclick="submitDeleteWallet()" class="modal-btn modal-btn-submit" style="background: linear-gradient(135deg, #ff4444, #cc0000);">Видалити</button>
                <button type="button" class="modal-btn modal-btn-cancel" onclick="closeDeleteWalletModal()">Скасувати</button>
            </div>
        </div>
    </div>

    <script>
        let chart = null;
        const es = new EventSource("/events");
        const walletList = document.getElementById('wallet-list');
        const terminal = document.getElementById('terminal');

        function initChart() {
            const ctx = document.getElementById('hashChart').getContext('2d');
            if (chart) chart.destroy();
            chart = new Chart(ctx, {
                type: 'line',
                data: { 
                    labels: Array(60).fill(''), 
                    datasets: [{
                        label: 'MH/s',
                        data: Array(60).fill(0),
                        borderColor: '#58a6ff',
                        backgroundColor: 'rgba(88, 166, 255, 0.1)',
                        borderWidth: 3,
                        fill: true,
                        tension: 0.4,
                        pointRadius: 0,
                        pointHoverRadius: 6
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: { legend: { display: false } },
                    scales: {
                        y: {
                            beginAtZero: true,
                            grid: { color: 'rgba(88, 166, 255, 0.1)' },
                            ticks: { color: '#79c0ff' }
                        },
                        x: { grid: { display: false }, ticks: { display: false } }
                    }
                }
            });
        }

        es.onmessage = (e) => {
            try {
                const data = JSON.parse(e.data);
                document.getElementById('hashrate').innerText = data.hashrate.toFixed(2);
                document.getElementById('blocks').innerText = data.total_blocks;
                document.getElementById('balance').innerText = data.total_balance;
                document.getElementById('uptime').innerText = 'Uptime: ' + data.uptime;
                document.getElementById('status').innerText = '🟢 ONLINE';

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
                        row.innerHTML = '<td id="name-' + w.address + '" class="wallet-name" style="color: #79c0ff; font-weight: 600; cursor: pointer;" onclick="openEditWalletModal(\'' + w.address + '\')">' + (w.name || 'Безімено') + '</td><td class="wallet-addr">' + (w.address || '?').substring(0, 12) + '...</td><td id="sess-' + w.address + '">0</td><td id="bal-' + w.address + '" class="wallet-balance">0</td><td><div class="wallet-actions"><button class="wallet-btn" onclick="openEditWalletModal(\'' + w.address + '\')">✏️</button><button class="wallet-btn wallet-btn-delete" onclick="openDeleteWalletModal(\'' + w.address + '\')">🗑️</button></div></td>';
                        walletList.appendChild(row);
                    }
                    const s = document.getElementById('sess-' + w.address);
                    const b = document.getElementById('bal-' + w.address);
                    const n = document.getElementById('name-' + w.address);
                    if (n && n.innerText != (w.name || 'Безімено')) {
                        n.innerText = w.name || 'Безімено';
                    }
                    if (s && s.innerText != w.session_mined) {
                        s.innerText = w.session_mined;
                        s.classList.add('wallet-pulse');
                        setTimeout(() => s.classList.remove('wallet-pulse'), 600);
                    }
                    if (b && b.innerText != w.server_balance) {
                        b.innerText = w.server_balance;
                        b.classList.add('wallet-pulse');
                        setTimeout(() => b.classList.remove('wallet-pulse'), 600);
                    }
                });

                if (data.new_logs && data.new_logs.length > 0) {
                    data.new_logs.forEach(log => {
                        const div = document.createElement('div');
                        div.className = 'log-entry log-' + log.type;
                        div.innerHTML = '<span class="log-time">[' + log.time + ']</span>' + log.message;
                        terminal.prepend(div);
                    });
                    while(terminal.children.length > 200) terminal.removeChild(terminal.lastChild);
                }
            } catch (e) { }
        };

        es.onerror = () => {
            document.getElementById('status').innerText = '🔴 OFFLINE';
        };

        initChart();

        function openAddWalletModal() {
            document.getElementById('addWalletModal').style.display = 'block';
            document.getElementById('walletAddress').value = '';
            document.getElementById('walletPassword').value = '';
            document.getElementById('walletAddress').focus();
        }

        function closeAddWalletModal() {
            document.getElementById('addWalletModal').style.display = 'none';
        }

        window.onclick = (event) => {
            const modal = document.getElementById('addWalletModal');
            if (event.target === modal) {
                modal.style.display = 'none';
            }
        };

        async function submitAddWallet(event) {
            event.preventDefault();
            const address = document.getElementById('walletAddress').value.trim();
            const password = document.getElementById('walletPassword').value;
            
            if (!address || address.length < 20) {
                alert('❌ Адреса повинна бути щонайменше 20 символів');
                return;
            }
            if (!password || password.length < 6) {
                alert('❌ Пароль повинен бути щонайменше 6 символів');
                return;
            }

            try {
                const response = await fetch('/api/add-wallet', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ address, password })
                });
                const result = await response.json();
                if (result.success) {
                    alert('✅ ' + result.message);
                    closeAddWalletModal();
                } else {
                    alert('❌ ' + result.message);
                }
            } catch (e) {
                alert('❌ Помилка при додаванні гаманця: ' + e.message);
            }
        }

        let currentEditWalletAddress = '';
        let currentDeleteWalletAddress = '';

        function openEditWalletModal(address) {
            currentEditWalletAddress = address;
            const nameCell = document.getElementById('name-' + address);
            const currentName = nameCell ? nameCell.innerText : '';
            document.getElementById('editWalletName').value = currentName === 'Безімено' ? '' : currentName;
            document.getElementById('editWalletModal').style.display = 'block';
            document.getElementById('editWalletName').focus();
        }

        function closeEditWalletModal() {
            document.getElementById('editWalletModal').style.display = 'none';
            currentEditWalletAddress = '';
        }

        async function submitEditWallet(event) {
            event.preventDefault();
            const name = document.getElementById('editWalletName').value.trim();
            
            try {
                const response = await fetch('/api/rename-wallet', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ address: currentEditWalletAddress, name: name || 'Безімено' })
                });
                const result = await response.json();
                if (result.success) {
                    document.getElementById('name-' + currentEditWalletAddress).innerText = name || 'Безімено';
                    closeEditWalletModal();
                } else {
                    alert('❌ ' + result.message);
                }
            } catch (e) {
                alert('❌ Помилка: ' + e.message);
            }
        }

        function openDeleteWalletModal(address) {
            currentDeleteWalletAddress = address;
            document.getElementById('deleteWalletPassword').value = '';
            document.getElementById('deleteWalletModal').style.display = 'block';
            document.getElementById('deleteWalletPassword').focus();
        }

        function closeDeleteWalletModal() {
            document.getElementById('deleteWalletModal').style.display = 'none';
            currentDeleteWalletAddress = '';
        }

        async function submitDeleteWallet() {
            const password = document.getElementById('deleteWalletPassword').value;
            if (!password) {
                alert('❌ Введіть пароль');
                return;
            }

            try {
                const response = await fetch('/api/delete-wallet', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ address: currentDeleteWalletAddress, password })
                });
                const result = await response.json();
                if (result.success) {
                    document.getElementById('w-' + currentDeleteWalletAddress).remove();
                    closeDeleteWalletModal();
                    alert('✅ Гаманець видалено');
                } else {
                    alert('❌ ' + result.message);
                }
            } catch (e) {
                alert('❌ Помилка: ' + e.message);
            }
        }

        window.addEventListener('click', (event) => {
            const editModal = document.getElementById('editWalletModal');
            const deleteModal = document.getElementById('deleteWalletModal');
            const addModal = document.getElementById('addWalletModal');
            if (event.target === editModal) closeEditWalletModal();
            if (event.target === deleteModal) closeDeleteWalletModal();
            if (event.target === addModal) closeAddWalletModal();
        });
    </script>
</body>
</html>
`
