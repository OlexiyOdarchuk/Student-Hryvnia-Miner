package wallets

import (
	"fmt"
	"shminer/backend"
	"shminer/backend/types"
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
	stats.dataMutex.RLock()
	defer stats.dataMutex.RUnlock()

	if stats, ok := stats.walletDataMap[address]; ok {
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
	stats.dataMutex.RLock()
	defer stats.dataMutex.RUnlock()

	cp := make([]string, len(Wallets))
	copy(cp, Wallets)
	return cp
}

func syncStorage() {
	var list []types.WalletStats
	for _, addr := range Wallets {
		if stats, ok := stats.walletDataMap[addr]; ok {
			list = append(list, *stats)
		}
	}
	backend.CurrentStorage.Wallets = list
}

func AddWalletSafe(name, address, privateKey string) error {
	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	for _, w := range stats.walletDataMap {
		if w.Address == address {
			return nil
		}
		if w.Name == name {
			return fmt.Errorf("гаманець з назвою '%s' вже існує", name)
		}
	}

	Wallets = append(Wallets, address)
	stats.walletDataMap[address] = &types.WalletStats{
		Address:    address,
		Name:       name,
		PrivateKey: privateKey,
		Working:    true,
	}

	syncStorage()
	return backend.SaveStorage(backend.sessionPassword, backend.CurrentStorage)
}

func DeleteWallet(address string) error {
	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	newWallets := []string{}
	for _, w := range Wallets {
		if w != address {
			newWallets = append(newWallets, w)
		}
	}
	Wallets = newWallets
	delete(stats.walletDataMap, address)

	syncStorage()
	return backend.SaveStorage(backend.sessionPassword, backend.CurrentStorage)
}

func RenameWallet(address, newName string) error {
	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	for addr, w := range stats.walletDataMap {
		if addr != address && w.Name == newName {
			return fmt.Errorf("назва '%s' вже використовується", newName)
		}
	}

	if stats, ok := stats.walletDataMap[address]; ok {
		stats.Name = newName
		syncStorage()
		return backend.SaveStorage(backend.sessionPassword, backend.CurrentStorage)
	}
	return nil
}

func ToggleWalletMining(address string) bool {
	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	if stats, ok := stats.walletDataMap[address]; ok {
		stats.Working = !stats.Working

		syncStorage()
		backend.SaveStorage(backend.sessionPassword, backend.CurrentStorage)

		return stats.Working
	}
	return false
}

func SetAllMining(state bool) {
	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	for _, stats := range stats.walletDataMap {
		stats.Working = state
	}

	syncStorage()
	backend.SaveStorage(backend.sessionPassword, backend.CurrentStorage)
}

func UpdateWalletKey(address, privateKey string) error {
	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	if stats, ok := stats.walletDataMap[address]; ok {
		stats.PrivateKey = privateKey
		syncStorage()
		return backend.SaveStorage(backend.sessionPassword, backend.CurrentStorage)
	}
	return nil
}

func GetPrivateKey(address string) string {
	stats.dataMutex.RLock()
	defer stats.dataMutex.RUnlock()

	if stats, ok := stats.walletDataMap[address]; ok {
		return stats.PrivateKey
	}
	return ""
}
