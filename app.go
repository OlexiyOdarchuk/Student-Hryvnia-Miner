package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	stdRuntime "runtime"
	"shminer/backend/app"
	"shminer/backend/config"
	"shminer/backend/types"
	"strings"
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

func (a *App) IsMining() bool {
	return a.backendApp.IsMining()
}

func (a *App) SetMining(state bool) {
	a.backendApp.SetMining(state)
}

func (a *App) SendTestTelegramMessage(token, chatID string) string {
	if err := a.backendApp.SendTestTelegramMessage(token, chatID); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) ResolveTelegramChatID(token, chatID string) (string, error) {
	return a.backendApp.ResolveTelegramChatID(token, chatID)
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

func (a *App) HasPassword() bool {
	return a.backendApp.HasPassword()
}

func (a *App) GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"cpu_cores": stdRuntime.NumCPU(),
	}
}

// GetLocalIP returns the first non-loopback IPv4 address of an interface that
// is up and not a point-to-point/virtual link. Returns "" if none is found.
func (a *App) GetLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		name := strings.ToLower(iface.Name)
		if strings.HasPrefix(name, "docker") ||
			strings.HasPrefix(name, "br-") ||
			strings.HasPrefix(name, "virbr") ||
			strings.HasPrefix(name, "veth") ||
			strings.HasPrefix(name, "tun") ||
			strings.HasPrefix(name, "tap") {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip4 := ip.To4()
			if ip4 == nil || !ip4.IsPrivate() {
				continue
			}
			return ip4.String()
		}
	}
	return ""
}

// GetDashboardURL returns a ready-to-share LAN URL for the dashboard, or ""
// if the machine has no routable LAN address.
func (a *App) GetDashboardURL() string {
	ip := a.GetLocalIP()
	if ip == "" {
		return ""
	}
	port := strings.TrimPrefix(config.Config.ServerPort, ":")
	if port == "" {
		port = "8080"
	}
	return "http://" + ip + ":" + port
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

	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Filters: []string{"^SHMiner-"},
	})
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

	if !canSelfReplace(exe) {
		msg := "Доступна нова версія " + latest.Version() +
			". Завантажте вручну: https://github.com/" + repo + "/releases/latest"
		a.emitUpdateStatus("info", msg)
		return msg, nil
	}

	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		return "", errors.New("error during update: " + err.Error())
	}

	msg := "Оновлено до " + latest.Version() + ". Перезапустіть програму."
	a.emitUpdateStatus("success", msg)
	return msg, nil
}

func (a *App) emitUpdateStatus(kind, msg string) {
	runtime.EventsEmit(a.ctx, "update_status", map[string]string{
		"type":    kind,
		"message": msg,
	})
}

func canSelfReplace(exe string) bool {
	if stdRuntime.GOOS != "linux" {
		return true
	}
	if os.Getenv("APPIMAGE") != "" {
		return false
	}
	for _, p := range []string{"/usr/bin/", "/usr/local/bin/", "/opt/"} {
		if strings.HasPrefix(exe, p) {
			return false
		}
	}
	return true
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
