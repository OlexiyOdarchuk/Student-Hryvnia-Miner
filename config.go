package main

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
)

var (
	wallets      []string
	walletsMutex sync.RWMutex
)

var Config struct {
	BaseURL       string
	ServerPort    string
	Difficulty    int
	HTTPTimeout   time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
	BalanceFreq   time.Duration
	AdminPassword string
}

func LoadConfig() {
	Config.BaseURL = getEnvOrDefault("BASE_URL", DefaultBaseURL)
	Config.ServerPort = getEnvOrDefault("SERVER_PORT", DefaultServerPort)
	Config.Difficulty = getEnvAsInt("DIFFICULTY", DefaultDifficulty)
	Config.HTTPTimeout = getEnvAsTimeDuration("HTTP_TIMEOUT", int(DefaultHTTPTimeout.Seconds()))
	Config.MaxRetries = getEnvAsInt("MAX_RETRIES", DefaultMaxRetries)
	Config.RetryDelay = time.Duration(getEnvAsInt("RETRY_DELAY_MS", int(DefaultRetryDelay.Milliseconds()))) * time.Millisecond
	Config.BalanceFreq = time.Duration(getEnvAsInt("BALANCE_UPDATE_FREQ", int(DefaultBalanceUpdateFreq.Seconds()))) * time.Second
	Config.AdminPassword = getEnvOrDefault("ADMIN_PASSWORD", DefaultAdminPassword)
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if num, err := strconv.Atoi(val); err == nil {
			return num
		}
	}
	return defaultValue
}

func getEnvAsTimeDuration(key string, defaultSeconds int) time.Duration {
	if val := os.Getenv(key); val != "" {
		if sec, err := strconv.Atoi(val); err == nil {
			return time.Duration(sec) * time.Second
		}
	}
	return time.Duration(defaultSeconds) * time.Second
}

func watchEnvFile() {
	envPath, _ := filepath.Abs(".env")

	watcher, _ := fsnotify.NewWatcher()
	defer watcher.Close()

	watcher.Add(envPath)
	pushLog("👀 Watching "+envPath, "info")

	for {
		select {
		case e := <-watcher.Events:
			if e.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Chmod) != 0 {
				time.Sleep(EnvWatcherDebounce)
				clearEnvVariables()
				if err := godotenv.Load(envPath); err != nil {
					pushLog("❌ env reload error", "error")
					continue
				}
				LoadConfig()
				compileDifficultyBits(Config.Difficulty)
				pushLog("♻️ Конфіг перезавантажено", "info")
			}
		case err := <-watcher.Errors:
			pushLog("⚠️ watcher error: "+err.Error(), "error")
		}
	}
}

func clearEnvVariables() {
	envVars := []string{
		"WALLETS",
		"BASE_URL",
		"SERVER_PORT",
		"DIFFICULTY",
		"HTTP_TIMEOUT",
		"MAX_RETRIES",
		"RETRY_DELAY_MS",
		"BALANCE_CHECK_INTERVAL",
		"BALANCE_UPDATE_FREQ",
		"ADMIN_PASSWORD",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
