package backend

import (
	"sync"
)

var (
	Wallets      []string
	walletsMutex sync.RWMutex
)

func GetWallets() []string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	cp := make([]string, len(Wallets))
	copy(cp, Wallets)
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

func syncStorage() {
	var list []WalletStats
	for _, addr := range Wallets {
		if stats, ok := walletDataMap[addr]; ok {
			list = append(list, *stats)
		}
	}
	CurrentStorage.Wallets = list
}

func AddWalletSafe(name, address, privateKey string) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	for _, w := range Wallets {
		if w == address {
			return nil
		}
	}

	Wallets = append(Wallets, address)
	walletDataMap[address] = &WalletStats{
		Address:    address,
		Name:       name,
		PrivateKey: privateKey,
		Working:    true,
	}

	syncStorage()
	return SaveStorage(GetSessionPassword(), CurrentStorage)
}

func DeleteWallet(address string) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	newWallets := []string{}
	for _, w := range Wallets {
		if w != address {
			newWallets = append(newWallets, w)
		}
	}
	Wallets = newWallets
	delete(walletDataMap, address)

	syncStorage()
	return SaveStorage(GetSessionPassword(), CurrentStorage)
}

func RenameWallet(address, newName string) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	if stats, ok := walletDataMap[address]; ok {
		stats.Name = newName
		syncStorage()
		return SaveStorage(GetSessionPassword(), CurrentStorage)
	}
	return nil
}

func ToggleWalletMining(address string) bool {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	if stats, ok := walletDataMap[address]; ok {
		stats.Working = !stats.Working

		syncStorage()
		SaveStorage(GetSessionPassword(), CurrentStorage)

		return stats.Working
	}
	return false
}

func SetAllMining(state bool) {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	for _, stats := range walletDataMap {
		stats.Working = state
	}

	syncStorage()
	SaveStorage(GetSessionPassword(), CurrentStorage)
}

func UpdateWalletKey(address, privateKey string) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	if stats, ok := walletDataMap[address]; ok {
		stats.PrivateKey = privateKey
		syncStorage()
		return SaveStorage(GetSessionPassword(), CurrentStorage)
	}
	return nil
}

func GetPrivateKey(address string) string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if stats, ok := walletDataMap[address]; ok {
		return stats.PrivateKey
	}
	return ""
}
