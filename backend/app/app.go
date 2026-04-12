package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

type App struct {
	storageDriver *storage.Storage
	walletService *wallets.Wallets
	statsService  *stats.Stats
	storageBuffer chan struct{}
	shutdownWG    sync.WaitGroup
	minerClient   *miner.Miner
	nodeClient    nodeclient.NodeClient
	webDashboard  *web_dashboard.Server

	walletMu      *sync.RWMutex
	walletDataMap map[string]*types.WalletStats

	parentCtx context.Context
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
		storageBuffer: make(chan struct{}, 90),
		shutdownWG:    sync.WaitGroup{},
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
	a.parentCtx = ctx
	miningCtx, cancel := context.WithCancel(ctx)
	a.miningCtx = miningCtx
	a.cancel = cancel

	a.shutdownWG.Add(1)
	go a.statsService.StartSpeedMonitor(miningCtx, &a.shutdownWG)
	a.shutdownWG.Add(1)
	go a.statsService.StartBalanceUpdater(miningCtx, &a.shutdownWG)
	a.shutdownWG.Add(1)
	go a.statsService.StartTelemetryReporter(miningCtx, &a.shutdownWG, config.TelemetryProxyURL, config.Config.MinerID)
	a.shutdownWG.Add(1)
	go a.autoSaver(miningCtx)
	a.shutdownWG.Add(1)
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
	a.shutdownWG.Wait()
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
	defer a.shutdownWG.Done()
	cfg := &config.Config
	if cfg.Difficulty < 1 {
		cfg.Difficulty = 1
	}

	a.minerClient.CompileDifficultyBits(cfg.Difficulty)
	a.statsService.ResetSessionMined()
	a.statsService.SetStartTime(time.Now())
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	slog.Info("🔨 Miner started...")

	var cachedPrevHash string
	failCh := make(chan struct{}, 10)

	type submitPayload struct {
		prev   string
		wallet string
		nonce  int
		ts     int64
		hash   string
	}
	submitCh := make(chan submitPayload, 200)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		creditedCount := 0

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if creditedCount > 0 {
					slog.Info("💰 " + strconv.Itoa(creditedCount) + " blocks credited")
					creditedCount = 0
					a.webDashboard.BroadcastUpdate()
				}
			case sub := <-submitCh:
				if a.nodeClient.SubmitBlock(sub.prev, sub.wallet, sub.nonce, sub.ts, sub.hash) {
					creditedCount++
					a.walletMu.Lock()
					a.statsService.SessionMinedIncrement()
					if ws, ok := a.walletDataMap[sub.wallet]; ok {
						ws.SessionMined++
						ws.TotalMined++
					}
					a.walletMu.Unlock()

					select {
					case a.storageBuffer <- struct{}{}:
					default:
					}
				} else {
					slog.Error("❌ Server rejected block. Resetting...")
					select {
					case failCh <- struct{}{}:
					default:
					}
				clearLoop:
					for {
						select {
						case <-submitCh:
						default:
							break clearLoop
						}
					}
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			slog.Info("🛑 Mining stopped")
			return
		case <-failCh:
			cachedPrevHash = ""
		default:
		}

		if cachedPrevHash == "" {
			cachedPrevHash = a.nodeClient.GetChainLastHashCached()
		}

		if cachedPrevHash == "" {
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

		a.walletMu.RLock()
		walletStats, exists := a.walletDataMap[currentWallet]
		isWorking := true
		if exists {
			isWorking = walletStats.Working
		}
		a.walletMu.RUnlock()

		if !isWorking {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		blockCtx, cancelBlock := context.WithCancel(ctx)
		go func(hashToTrack string) {
			checkFreq := time.Duration(cfg.BlockCheckFreqMs) * time.Millisecond
			if checkFreq < 1000*time.Millisecond {
				checkFreq = 5000 * time.Millisecond
			}
			ticker := time.NewTicker(checkFreq)
			fastTicker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()
			defer fastTicker.Stop()

			for {
				select {
				case <-blockCtx.Done():
					return
				case <-fastTicker.C:
					a.walletMu.RLock()
					ws, ok := a.walletDataMap[currentWallet]
					stillWorking := ok && ws.Working
					a.walletMu.RUnlock()
					if !stillWorking {
						cancelBlock()
						return
					}
				case <-ticker.C:
					latestHash := a.nodeClient.GetChainLastHashCached()
					if latestHash != "" && latestHash != hashToTrack {
						slog.Info("⚠️ Block updated by network. Restarting...")
						select {
						case failCh <- struct{}{}:
						default:
						}
						cancelBlock()
						return
					}
				}
			}
		}(cachedPrevHash)

		newHash, nonce, ts := a.minerClient.MineBlock(blockCtx, cachedPrevHash, currentWallet)
		cancelBlock()

		if newHash == "" {
			continue
		}

		payload := submitPayload{
			prev:   cachedPrevHash,
			wallet: currentWallet,
			nonce:  nonce,
			ts:     ts,
			hash:   newHash,
		}

		select {
		case submitCh <- payload:
		case <-ctx.Done():
			return
		}

		cachedPrevHash = newHash
	}
}

func (a *App) IsStorageInitialized() bool {
	return a.storageDriver.CheckExists()
}

func (a *App) InitStorage(password string) error {
	if err := a.storageDriver.InitStorage(password); err != nil {
		return err
	}
	a.syncWalletSnapshot()
	return nil
}

func (a *App) UnlockStorage(password string) error {
	if err := a.storageDriver.LoadStorage(password); err != nil {
		return err
	}
	a.syncWalletSnapshot()
	return nil
}

func (a *App) syncWalletSnapshot() {
	snapshot := a.storageDriver.GetStorage()
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
		return "", errors.New("private key not found")
	}
	return key, nil
}

