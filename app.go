package main

import (
	"context"
	stdRuntime "runtime"
	"shminer/backend/app"
	"shminer/backend/config"
	"shminer/backend/types"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	cancelMining  context.CancelFunc
	miningStarted bool
	backendApp    app.Backend
}

func NewApp() *App {
	return &App{
		backendApp: app.Init(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logCallback := func(entry types.LogEntry) {
		runtime.EventsEmit(ctx, "log", entry)
	}
	a.backendApp.StartApp(logCallback)
}

func (a *App) startMining() {
	if a.miningStarted {
		return
	}

	miningCtx, cancel := context.WithCancel(context.Background())
	a.cancelMining = cancel
	a.miningStarted = true

	a.backendApp.StartMining(miningCtx)

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-miningCtx.Done():
				return
			case <-ticker.C:
				data := a.backendApp.GetDashboardData()
				runtime.EventsEmit(a.ctx, "stats", data)
			}
		}
	}()
}

func (a *App) shutdown(ctx context.Context) {
	if a.cancelMining != nil {
		a.cancelMining()
		a.backendApp.StopMining()
	}
}

// --- Auth Methods ---

func (a *App) IsStorageInitialized() bool {
	return a.backendApp.IsStorageInitialized()
}

func (a *App) InitStorage(password string) string {
	if err := a.backendApp.InitStorage(password); err != nil {
		return err.Error()
	}
	a.startMining()
	return ""
}

func (a *App) UnlockStorage(password string) string {
	if err := a.backendApp.UnlockStorage(password); err != nil {
		return "Невірний пароль"
	}
	a.startMining()
	return ""
}

// --- Exposed methods ---

func (a *App) GetDashboardData() types.DashboardData {
	return a.backendApp.GetDashboardData()
}

func (a *App) GetWallets() []string {
	return a.backendApp.GetWallets()
}

func (a *App) AddWallet(name, address, privateKey string) string {
	if err := a.backendApp.AddWallet(name, address, privateKey); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) ImportWalletJSON(jsonContent string) string {
	if err := a.backendApp.ImportWalletJSON(jsonContent); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) GetWalletJSONSecure(address, password string) (string, error) {
	return a.backendApp.GetWalletJSONSecure(address, password)
}

func (a *App) DeleteWallet(address, password string) string {
	if err := a.backendApp.DeleteWallet(address, password); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) RenameWallet(address, newName string) string {
	if err := a.backendApp.RenameWallet(address, newName); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) UpdateWalletKey(address, key, password string) string {
	if err := a.backendApp.UpdateWalletKey(address, key, password); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) GetWalletKey(address, password string) (string, error) {
	return a.backendApp.GetWalletKey(address, password)
}

func (a *App) ToggleWallet(address string) bool {
	return a.backendApp.ToggleWallet(address)
}

func (a *App) SetGlobalMining(state bool) {
	_ = a.backendApp.SetGlobalMining(state)
}

// Settings

func (a *App) GetConfig() config.AppConfig {
	return a.backendApp.GetConfig()
}

func (a *App) UpdateConfig(newConf config.AppConfig, password string) string {
	if err := a.backendApp.UpdateConfig(newConf, password); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) ChangePassword(oldPass, newPass string) string {
	if err := a.backendApp.ChangePassword(oldPass, newPass); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"cpu_cores": stdRuntime.NumCPU(),
	}
}
