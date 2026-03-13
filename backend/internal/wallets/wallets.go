package wallets

import (
	"encoding/json"
	"errors"
	"shminer/backend/types"
	"sync"
)

type WalletExport struct {
	Name string `json:"name"`
	Pub  string `json:"pub"`
	Priv string `json:"priv"`
}

//go:generate mockgen -source=wallets.go -destination=mocks_test.go -package=wallets
type Storage interface {
	SaveStorage(password string, data types.StorageData) error
	GetStorage() types.StorageData
	UpdateWallets(newWallets []types.WalletStats)
	GetSessionPassword() string
}
type Wallets struct {
	Wallets       []string
	mu            *sync.RWMutex
	walletDataMap map[string]*types.WalletStats
	storage       Storage
}

func New(store Storage, mu *sync.RWMutex, walletData map[string]*types.WalletStats) *Wallets {
	return &Wallets{
		Wallets:       []string{},
		mu:            muOrNew(mu),
		walletDataMap: walletData,
		storage:       store,
	}
}

func muOrNew(mu *sync.RWMutex) *sync.RWMutex {
	if mu == nil {
		return &sync.RWMutex{}
	}
	return mu
}

func (w *Wallets) Load(snapshot types.StorageData) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for addr := range w.walletDataMap {
		delete(w.walletDataMap, addr)
	}
	w.Wallets = w.Wallets[:0]

	for _, entry := range snapshot.Wallets {
		walletCopy := entry
		w.Wallets = append(w.Wallets, walletCopy.Address)
		w.walletDataMap[walletCopy.Address] = &walletCopy
	}

	w.storage.UpdateWallets(snapshot.Wallets)
}

func (w *Wallets) ExportWalletJSON(address string) (string, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if stats, ok := w.walletDataMap[address]; ok {
		w := WalletExport{
			Name: stats.Name,
			Pub:  stats.Address,
			Priv: stats.PrivateKey,
		}
		by, _ := json.Marshal(w)
		return string(by), nil
	}
	return "", errors.New("wallet not found")
}

func (w *Wallets) GetWallets() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	cp := make([]string, len(w.Wallets))
	copy(cp, w.Wallets)
	return cp
}

func (w *Wallets) SyncStorage() {
	var list []types.WalletStats
	for _, addr := range w.Wallets {
		if stats, ok := w.walletDataMap[addr]; ok {
			list = append(list, *stats)
		}
	}
	w.storage.UpdateWallets(list)
}

func (w *Wallets) AddWalletSafe(name, address, privateKey string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, w := range w.walletDataMap {
		if w.Address == address {
			return nil
		}
		if w.Name == name {
			return errors.New("wallet already exists")
		}
	}

	w.Wallets = append(w.Wallets, address)
	w.walletDataMap[address] = &types.WalletStats{
		Address:    address,
		Name:       name,
		PrivateKey: privateKey,
		Working:    true,
	}

	w.SyncStorage()
	return w.storage.SaveStorage(w.storage.GetSessionPassword(), w.storage.GetStorage())
}

func (w *Wallets) DeleteWallet(address string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	newWallets := make([]string, 0, 2)
	for _, w := range w.Wallets {
		if w != address {
			newWallets = append(newWallets, w)
		}
	}
	w.Wallets = newWallets
	delete(w.walletDataMap, address)

	w.SyncStorage()
	return w.storage.SaveStorage(w.storage.GetSessionPassword(), w.storage.GetStorage())
}

func (w *Wallets) RenameWallet(address, newName string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for addr, w := range w.walletDataMap {
		if addr != address && w.Name == newName {
			return errors.New("wallet already exists")
		}
	}

	if stats, ok := w.walletDataMap[address]; ok {
		stats.Name = newName
		w.SyncStorage()
		return w.storage.SaveStorage(w.storage.GetSessionPassword(), w.storage.GetStorage())
	}
	return nil
}

func (w *Wallets) ToggleWalletMining(address string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if stats, ok := w.walletDataMap[address]; ok {
		stats.Working = !stats.Working

		w.SyncStorage()
		w.storage.SaveStorage(w.storage.GetSessionPassword(), w.storage.GetStorage())

		return stats.Working
	}
	return false
}

func (w *Wallets) SetAllMining(state bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, stats := range w.walletDataMap {
		stats.Working = state
	}
	w.SyncStorage()
	return w.storage.SaveStorage(w.storage.GetSessionPassword(), w.storage.GetStorage())
}

func (w *Wallets) UpdateWalletKey(address, privateKey string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if stats, ok := w.walletDataMap[address]; ok {
		stats.PrivateKey = privateKey
		w.SyncStorage()
		return w.storage.SaveStorage(w.storage.GetSessionPassword(), w.storage.GetStorage())
	}
	return nil
}

func (w *Wallets) GetPrivateKey(address string) string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if stats, ok := w.walletDataMap[address]; ok {
		return stats.PrivateKey
	}
	return ""
}
