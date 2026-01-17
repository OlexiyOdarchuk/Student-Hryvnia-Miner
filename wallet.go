package main

import (
	"encoding/json"
	"log"
	"os"
)

func getWallets() []string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	cp := make([]string, len(wallets))
	copy(cp, wallets)
	return cp
}

func getWalletName(address string) string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if stats, ok := walletDataMap[address]; ok && stats.Name != "" {
		return stats.Name
	}
	return "Worker"
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
				PrivateKey:    cfg.PrivateKey,
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
			Address:    stats.Address,
			Name:       stats.Name,
			PrivateKey: stats.PrivateKey,
			Working:    stats.Working,
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
