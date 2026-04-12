<script lang="ts">
    import { onMount } from "svelte";
    import Chart from "chart.js/auto";
    import { stats, logs, connected, notifications } from "../stores";
    import { SetGlobalMining } from "../../wailsjs/go/main/App";

    let chartCanvas: HTMLCanvasElement;
    let chart: Chart;

    function openFocus() {
        document.dispatchEvent(new CustomEvent("toggle-focus"));
    }

    let isAnyWorking = false;
    let isOnline = false;
    let statusText = "ОФЛАЙН";
    let statusClass = "offline";
    let activeWalletsCount = 0;

    $: if ($stats || $connected) {
        isOnline = $connected;
        const wallets = $stats.wallets || [];
        isAnyWorking = wallets.some((w) => w.working);
        activeWalletsCount = wallets.filter((w) => w.working).length;

        statusText = !isOnline
            ? "ОФЛАЙН"
            : isAnyWorking
              ? "ОНЛАЙН"
              : "ПРИЗУПИНЕНО";
        statusClass = !isOnline ? "offline" : isAnyWorking ? "" : "paused";
    }

    let lastSessionBlocks = 0;
    let pulseClass = "";
    let pulseTimer: any = null;

    $: if ($stats && $stats.session_blocks > lastSessionBlocks) {
        if (pulseTimer) clearTimeout(pulseTimer);
        pulseClass = "pulse-active";
        pulseTimer = setTimeout(() => (pulseClass = ""), 1000);
        lastSessionBlocks = $stats.session_blocks;
    }
    $: if ($stats && lastSessionBlocks === 0 && $stats.session_blocks > 0) {
        lastSessionBlocks = $stats.session_blocks;
    }

    async function toggleGlobal() {
        const newState = !isAnyWorking;
        await SetGlobalMining(newState);

        if (newState) {
            notifications.success("Всі воркери запущено!");
        } else {
            notifications.info(
                "Всі воркери зупинено. Оновлення балансів активне.",
            );
        }
    }

    let chartInterval: any;

    onMount(() => {
        setTimeout(() => {
            if (!chartCanvas) return;
            try {
                const ctx = chartCanvas.getContext("2d");
                const gradient = ctx.createLinearGradient(0, 0, 0, 400);
                gradient.addColorStop(0, "rgba(129, 140, 248, 0.5)");
                gradient.addColorStop(1, "rgba(129, 140, 248, 0.0)");

                chart = new Chart(ctx, {
                    type: "line",
                    data: {
                        labels: Array(60).fill(""),
                        datasets: [
                            {
                                label: "Хешрейт (MH/s)",
                                data: Array(60).fill(0),
                                borderColor: "#818cf8",
                                backgroundColor: gradient,
                                borderWidth: 2,
                                tension: 0.4,
                                fill: true,
                                pointRadius: 0,
                            },
                        ],
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        interaction: { intersect: false, mode: "index" },
                        scales: {
                            x: { display: false },
                            y: {
                                beginAtZero: true,
                                grid: { color: "rgba(255, 255, 255, 0.05)" },
                                ticks: { color: "#64748b" },
                            },
                        },
                        plugins: {
                            legend: { display: false },
                            tooltip: {
                                backgroundColor: "rgba(15, 23, 42, 0.9)",
                                titleColor: "#fff",
                                bodyColor: "#cbd5e1",
                                borderColor: "rgba(255, 255, 255, 0.1)",
                                borderWidth: 1,
                                padding: 10,
                                displayColors: false,
                            },
                        },
                    },
                });

                chartInterval = setInterval(() => {
                    let currentStats;
                    stats.subscribe(v => currentStats = v)();
                    if (chart && currentStats) {
                        try {
                            const ds = chart.data.datasets[0].data;
                            if (ds.length > 60) ds.shift();
                            ds.push(currentStats.hashrate);
                            chart.update("none");
                        } catch (e) {
                            console.error("Chart update failed", e);
                        }
                    }
                }, 1000);
            } catch (e) {
                console.error("Chart init failed", e);
            }
        }, 100);

        return () => {
            if (chartInterval) clearInterval(chartInterval);
        };
    });
</script>

