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
	BaseURL      string
	ServerPort   string
	Difficulty   string
	HTTPTimeout  time.Duration
	MaxRetries   int
	RetryDelay   time.Duration
	BalanceCheck time.Duration
	BalanceFreq  time.Duration
}

func LoadConfig() {

	Config.BaseURL = os.Getenv("BASE_URL")
	if Config.BaseURL == "" {
		Config.BaseURL = "https://s-hryvnia-1.onrender.com"
	}

	Config.ServerPort = os.Getenv("SERVER_PORT")
	if Config.ServerPort == "" {
		Config.ServerPort = ":8090"
	}

	Config.Difficulty = os.Getenv("DIFFICULTY")
	if Config.Difficulty == "" {
		Config.Difficulty = "00000"
	}

	timeout := os.Getenv("HTTP_TIMEOUT")
	if timeout == "" {
		Config.HTTPTimeout = 5 * time.Second
	} else {
		if sec, err := strconv.Atoi(timeout); err == nil {
			Config.HTTPTimeout = time.Duration(sec) * time.Second
		}
	}

	retries := os.Getenv("MAX_RETRIES")
	if retries == "" {
		Config.MaxRetries = 3
	} else {
		if r, err := strconv.Atoi(retries); err == nil {
			Config.MaxRetries = r
		}
	}

	retryDelay := os.Getenv("RETRY_DELAY_MS")
	if retryDelay == "" {
		Config.RetryDelay = 100 * time.Millisecond
	} else {
		if ms, err := strconv.Atoi(retryDelay); err == nil {
			Config.RetryDelay = time.Duration(ms) * time.Millisecond
		}
	}

	balanceCheck := os.Getenv("BALANCE_CHECK_INTERVAL")
	if balanceCheck == "" {
		Config.BalanceCheck = 1 * time.Second
	} else {
		if sec, err := strconv.Atoi(balanceCheck); err == nil {
			Config.BalanceCheck = time.Duration(sec) * time.Second
		}
	}

	balanceFreq := os.Getenv("BALANCE_UPDATE_FREQ")
	if balanceFreq == "" {
		Config.BalanceFreq = 5 * time.Second
	} else {
		if sec, err := strconv.Atoi(balanceFreq); err == nil {
			Config.BalanceFreq = time.Duration(sec) * time.Second
		}
	}
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
				time.Sleep(100 * time.Millisecond)
				os.Unsetenv("WALLETS")
				os.Unsetenv("BASE_URL")
				os.Unsetenv("SERVER_PORT")
				os.Unsetenv("DIFFICULTY")
				os.Unsetenv("HTTP_TIMEOUT")
				os.Unsetenv("MAX_RETRIES")
				os.Unsetenv("RETRY_DELAY_MS")
				os.Unsetenv("BALANCE_CHECK_INTERVAL")
				os.Unsetenv("BALANCE_UPDATE_FREQ")
				if err := godotenv.Load(envPath); err != nil {
					pushLog("❌ env reload error", "error")
					continue
				}
				LoadConfig()
				reloadWallets()
				pushLog("♻️ Конфіг и гаманці перезавантажені", "info")
			}
		case err := <-watcher.Errors:
			pushLog("⚠️ watcher error: "+err.Error(), "error")
		}
	}
}
