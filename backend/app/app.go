package app

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"shminer/backend/config"
	"shminer/backend/internal/logger"
	"shminer/backend/internal/miner"
	"shminer/backend/internal/nodeclient"
	"shminer/backend/internal/stats"
	"shminer/backend/internal/storage"
	"shminer/backend/internal/wallets"
	"shminer/backend/internal/web_dashboard"
	"shminer/backend/types"
)

type App struct {
	storageDriver *storage.Storage
	walletService *wallets.Wallets
	statsService  *stats.Stats
	minerClient   *miner.Miner
	nodeClient    nodeclient.NodeClient
	webDashboard  *web_dashboard.Server

	walletMu      *sync.RWMutex
	walletDataMap map[string]*types.WalletStats

	miningCtx context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
}

type noopWebDashBoard struct{}

func (noopWebDashBoard) BroadcastUpdate() {}

func New() *App {
	walletMu := &sync.RWMutex{}
	walletDataMap := make(map[string]*types.WalletStats)
	storageDriver := storage.NewDriver()

	httpClient := &http.Client{
		Timeout: time.Duration(config.Config.HTTPTimeout) * time.Second,
	}

	node := nodeclient.NewApiClient(
		config.Config.BaseURL,
		httpClient,
		time.Duration(config.Config.RetryDelayMs)*time.Millisecond,
		time.Duration(config.ExponentialBackoffMaxMs)*time.Millisecond,
		int(config.Config.MaxRetries),
	)

	walletService := wallets.New(storageDriver, walletMu, walletDataMap)
	statsService := stats.InitStats(&types.Stats{}, walletDataMap, walletMu, node, noopWebDashBoard{}, walletService, time.Duration(config.Config.BalanceFreqS)*time.Second)
	dashboard := web_dashboard.NewServer(config.Config.ServerPort, statsService)
	statsService.SetWebDashboard(dashboard)
	minerClient := miner.InitMiner(statsService.HashCountPtr(), node, int(config.Config.Threads))

	app := &App{
		storageDriver: storageDriver,
		walletService: walletService,
		statsService:  statsService,
		minerClient:   minerClient,
		nodeClient:    node,
		webDashboard:  dashboard,
		walletMu:      walletMu,
		walletDataMap: walletDataMap,
	}

	go dashboard.StartWebServer()
	return app
}

func (a *App) StartApp(logCallback func(types.LogEntry)) {
	uiLogger := slog.New(&logger.UIHandler{
		LogCallback: logCallback,
		Handler:     slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	})
	slog.SetDefault(uiLogger)
}

func (a *App) StartMining(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.miningCtx != nil && a.cancel != nil {
		return
	}

	a.applyConfig()

	if ctx == nil {
		ctx = context.Background()
	}
	miningCtx, cancel := context.WithCancel(ctx)
	a.miningCtx = miningCtx
	a.cancel = cancel

	go a.statsService.StartSpeedMonitor(miningCtx)
	go a.statsService.StartBalanceUpdater(miningCtx)
	go a.runMiningLoop(miningCtx)
}

func (a *App) StopMining() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
		a.miningCtx = nil
	}
}

func (a *App) applyConfig() {
	httpClient := &http.Client{
		Timeout: time.Duration(config.Config.HTTPTimeout) * time.Second,
	}

	node := nodeclient.NewApiClient(
		config.Config.BaseURL,
		httpClient,
		time.Duration(config.Config.RetryDelayMs)*time.Millisecond,
		time.Duration(config.ExponentialBackoffMaxMs)*time.Millisecond,
		int(config.Config.MaxRetries),
	)
	a.nodeClient = node
	a.statsService.SetNodeClient(node)
	a.statsService.SetBalanceFreq(time.Duration(config.Config.BalanceFreqS) * time.Second)
	a.minerClient = miner.InitMiner(a.statsService.HashCountPtr(), node, int(config.Config.Threads))
}