<div class="content-wrapper">
    <div class="dash-header" style="flex-shrink: 0;">
        <div class="page-title">Головна</div>

        <div style="display: flex; gap: 15px; align-items: center;">
            <div class="dash-controls">
                <button
                    class="btn-icon"
                    on:click={toggleGlobal}
                    title={isAnyWorking ? "Зупинити всі" : "Запустити всі"}
                    style="width: auto; padding: 0 20px; font-weight: 600; gap: 8px;"
                >
                    <i
                        class="fas {isAnyWorking ? 'fa-stop' : 'fa-play'}"
                        style="color: {isAnyWorking
                            ? 'var(--danger)'
                            : 'var(--success)'}"
                    ></i>
                    <span>{isAnyWorking ? "СТОП" : "СТАРТ"}</span>
                </button>
            </div>

            <div class="live-badge {statusClass}">
                <div class="dot"></div>
                {statusText}
            </div>
        </div>
    </div>

    <div class="grid-3" style="flex-shrink: 0;">
        <div class="glass-card stat-card">
            <i class="fas fa-bolt stat-icon-bg"></i>
            <div class="stat-label">Хешрейт мережі</div>
            <div class="stat-value">{$stats.hashrate.toFixed(2)}</div>
            <div class="stat-sub">MH/s Середня швидкість</div>
        </div>
        <div
            class="glass-card stat-card"
            style="border-top-color: var(--success);"
        >
            <i class="fas fa-wallet stat-icon-bg"></i>
            <div class="stat-label">Загальний баланс</div>
            <div class="stat-value" style="color: var(--success);">
                {$stats.total_balance.toFixed(2)}
            </div>
            <div class="stat-sub">S-UAH Накопичено</div>
        </div>
        <div
            class="glass-card stat-card"
            style="border-top-color: var(--neon-cyan);"
        >
            <i class="fas fa-clock stat-icon-bg"></i>
            <div class="stat-label">Аптайм</div>
            <div class="stat-value uptime-counter">{$stats.uptime}</div>
            <div class="stat-sub">Час роботи ноди</div>
        </div>
    </div>

    <div class="chart-grid">
        <div
            class="glass-card chart-section no-hover"
            style="grid-column: 1; display: flex; flex-direction: column; height: 100%;"
        >
            <div
                style="display:flex; justify-content:space-between; margin-bottom:10px; flex-shrink: 0;"
            >
                <div class="stat-label">Історія Хешрейту</div>
                <button
                    class="btn-icon"
                    on:click={openFocus}
                    title="Відкрити режим фокусу"
                >
                    <i class="fas fa-expand"></i>
                </button>
            </div>
            <div
                style="flex:1; position:relative; width: 100%; min-height: 0; overflow: hidden;"
            >
                <canvas
                    bind:this={chartCanvas}
                    style="width: 100%; height: 100%; display: block;"
                ></canvas>
            </div>
        </div>

        <div
            class="glass-card recent-logs no-hover"
            style="grid-column: 2; display: flex; flex-direction: column; height: 100%; max-height: none;"
        >
            <div
                style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; flex-shrink: 0;"
            >
                <div class="stat-label">Остання активність</div>
            </div>
            <div
                id="quick-stats-container"
                style="flex: 1; overflow-y: auto; min-height: 0;"
            >
                {#each $logs.slice().reverse().slice(0, 50) as log}
                    <div class="log-row-mini {log.type}">
                        <span
                            style="opacity:0.5; margin-right: 8px; font-size: 0.8em;"
                            >{log.time}</span
                        >
                        <span>{log.message}</span>
                    </div>
                {/each}
                {#if $logs.length === 0}
                    <div
                        class="wallet-select-placeholder"
                        style="padding: 20px; text-align: center; color: #64748b; border: none; background: transparent;"
                    >
                        <i
                            class="fas fa-info-circle"
                            style="font-size: 2rem; margin-bottom: 10px; opacity: 0.5;"
                        ></i>
                        <div>Очікування подій...</div>
                    </div>
                {/if}
            </div>
        </div>
    </div>

    <div
        class="glass-card no-hover"
        style="flex-shrink: 0; padding: 20px; display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; text-align: center;"
        id="fun-stats-section"
    >
        <div
            style="padding: 10px; background: rgba(129, 140, 248, 0.05); border-radius: 12px; border: 1px solid rgba(129, 140, 248, 0.1);"
        >
            <div
                style="font-size: 1.5rem; font-weight: 800; font-family: var(--font-mono); color: var(--primary); margin-bottom: 4px;"
            >
                {$stats.hashrate.toFixed(2)}
            </div>
            <div
                style="font-size: 0.75rem; color: #94a3b8; text-transform: uppercase;"
            >
                Швидкість
            </div>
        </div>
        <div
            style="padding: 10px; background: rgba(16, 185, 129, 0.05); border-radius: 12px; border: 1px solid rgba(16, 185, 129, 0.1);"
        >
            <div
                style="font-size: 1.5rem; font-weight: 800; font-family: var(--font-mono); color: var(--success); margin-bottom: 4px;"
            >
                {$stats.total_balance.toFixed(2)}
            </div>
            <div
                style="font-size: 0.75rem; color: #94a3b8; text-transform: uppercase;"
            >
                Зароблено
            </div>
        </div>
        <div
            class={pulseClass}
            style="padding: 10px; background: rgba(251, 191, 36, 0.05); border-radius: 12px; border: 1px solid rgba(251, 191, 36, 0.1); transition: box-shadow 0.2s, border-color 0.2s;"
        >
            <div
                style="font-size: 1.5rem; font-weight: 800; font-family: var(--font-mono); color: var(--warning); margin-bottom: 4px;"
            >
                {$stats.session_blocks}
            </div>
            <div
                style="font-size: 0.75rem; color: #94a3b8; text-transform: uppercase;"
            >
                Блоки
            </div>
        </div>
        <div
            style="padding: 10px; background: rgba(6, 182, 212, 0.05); border-radius: 12px; border: 1px solid rgba(6, 182, 212, 0.1);"
        >
            <div
                style="font-size: 1.5rem; font-weight: 800; font-family: var(--font-mono); color: var(--neon-cyan); margin-bottom: 4px;"
            >
                {activeWalletsCount}
            </div>
            <div
                style="font-size: 0.75rem; color: #94a3b8; text-transform: uppercase;"
            >
                Активні
            </div>
        </div>
    </div>
</div>