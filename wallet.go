package main

import (
	"os"
	"strings"
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

func getWallets() []string {
	walletsMutex.RLock()
	defer walletsMutex.RUnlock()
	cp := make([]string, len(wallets))
	copy(cp, wallets)
	return cp
}
