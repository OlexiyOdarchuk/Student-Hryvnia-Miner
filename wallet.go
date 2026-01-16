package main

import (
	"os"
	"strings"
	"sync"
)

var walletNames = make(map[string]string)
var walletNamesMutex sync.RWMutex

func loadWalletsFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("WALLETS"))
	if raw == "" {
		panic("❌ ENV WALLETS не задано")
	}

	parts := strings.Split(raw, ",")
	var res []string
	for _, w := range parts {
		w = strings.TrimSpace(w)
		if len(w) < MinWalletAddressLen {
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

func setWalletName(address, name string) {
	walletNamesMutex.Lock()
	defer walletNamesMutex.Unlock()
	walletNames[address] = name
}

func getWalletName(address string) string {
	walletNamesMutex.RLock()
	defer walletNamesMutex.RUnlock()
	if name, ok := walletNames[address]; ok && name != "" {
		return name
	}
	return "Безімено"
}

func deleteWalletName(address string) {
	walletNamesMutex.Lock()
	defer walletNamesMutex.Unlock()
	delete(walletNames, address)
}