func (a *App) runMiningLoop(ctx context.Context) {
	cfg := &config.Config
	if cfg.Difficulty < 1 {
		cfg.Difficulty = 1
	}

	a.minerClient.CompileDifficultyBits(cfg.Difficulty)
	a.statsService.ResetSessionMined()
	a.statsService.SetStartTime(time.Now())
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	slog.Info("🔨 Miner started...")

	for {
		select {
		case <-ctx.Done():
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

		ws := a.walletService.GetWallets()
		if len(ws) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		currentWallet := ws[rnd.Intn(len(ws))]

		a.walletMu.Lock()
		walletStats, exists := a.walletDataMap[currentWallet]
		isWorking := true
		if exists {
			isWorking = walletStats.Working
		}
		a.walletMu.Unlock()

		if !isWorking {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		success := a.minerClient.MineBlock(prevHash, currentWallet)

		if success {
			a.walletMu.Lock()
			a.statsService.SessionMinedIncrement()
			if ws, ok := a.walletDataMap[currentWallet]; ok {
				ws.SessionMined++
				ws.TotalMined++
			}
			a.walletMu.Unlock()

			a.walletService.SyncStorage()
			storage.SaveStorage(storage.GetSessionPassword(), storage.GetStorage())

			go func(wallet string) {
				a.statsService.UpdateSingleBalance(wallet)
				a.webDashboard.BroadcastUpdate()
			}(currentWallet)

			a.webDashboard.BroadcastUpdate()
		}

		time.Sleep(config.MinerSleepInterval)
	}
}

func (a *App) IsStorageInitialized() bool {
	return storage.StorageExists()
}

func (a *App) InitStorage(password string) error {
	if err := storage.InitStorage(password); err != nil {
		return err
	}
	a.syncWalletSnapshot()
	return nil
}

func (a *App) UnlockStorage(password string) error {
	if err := storage.LoadStorage(password); err != nil {
		return err
	}
	a.syncWalletSnapshot()
	return nil
}

func (a *App) syncWalletSnapshot() {
	snapshot := storage.GetStorage()
	a.walletService.Load(snapshot)
}

func (a *App) GetDashboardData() types.DashboardData {
	return a.statsService.GetDashboardData()
}

func (a *App) GetWallets() []string {
	return a.walletService.GetWallets()
}

func (a *App) AddWallet(name, address, privateKey string) error {
	return a.walletService.AddWalletSafe(name, address, privateKey)
}

func (a *App) RenameWallet(address, newName string) error {
	return a.walletService.RenameWallet(address, newName)
}

func (a *App) DeleteWallet(address, password string) error {
	if err := a.verifyPassword(password); err != nil {
		return err
	}
	return a.walletService.DeleteWallet(address)
}

func (a *App) ToggleWallet(address string) bool {
	return a.walletService.ToggleWalletMining(address)
}

func (a *App) SetGlobalMining(state bool) error {
	return a.walletService.SetAllMining(state)
}

func (a *App) UpdateWalletKey(address, privateKey, password string) error {
	if err := a.verifyPassword(password); err != nil {
		return err
	}
	return a.walletService.UpdateWalletKey(address, privateKey)
}

func (a *App) GetWalletKey(address, password string) (string, error) {
	if err := a.verifyPassword(password); err != nil {
		return "", err
	}
	key := a.walletService.GetPrivateKey(address)
	if key == "" {
		return "", errors.New("Приватний ключ не знайдено")
	}
	return key, nil
}

func (a *App) ImportWalletJSON(jsonContent string) error {
	var payload wallets.WalletExport
	if err := json.Unmarshal([]byte(jsonContent), &payload); err != nil {
		return errors.New("Невірний формат JSON")
	}
	if payload.Name == "" || payload.Pub == "" || payload.Priv == "" {
		return errors.New("JSON повинен містити поля name, pub, priv")
	}
	return a.walletService.AddWalletSafe(payload.Name, payload.Pub, payload.Priv)
}

func (a *App) GetWalletJSONSecure(address, password string) (string, error) {
	if err := a.verifyPassword(password); err != nil {
		return "", err
	}
	return a.walletService.ExportWalletJSON(address)
}

func (a *App) GetConfig() config.AppConfig {
	return config.Config
}

func (a *App) UpdateConfig(newConf config.AppConfig, password string) error {
	if err := a.verifyPassword(password); err != nil {
		return err
	}
	config.Config.Update(newConf)
	if err := storage.PersistConfig(password); err != nil {
		return err
	}
	a.applyConfig()
	return nil
}

func (a *App) ChangePassword(oldPass, newPass string) error {
	return storage.ChangePassword(oldPass, newPass)
}

func (a *App) verifyPassword(password string) error {
	if password != storage.GetSessionPassword() {
		return errors.New("Невірний пароль")
	}
	return nil
}
