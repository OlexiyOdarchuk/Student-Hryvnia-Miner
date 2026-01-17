package main

import (
	"encoding/json"
	"log"
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

func loadWallets() {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	fileData, err := os.ReadFile(walletsFile)
	if os.IsNotExist(err) {
		_ = os.WriteFile(walletsFile, []byte("[]"), 0644)
		return
	}

	var configs []WalletConfig
	if err := json.Unmarshal(fileData, &configs); err != nil {
		log.Println("⚠️ Помилка читання wallets.json:", err)
		return
	}

	wallets = []string{}

	for _, cfg := range configs {
		wallets = append(wallets, cfg.Address)

		if _, exists := walletDataMap[cfg.Address]; !exists {
			walletDataMap[cfg.Address] = &WalletStats{
				Address:       cfg.Address,
				Name:          cfg.Name,
				Working:       cfg.Working,
				SessionMined:  0,
				ServerBalance: 0,
			}
		} else {

			walletDataMap[cfg.Address].Name = cfg.Name
			walletDataMap[cfg.Address].Working = cfg.Working
		}
	}

	log.Printf("📂 Завантажено %d гаманців з JSON", len(wallets))
}

func saveWallets() {
	dataMutex.RLock()

	var configs []WalletConfig

	for _, stats := range walletDataMap {
		configs = append(configs, WalletConfig{
			Address: stats.Address,
			Name:    stats.Name,
			Working: stats.Working,
		})
	}
	dataMutex.RUnlock()

	fileData, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		log.Println("⚠️ Помилка формування JSON:", err)
		return
	}

	err = os.WriteFile(walletsFile, fileData, 0644)
	if err != nil {
		log.Println("⚠️ Помилка запису файлу:", err)
	}
}