func (a *App) ImportWalletJSON(jsonContent string) error {
	var payload wallets.WalletExport
	if err := json.Unmarshal([]byte(jsonContent), &payload); err != nil {
		return errors.New("incorrect format JSON")
	}
	if payload.Name == "" || payload.Pub == "" || payload.Priv == "" {
		return errors.New("JSON must contain the fields name, pub, priv")
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
	if err := a.storageDriver.PersistConfig(password); err != nil {
		return err
	}

	a.mu.Lock()
	wasMining := a.miningCtx != nil
	pCtx := a.parentCtx
	a.mu.Unlock()

	if wasMining {
		a.StopMining()
		a.StartMining(pCtx)
	} else {
		a.mu.Lock()
		a.applyConfig()
		a.mu.Unlock()
	}

	return nil
}

func (a *App) ChangePassword(oldPass, newPass string) error {
	return a.storageDriver.ChangePassword(oldPass, newPass)
}

func (a *App) verifyPassword(password string) error {
	if password != a.storageDriver.GetSessionPassword() {
		return errors.New("invalid password")
	}
	return nil
}

func (a *App) SendTransaction(from, to, password string, amount int) error {
	privKeyStr, err := a.GetWalletKey(from, password)
	if err != nil {
		return err
	}

	privBytes, err := hex.DecodeString(privKeyStr)
	if err != nil {
		return errors.New("private key decryption error")
	}

	privKey := secp256k1.PrivKeyFromBytes(privBytes)

	txObj := types.TxPayload{
		From:   from,
		To:     to,
		Amount: amount,
		Fee:    0,
	}

	jsonBytes, err := json.Marshal(txObj)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(jsonBytes)

	signature := ecdsa.Sign(privKey, hash[:])
	txObj.Signature = hex.EncodeToString(signature.Serialize())

	return a.nodeClient.SendTransaction(txObj)
}

func (a *App) SendMessageToDeveloper(contact, message string) {
	a.statsService.SendMessageToDeveloper(config.TelemetryProxyURL, config.Config.MinerID, contact, message)
}

func (a *App) autoSaver(ctx context.Context) {
	defer a.shutdownWG.Done()
	var buffer []struct{}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(buffer) > 0 {
				a.walletService.SyncStorage()
				if err := a.storageDriver.SaveStorage(a.storageDriver.GetSessionPassword(), a.storageDriver.GetStorage()); err != nil {
					slog.Error("❌ Save storage on shutdown failed", "error", err, "count", len(buffer))
				} else {
					slog.Info("✅ Save storage on shutdown", "count", len(buffer))
				}
			}
			return
		case <-a.storageBuffer:
			buffer = append(buffer, struct{}{})
			if len(buffer) >= 60 {
				a.walletService.SyncStorage()
				err := a.storageDriver.SaveStorage(a.storageDriver.GetSessionPassword(), a.storageDriver.GetStorage())
				if err != nil {
					slog.Error("❌ Save Storage error by count", "error", err)
					continue
				}
				buffer = nil
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				a.walletService.SyncStorage()
				err := a.storageDriver.SaveStorage(a.storageDriver.GetSessionPassword(), a.storageDriver.GetStorage())
				if err != nil {
					slog.Error("❌ Save Storage error by timer", "error", err)
				}
				buffer = nil
			}
		}
	}
}
