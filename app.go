package main

import (
	"context"
	"fmt"
	"shminer/backend"
	stdRuntime "runtime"
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

	backend.LogCallback = func(entry backend.LogEntry) {
		runtime.EventsEmit(ctx, "log", entry)
	}
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
				data := backend.GetDashboardData()
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
	return backend.StorageExists()
}

func (a *App) InitStorage(password string) string {
	err := backend.InitStorage(password)
	if err != nil {
		return err.Error()
	}
	a.startMining()
	return ""
}

func (a *App) UnlockStorage(password string) string {
	err := backend.LoadStorage(password)
	if err != nil {
		return "Невірний пароль"
	}
	a.startMining()
	return ""
}

// --- Exposed methods ---

func (a *App) GetDashboardData() backend.DashboardData {
	return backend.GetDashboardData()
}

func (a *App) GetWallets() []string {
	return backend.GetWallets()
}

func (a *App) AddWallet(name, address, privateKey string) string {
	err := backend.AddWalletSafe(name, address, privateKey)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) DeleteWallet(address, password string) string {
	if password != backend.GetSessionPassword() {
		return "Невірний пароль"
	}
	err := backend.DeleteWallet(address)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) RenameWallet(address, newName string) {
	backend.RenameWallet(address, newName)
}

func (a *App) UpdateWalletKey(address, key, password string) string {
	if password != backend.GetSessionPassword() {
		return "Невірний пароль"
	}
	err := backend.UpdateWalletKey(address, key)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) GetWalletKey(address, password string) (string, error) {
	if password != backend.GetSessionPassword() {
		return "", fmt.Errorf("Невірний пароль")
	}

	privKey := backend.GetPrivateKey(address)
	if privKey == "" {
		return "", fmt.Errorf("Приватний ключ не знайдено")
	}

	return privKey, nil
}

func (a *App) ToggleWallet(address string) bool {
	return backend.ToggleWalletMining(address)
}

func (a *App) SetGlobalMining(state bool) {
	backend.SetAllMining(state)
}

func (a *App) SendTransaction(from, to string, amount float64, password string) string {
	if password != backend.GetSessionPassword() {
		return "Невірний пароль"
	}

	privKey := backend.GetPrivateKey(from)
	_, err := backend.SendTransaction(from, privKey, to, amount)
	if err != nil {
		return err.Error()
	}
	return ""
}

// Settings
func (a *App) GetConfig() backend.AppConfig {
	return backend.CurrentStorage.Config
}

func (a *App) UpdateConfig(newConf backend.AppConfig, password string) string {
	if password != backend.GetSessionPassword() {
		return "Невірний пароль"
	}
	err := backend.UpdateConfig(password, newConf)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) ChangePassword(oldPass, newPass string) string {
	err := backend.ChangePassword(oldPass, newPass)
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
