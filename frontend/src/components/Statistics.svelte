<script>
    import { onMount } from 'svelte';
    import { stats } from '../stores';
    import Chart from 'chart.js/auto';

    let balanceCanvas;
    let blocksCanvas;
    let balanceChart;
    let blocksChart;

    // Reactive update for charts
    $: if ($stats && balanceChart && blocksChart) {
        updateCharts();
    }

    function updateCharts() {
        const wallets = $stats.wallets || [];
        const labels = wallets.map(w => w.name || w.address.substring(0, 8));
        const balances = wallets.map(w => w.server_balance);
        const blocks = wallets.map(w => w.total_mined);

        // Update Balance Chart (Pie)
        balanceChart.data.labels = labels;
        balanceChart.data.datasets[0].data = balances;
        balanceChart.update();

        // Update Blocks Chart (Bar)
        blocksChart.data.labels = labels;
        blocksChart.data.datasets[0].data = blocks;
        blocksChart.update();
    }

    onMount(() => {
        // Balance Chart
        balanceChart = new Chart(balanceCanvas, {
            type: 'doughnut',
            data: {
                labels: [],
                datasets: [{
                    data: [],
                    backgroundColor: [
                        '#818cf8', '#34d399', '#f472b6', '#fbbf24', '#06b6d4'
                    ],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { position: 'right', labels: { color: '#94a3b8' } }
                }
            }
        });

        // Blocks Chart
        blocksChart = new Chart(blocksCanvas, {
            type: 'bar',
            data: {
                labels: [],
                datasets: [{
                    label: 'Загальні блоки',
                    data: [],
                    backgroundColor: '#818cf8',
                    borderRadius: 8
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { 
                        beginAtZero: true,
                        grid: { color: 'rgba(255,255,255,0.05)' },
                        ticks: { color: '#64748b', stepSize: 1 }
                    },
                    x: {
                        grid: { display: false },
                        ticks: { color: '#64748b' }
                    }
                },
                plugins: {
                    legend: { display: false }
                }
            }
        });

        // Initial update
        if ($stats.wallets) updateCharts();
    });
</script>

<div class="content-wrapper">
    
    <div class="dash-header" style="flex-shrink: 0;">
        <div class="page-title">Детальна статистика</div>
    </div>

    <!-- Top Cards (4 Grid) -->
    <div class="grid-4" style="flex-shrink: 0;">
        <div class="glass-card stat-card">
            <i class="fas fa-server stat-icon-bg"></i>
            <div class="stat-label">Активні гаманці</div>
            <div class="stat-value" style="color: var(--primary);">{($stats.wallets || []).filter(w => w.working).length}</div>
            <div class="stat-sub">Зараз майнять</div>
        </div>
        
        <div class="glass-card stat-card" style="border-top-color: var(--warning);">
            <i class="fas fa-cube stat-icon-bg"></i>
            <div class="stat-label">Сесійні блоки</div>
            <div class="stat-value" style="color: var(--warning);">{$stats.session_blocks}</div>
            <div class="stat-sub">За цей запуск</div>
        </div>

        <div class="glass-card stat-card" style="border-top-color: var(--accent);">
            <i class="fas fa-layer-group stat-icon-bg"></i>
            <div class="stat-label">Всього блоків</div>
            <div class="stat-value" style="color: var(--accent);">{$stats.lifetime_blocks}</div>
            <div class="stat-sub">За весь час</div>
        </div>
        
        <div class="glass-card stat-card" style="border-top-color: var(--neon-cyan);">
            <i class="fas fa-clock stat-icon-bg"></i>
            <div class="stat-label">Час сесії</div>
            <div class="stat-value uptime-counter" style="color: var(--neon-cyan); font-size: 2rem;">{$stats.uptime}</div>
            <div class="stat-sub">Час роботи</div>
        </div>
    </div>

    <!-- Charts - This section GROWS (flex: 1) -->
    <div class="chart-grid">
        <div class="glass-card" style="padding: 24px; display: flex; flex-direction: column; height: 100%;">
            <div class="stat-label" style="margin-bottom: 20px; flex-shrink: 0;">Розподіл балансу</div>
            <div style="flex: 1; position: relative; width: 100%; min-height: 0; overflow: hidden;">
                <canvas bind:this={balanceCanvas} style="width: 100%; height: 100%; display: block;"></canvas>
            </div>
        </div>
        
        <div class="glass-card" style="padding: 24px; display: flex; flex-direction: column; height: 100%;">
            <div class="stat-label" style="margin-bottom: 20px; flex-shrink: 0;">Блоки по гаманцях (Всього)</div>
            <div style="flex: 1; position: relative; width: 100%; min-height: 0; overflow: hidden;">
                <canvas bind:this={blocksCanvas} style="width: 100%; height: 100%; display: block;"></canvas>
            </div>
        </div>
    </div>

    <!-- Bottom Table -->
    <div class="glass-card" style="padding: 24px; flex-shrink: 0; max-height: 300px; overflow-y: auto;">
        <div class="stat-label" style="margin-bottom: 20px;">Продуктивність гаманців</div>
        <div class="table-wrap">
            <table class="lux-table">
                <thead>
                    <tr>
                        <th>Гаманець</th>
                        <th>Статус</th>
                        <th>С. Блоки</th>
                        <th>З. Блоки</th>
                        <th>Баланс</th>
                    </tr>
                </thead>
                <tbody>
                    {#each ($stats.wallets || []) as wallet}
                    <tr>
                        <td>{wallet.name}</td>
                        <td>
                            {#if wallet.working}
                                <span class="status-badge status-active">АКТИВНИЙ</span>
                            {:else}
                                <span class="status-badge status-paused">ПАУЗА</span>
                            {/if}
                        </td>
                        <td>{wallet.session_mined}</td>
                        <td>{wallet.total_mined}</td>
                        <td>{wallet.server_balance.toFixed(2)} S-UAH</td>
                    </tr>
                    {/each}
                    {#if ($stats.wallets || []).length === 0}
                        <tr><td colspan="5" style="text-align: center; color: #64748b; padding: 20px;">Немає гаманців</td></tr>
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>

<style>
    .grid-4 {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
        gap: 24px;
    }
</style>