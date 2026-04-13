package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	stdRuntime "runtime"
	"shminer/backend/app"
	"shminer/backend/config"
	"shminer/backend/types"
	"time"

	_ "embed"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed wails.json
var wailsConfig []byte

type App struct {
	ctx           context.Context
	cancelMining  context.CancelFunc
	backendApp    app.Backend
	miningStarted bool
}

type UpdateResult struct {
	Version string `json:"version"`
	Body    string `json:"body"`
	Found   bool   `json:"found"`
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

	go func() {
		msg, err := a.CheckAndApplyUpdate()
		if err != nil {
			slog.Error("Update error", "err", err)
		} else {
			slog.Info(msg)
		}
	}()
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
		ticker := time.NewTicker(1000 * time.Millisecond)
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

func (a *App) TryAutoLogin() bool {
	success, err := a.backendApp.TryAutoLogin()

	if err == nil && success {
		a.startMining()
		return true
	}

	if err != nil && err.Error() != "Невірний пароль" && err.Error() != "not_found" {
		slog.Error("Помилка автологіну", "err", err)
	}

	return false
}

func (a *App) GenerateKeyPair() (map[string]string, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	privBytes := privKey.Serialize()
	pubBytes := privKey.PubKey().SerializeUncompressed()

	return map[string]string{
		"public":  hex.EncodeToString(pubBytes),
		"private": hex.EncodeToString(privBytes),
	}, nil
}

func (a *App) SendTransaction(from, to, password string, amount int) string {
	err := a.backendApp.SendTransaction(from, to, password, amount)
	if err != nil {
		return err.Error()
	}
	return ""
}

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

func (a *App) SendMessageToDeveloper(contact, message string) {
	a.backendApp.SendMessageToDeveloper(contact, message)
}

func (a *App) CheckAndApplyUpdate() (string, error) {
	var version struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(wailsConfig, &version); err != nil {
		return "", errors.New("error reading wails.json: " + err.Error())
	}

	currentVersion := version.Version
	if currentVersion[0] != 'v' {
		currentVersion = "v" + currentVersion
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{})
	if err != nil {
		return "", err
	}

	repo := "OlexiyOdarchuk/Student-Hryvnia-Miner"

	latest, found, err := updater.DetectLatest(context.Background(), selfupdate.ParseSlug(repo))
	if err != nil {
		return "", errors.New("update search error: " + err.Error())
	}

	if !found || latest.LessOrEqual(currentVersion) {
		slog.Debug("Version update", "found", found, "latest", latest)
		return "You have the latest version. No update is needed", nil
	}

	runtime.EventsEmit(a.ctx, "update_available", UpdateResult{
		Found:   true,
		Version: latest.Version(),
		Body:    latest.ReleaseNotes,
	})

	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		return "", errors.New("error during update: " + err.Error())
	}

	return "The update was successful. Please restart the app.", nil
}

func (a *App) GetConfigFilePath() string {
	return a.backendApp.GetConfigFilePath()
}

func (a *App) OpenConfigFolder() {
	filePath := a.GetConfigFilePath()
	dir := filepath.Dir(filePath)

	var err error
	switch stdRuntime.GOOS {
	case "windows":
		err = exec.Command("explorer", dir).Start()
	case "darwin":
		err = exec.Command("open", dir).Start()
	case "linux":
		err = exec.Command("xdg-open", dir).Start()
	}

	if err != nil {
		slog.Error("Не вдалося відкрити папку конфігу", "err", err)
	}
}
