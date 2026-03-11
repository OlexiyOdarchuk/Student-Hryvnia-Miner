package main

import (
	"context"
	"encoding/json"
	"fmt"
	stdRuntime "runtime"
	"shminer/backend"
	"shminer/backend/app"
	"shminer/backend/app/config"
	"shminer/backend/internal/storage"
	"shminer/backend/internal/wallets"
	"shminer/backend/types"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	cancelMining  context.CancelFunc
	miningStarted bool
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	logCallback := func(entry types.LogEntry) {
		runtime.EventsEmit(ctx, "log", entry)
	}
	app.StartApp(ctx, logCallback)
}

func (a *App) startMining() {
	if a.miningStarted {
		return
	}

	miningCtx, cancel := context.WithCancel(context.Background())
	a.cancelMining = cancel
	a.miningStarted = true

	go backend.StartMiningLoop(miningCtx)

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-miningCtx.Done():
				return
			case <-ticker.C:
				data := stats.GetDashboardData()
				runtime.EventsEmit(a.ctx, "stats", data)
			}
		}
	}()
}

func (a *App) shutdown(ctx context.Context) {
	if a.cancelMining != nil {
		a.cancelMining()
	}
}

// --- Auth Methods ---

func (a *App) IsStorageInitialized() bool {
	return storage.StorageExists()
}

func (a *App) InitStorage(password string) string {
	err := storage.InitStorage(password)
	if err != nil {
		return err.Error()
	}
	a.startMining()
	return ""
}

func (a *App) UnlockStorage(password string) string {
	err := storage.LoadStorage(password)
	if err != nil {
		return "Невірний пароль"
	}
	a.startMining()
	return ""
}

// --- Exposed methods ---

func (a *App) GetDashboardData() types.DashboardData {
	return stats.GetDashboardData()
}

func (a *App) GetWallets() []string {
	return wallets.GetWallets()
}

func (a *App) AddWallet(name, address, privateKey string) string {
	err := wallets.AddWalletSafe(name, address, privateKey)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) ImportWalletJSON(jsonContent string) string {
	var w wallets.WalletExport
	err := json.Unmarshal([]byte(jsonContent), &w)
	if err != nil {
		return "Невірний формат JSON"
	}
	if w.Name == "" || w.Pub == "" || w.Priv == "" {
		return "JSON повинен містити поля name, pub, priv"
	}

	err = wallets.AddWalletSafe(w.Name, w.Pub, w.Priv)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) GetWalletJSONSecure(address, password string) (string, error) {
	if password != storage.GetSessionPassword() {
		return "", fmt.Errorf("Невірний пароль")
	}

	return wallets.ExportWalletJSON(address)
}

func (a *App) DeleteWallet(address, password string) string {
	if password != storage.GetSessionPassword() {
		return "Невірний пароль"
	}
	err := wallets.DeleteWallet(address)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) RenameWallet(address, newName string) string {
	err := wallets.RenameWallet(address, newName)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) UpdateWalletKey(address, key, password string) string {
	if password != storage.GetSessionPassword() {
		return "Невірний пароль"
	}
	err := wallets.UpdateWalletKey(address, key)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) GetWalletKey(address, password string) (string, error) {
	if password != storage.GetSessionPassword() {
		return "", fmt.Errorf("Невірний пароль")
	}

	privKey := wallets.GetPrivateKey(address)
	if privKey == "" {
		return "", fmt.Errorf("Приватний ключ не знайдено")
	}

	return privKey, nil
}

func (a *App) ToggleWallet(address string) bool {
	return wallets.ToggleWalletMining(address)
}

func (a *App) SetGlobalMining(state bool) {
	wallets.SetAllMining(state)
}

// Settings
func (a *App) GetConfig() types.AppConfig {
	return storage.CurrentStorage.Config
}

func (a *App) UpdateConfig(newConf types.AppConfig, password string) string {
	if password != storage.GetSessionPassword() {
		return "Невірний пароль"
	}
	err := config.UpdateConfig(password, newConf)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) ChangePassword(oldPass, newPass string) string {
	err := storage.ChangePassword(oldPass, newPass)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"cpu_cores": stdRuntime.NumCPU(),
	}
}
