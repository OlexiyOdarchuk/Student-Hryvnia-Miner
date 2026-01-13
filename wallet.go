package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
)

func loadWalletsFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("WALLETS"))
	if raw == "" {
		panic("❌ ENV WALLETS не задано")
	}

	parts := strings.Split(raw, ",")
	var res []string
	for _, w := range parts {
		w = strings.TrimSpace(w)
		if len(w) < 20 {
			panic("❌ Некоректний гаманець: " + w)
		}
		res = append(res, w)
	}

	if len(res) == 0 {
		panic("❌ Порожній список гаманців")
	}

	return res
}

func reloadWallets() {
	newWallets := loadWalletsFromEnv()

	walletsMutex.Lock()
	defer walletsMutex.Unlock()

	dataMutex.Lock()
	defer dataMutex.Unlock()

	// додати нові
	for _, w := range newWallets {
		if _, ok := walletDataMap[w]; !ok {
			walletDataMap[w] = &WalletStats{
				Address: w,
				Short:   "..." + w[len(w)-8:],
			}
			pushLog("➕ Додано гаманець "+w[len(w)-6:], "info")
		}
	}

	// видалити відсутні
	for w := range walletDataMap {
		if !contains(newWallets, w) {
			delete(walletDataMap, w)
			pushLog("➖ Видалено гаманець "+w[len(w)-6:], "info")
		}
	}

	wallets = newWallets
}

func contains(arr []string, v string) bool {
	for _, x := range arr {
		if x == v {
			return true
		}
	}
	return false
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
				time.Sleep(150 * time.Millisecond)
				os.Unsetenv("WALLETS")
				if err := godotenv.Load(envPath); err != nil {
					pushLog("❌ env reload error", "error")
					continue
				}
				reloadWallets()
			}
		case err := <-watcher.Errors:
			pushLog("⚠️ watcher error: "+err.Error(), "error")
		}
	}
}

func getWallets() []string {
	walletsMutex.RLock()
	defer walletsMutex.RUnlock()
	cp := make([]string, len(wallets))
	copy(cp, wallets)
	return cp
}
