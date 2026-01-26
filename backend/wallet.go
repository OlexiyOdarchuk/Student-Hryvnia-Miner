package backend

import (
	"fmt"
	"sync"
)

type WalletExport struct {
	Name string `json:"name"`
	Pub  string `json:"pub"`
	Priv string `json:"priv"`
}

var (
	Wallets      []string
	walletsMutex sync.RWMutex
)

func ExportWalletJSON(address string) (string, error) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if stats, ok := walletDataMap[address]; ok {
		w := WalletExport{
			Name: stats.Name,
			Pub:  stats.Address,
			Priv: stats.PrivateKey,
		}
		return fmt.Sprintf(`{"name":"%s","pub":"%s","priv":"%s"}`, w.Name, w.Pub, w.Priv), nil
	}
	return "", fmt.Errorf("wallet not found")
}

func GetWallets() []string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	cp := make([]string, len(Wallets))
	copy(cp, Wallets)
	return cp
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

	for _, w := range walletDataMap {
		if w.Address == address {
			return nil
		}
		if w.Name == name {
			return fmt.Errorf("гаманець з назвою '%s' вже існує", name)
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
	return SaveStorage(sessionPassword, CurrentStorage)
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
	return SaveStorage(sessionPassword, CurrentStorage)
}

func RenameWallet(address, newName string) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	for addr, w := range walletDataMap {
		if addr != address && w.Name == newName {
			return fmt.Errorf("назва '%s' вже використовується", newName)
		}
	}

	if stats, ok := walletDataMap[address]; ok {
		stats.Name = newName
		syncStorage()
		return SaveStorage(sessionPassword, CurrentStorage)
	}
	return nil
}

func ToggleWalletMining(address string) bool {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	if stats, ok := walletDataMap[address]; ok {
		stats.Working = !stats.Working

		syncStorage()
		SaveStorage(sessionPassword, CurrentStorage)

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
	SaveStorage(sessionPassword, CurrentStorage)
}

func UpdateWalletKey(address, privateKey string) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	if stats, ok := walletDataMap[address]; ok {
		stats.PrivateKey = privateKey
		syncStorage()
		return SaveStorage(sessionPassword, CurrentStorage)
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
