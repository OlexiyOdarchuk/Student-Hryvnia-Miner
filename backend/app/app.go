package app

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"shminer/backend/config"
	"shminer/backend/internal/logger"
	"shminer/backend/types"
	"sync"
	"time"
)

type Miner interface {
	MineBlock(prevHash string, wallet string) bool
	CompileDifficultyBits(bits int)
}

type WebDashboardServer interface {
	BroadcastUpdate()
	StartWebServer()
}

type Wallet interface {
	GetWallets() []string
	SyncStorage()
}

type NodeClient interface {
	GetChainLastHashCached() string
	SubmitBlock(prev, wallet string, nonce int, ts int64, hash string) bool
	GetBalance(addr string) float64
}

type Stats interface {
	StartSpeedMonitor(ctx context.Context)
	StartBalanceUpdater(ctx context.Context)
	UpdateSingleBalance(wallet string)
	GetDashboardData() types.DashboardData
	SessionMinedIncrement()
}

type App struct {
	startTime     time.Time
	walletDataMap map[string]*types.WalletStats
	stats         Stats
	webDashboard  WebDashboardServer
	config        *config.AppConfig
	dashboard     *types.DashboardData
	mu            sync.RWMutex
	ctx           context.Context
	nodeClient    NodeClient
	minerClient   Miner
	walletClient  Wallet
}

func (a *App) StartApp(logCallback func(entry types.LogEntry)) {
	uiLogger := slog.New(&logger.UIHandler{LogCallback: logCallback, Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})})
	slog.SetDefault(uiLogger)

}

func (a *App) startMiningLoop() {
	if a.config.Difficulty < 1 {
		a.config.Difficulty = 1
	}

	a.minerClient.CompileDifficultyBits(a.config.Difficulty)
	a.mu.Lock()
	a.startTime = time.Now()
	a.mu.Unlock()

	rand.New(rand.NewSource(time.Now().UnixNano()))

	go a.stats.StartSpeedMonitor(a.ctx)
	go a.stats.StartBalanceUpdater(a.ctx)
	go a.webDashboard.StartWebServer()

	slog.Info("🔨 Miner started...")

	for {
		select {
		case <-a.ctx.Done():
			slog.Info("🛑 Mining stopped")
			return
		default:
		}

		prevHash := a.nodeClient.GetChainLastHashCached()
		if prevHash == "" {
			slog.Error("⚠️ No connection to the server. Restarting miner...")
			time.Sleep(2 * time.Second)
			continue
		}

		ws := a.walletClient.GetWallets()
		if len(ws) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		currentWallet := ws[rand.Intn(len(ws))]

		a.mu.Lock()
		walletStats, exists := a.walletDataMap[currentWallet]
		isWorking := true
		if exists {
			isWorking = walletStats.Working
		}
		a.mu.Unlock()

		if !isWorking {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		success := a.minerClient.MineBlock(prevHash, currentWallet)

		if success {
			a.mu.Lock()
			a.stats.SessionMinedIncrement()
			if ws, ok := a.walletDataMap[currentWallet]; ok {
				ws.SessionMined++
				ws.TotalMined++
			}

			a.walletClient.SyncStorage()
			SaveStorage(sessionPassword, CurrentStorage)
			a.mu.Unlock()

			go func() {
				a.stats.UpdateSingleBalance(currentWallet)
				a.webDashboard.BroadcastUpdate()
			}()

			a.webDashboard.BroadcastUpdate()
		}

		time.Sleep(config.MinerSleepInterval)
	}
}
