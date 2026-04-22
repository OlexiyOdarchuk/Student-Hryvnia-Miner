<script lang="ts">
    import { onMount } from "svelte";
    import Chart from "chart.js/auto";
    import { stats, logs, connected, notifications } from "../stores";
    import { SetMining } from "../../wailsjs/go/main/App";

    let chartCanvas: HTMLCanvasElement;
    let chart: Chart;

    function openFocus() {
        document.dispatchEvent(new CustomEvent("toggle-focus"));
    }

    let isMining = false;
    let isOnline = false;
    let statusText = "ОФЛАЙН";
    let statusClass = "offline";
    let activeWalletsCount = 0;
    let totalWalletsCount = 0;

    $: if ($stats || $connected) {
        isOnline = $connected;
        const wallets = $stats.wallets || [];
        totalWalletsCount = wallets.length;
        isMining = $stats.is_mining === true;
        activeWalletsCount = wallets.filter((w) => w.working).length;

        statusText = !isOnline
            ? "ОФЛАЙН"
            : isMining
              ? "ОНЛАЙН"
              : "ПРИЗУПИНЕНО";
        statusClass = !isOnline ? "offline" : isMining ? "" : "paused";
    }

    let lastSessionBlocks = 0;
    let pulseClass = "";
    let pulseTimer: any = null;

    type LogFilter = "all" | "credit" | "error";
    let activeFilter: LogFilter = "all";

    function matchesFilter(log: any): boolean {
        if (activeFilter === "all") return true;
        if (activeFilter === "error") return log.type === "ERROR";
        if (activeFilter === "credit") {
            return (
                typeof log.message === "string" &&
                log.message.includes("credited")
            );
        }
        return true;
    }

    $: displayLogs = $logs
        .slice()
        .reverse()
        .filter(matchesFilter)
        .slice(0, 80);

    $: errorCount = $logs.filter((l) => l.type === "ERROR").length;
    $: creditEntries = $logs.filter(
        (l) =>
            typeof l.message === "string" && l.message.includes("credited"),
    );

    $: queueLen = $stats.submit_queue_len || 0;
    $: creditedPerMin = $stats.blocks_per_min || 0;
    $: foundPerMin = $stats.found_per_min || 0;

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
        const newState = !isMining;
        await SetMining(newState);

        if (newState) {
            notifications.success("Майнінг запущено!");
        } else {
            notifications.info(
                "Майнінг призупинено. Оновлення балансів активне.",
            );
        }
    }

    function logIcon(log: any): string {
        const msg = typeof log.message === "string" ? log.message : "";
        if (log.type === "ERROR") return "fa-circle-exclamation";
        if (msg.includes("credited")) return "fa-coins";
        if (msg.includes("Miner started")) return "fa-play";
        if (msg.includes("Mining stopped")) return "fa-stop";
        if (msg.includes("Block updated")) return "fa-arrows-rotate";
        if (msg.includes("No connection")) return "fa-plug-circle-xmark";
        return "fa-circle-info";
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
                    const currentStats = $stats;
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
                    title={isMining ? "Зупинити майнінг" : "Запустити майнінг"}
                    style="width: auto; padding: 0 20px; font-weight: 600; gap: 8px;"
                >
                    <i
                        class="fas {isMining ? 'fa-stop' : 'fa-play'}"
                        style="color: {isMining
                            ? 'var(--danger)'
                            : 'var(--success)'}"
                    ></i>
                    <span>{isMining ? "СТОП" : "СТАРТ"}</span>
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
            <div class="log-header">
                <div class="stat-label">Журнал подій</div>
                <div class="log-filters">
                    <button
                        class="log-filter-btn"
                        class:active={activeFilter === "all"}
                        on:click={() => (activeFilter = "all")}
                        title="Усі події"
                    >
                        Всі
                        <span class="filter-count">{$logs.length}</span>
                    </button>
                    <button
                        class="log-filter-btn credit"
                        class:active={activeFilter === "credit"}
                        on:click={() => (activeFilter = "credit")}
                        title="Тільки зараховані блоки"
                    >
                        <i class="fas fa-coins"></i>
                        <span class="filter-count">{creditEntries.length}</span>
                    </button>
                    <button
                        class="log-filter-btn error"
                        class:active={activeFilter === "error"}
                        on:click={() => (activeFilter = "error")}
                        title="Тільки помилки"
                    >
                        <i class="fas fa-circle-exclamation"></i>
                        <span class="filter-count">{errorCount}</span>
                    </button>
                </div>
            </div>
            <div class="log-body">
                {#each displayLogs as log (log.id || log.message + log.time)}
                    <div class="log-row-mini {log.type}">
                        <i class="fas {logIcon(log)} log-icon"></i>
                        <span class="log-time">{log.time}</span>
                        <span class="log-msg">{log.message}</span>
                    </div>
                {/each}
                {#if displayLogs.length === 0}
                    <div class="log-empty">
                        <i class="fas fa-satellite-dish"></i>
                        <div>
                            {#if $logs.length === 0}
                                Очікування подій...
                            {:else}
                                Немає подій за обраним фільтром
                            {/if}
                        </div>
                    </div>
                {/if}
            </div>
        </div>
    </div>

    <div
        class="metric-grid"
        style="flex-shrink: 0;"
        id="fun-stats-section"
    >
        <div class="glass-card no-hover metric-cell {pulseClass}" style="--c: var(--warning);">
            <i class="fas fa-coins metric-icon"></i>
            <div class="metric-val">{$stats.session_blocks}</div>
            <div class="metric-label">Зараховано</div>
            <div class="metric-sub-line">{creditedPerMin.toFixed(2)} / хв</div>
        </div>
        <div class="glass-card no-hover metric-cell" style="--c: var(--accent);">
            <i class="fas fa-hammer metric-icon"></i>
            <div class="metric-val">{$stats.session_found || 0}</div>
            <div class="metric-label">Намайнено</div>
            <div class="metric-sub-line">{foundPerMin.toFixed(2)} / хв</div>
        </div>
        <div class="glass-card no-hover metric-cell" style="--c: var(--neon-cyan);">
            <i class="fas fa-layer-group metric-icon"></i>
            <div class="metric-val">{queueLen}</div>
            <div class="metric-label">У черзі</div>
            <div class="metric-sub-line">очікують відправки</div>
        </div>
        <div class="glass-card no-hover metric-cell" style="--c: var(--success);">
            <i class="fas fa-wallet metric-icon"></i>
            <div class="metric-val">
                {activeWalletsCount}<span class="metric-sub"
                    >/{totalWalletsCount}</span
                >
            </div>
            <div class="metric-label">Активні гаманці</div>
            <div class="metric-sub-line">у роботі зараз</div>
        </div>
    </div>
</div>

<style>
    .metric-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
        gap: 14px;
    }
    .metric-cell {
        position: relative;
        padding: 18px 18px 14px;
        background: linear-gradient(
            140deg,
            color-mix(in srgb, var(--c) 12%, transparent) 0%,
            color-mix(in srgb, var(--c) 3%, transparent) 65%,
            transparent 100%
        );
        border-top: 2px solid color-mix(in srgb, var(--c) 55%, transparent);
        overflow: hidden;
    }
    .metric-icon {
        position: absolute;
        right: 14px;
        top: 14px;
        font-size: 1.4rem;
        color: color-mix(in srgb, var(--c) 35%, transparent);
    }
    .metric-val {
        font-size: 2rem;
        font-weight: 800;
        font-family: var(--font-mono);
        color: var(--c);
        margin-bottom: 4px;
        line-height: 1;
    }
    .metric-sub {
        font-size: 1rem;
        color: #64748b;
        font-weight: 600;
    }
    .metric-label {
        font-size: 0.72rem;
        color: #cbd5e1;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        font-weight: 700;
    }
    .metric-sub-line {
        margin-top: 6px;
        font-size: 0.75rem;
        color: #64748b;
        font-family: var(--font-mono);
    }

    .log-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 15px;
        flex-shrink: 0;
        gap: 10px;
        flex-wrap: wrap;
    }
    .log-filters {
        display: flex;
        gap: 6px;
    }
    .log-filter-btn {
        background: rgba(255, 255, 255, 0.04);
        border: 1px solid rgba(255, 255, 255, 0.08);
        color: #cbd5e1;
        border-radius: 8px;
        padding: 5px 10px;
        font-size: 0.75rem;
        cursor: pointer;
        display: inline-flex;
        align-items: center;
        gap: 6px;
        transition: all 0.2s;
    }
    .log-filter-btn:hover {
        background: rgba(255, 255, 255, 0.08);
    }
    .log-filter-btn.active {
        background: rgba(129, 140, 248, 0.18);
        border-color: rgba(129, 140, 248, 0.4);
        color: white;
    }
    .log-filter-btn.credit.active {
        background: rgba(251, 191, 36, 0.18);
        border-color: rgba(251, 191, 36, 0.45);
    }
    .log-filter-btn.error.active {
        background: rgba(239, 68, 68, 0.18);
        border-color: rgba(239, 68, 68, 0.45);
    }
    .filter-count {
        background: rgba(0, 0, 0, 0.3);
        border-radius: 10px;
        padding: 1px 7px;
        font-size: 0.7rem;
        font-family: var(--font-mono);
    }

    .log-body {
        flex: 1;
        overflow-y: auto;
        min-height: 0;
    }
    .log-icon {
        width: 16px;
        text-align: center;
        margin-right: 8px;
        opacity: 0.8;
    }
    .log-time {
        opacity: 0.5;
        margin-right: 8px;
        font-size: 0.8em;
        font-family: var(--font-mono);
    }
    .log-row-mini.ERROR .log-icon {
        color: var(--danger);
    }

    .log-empty {
        padding: 30px 20px;
        text-align: center;
        color: #64748b;
    }
    .log-empty i {
        font-size: 2rem;
        margin-bottom: 10px;
        opacity: 0.5;
        display: block;
    }

    @keyframes pulse {
        0% {
            box-shadow: 0 0 0 0 color-mix(in srgb, var(--c) 45%, transparent);
        }
        100% {
            box-shadow: 0 0 0 12px transparent;
        }
    }
    .pulse-active {
        animation: pulse 1s ease-out;
        border-color: var(--c) !important;
    }
</style>
