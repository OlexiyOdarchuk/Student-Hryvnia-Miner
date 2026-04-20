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

func (w *Wallets) snapshotLocked() []types.WalletStats {
	list := make([]types.WalletStats, 0, len(w.Wallets))
	for _, addr := range w.Wallets {
		if stats, ok := w.walletDataMap[addr]; ok {
			list = append(list, *stats)
		}
	}
	return list
}

func (w *Wallets) Load(snapshot types.StorageData) {
	w.mu.Lock()
	for addr := range w.walletDataMap {
		delete(w.walletDataMap, addr)
	}
	w.Wallets = w.Wallets[:0]

	for _, entry := range snapshot.Wallets {
		walletCopy := entry
		w.Wallets = append(w.Wallets, walletCopy.Address)
		w.walletDataMap[walletCopy.Address] = &walletCopy
	}
	list := w.snapshotLocked()
	w.mu.Unlock()

	w.storage.UpdateWallets(list)
}

func (w *Wallets) ExportWalletJSON(address string) (string, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if stats, ok := w.walletDataMap[address]; ok {
		out := WalletExport{
			Name: stats.Name,
			Pub:  stats.Address,
			Priv: stats.PrivateKey,
		}
		by, _ := json.Marshal(out)
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
	w.mu.RLock()
	list := w.snapshotLocked()
	w.mu.RUnlock()

	w.storage.UpdateWallets(list)
}

func (w *Wallets) AddWalletSafe(name, address, privateKey string) error {
	w.mu.Lock()
	for _, stats := range w.walletDataMap {
		if stats.Address == address {
			w.mu.Unlock()
			return nil
		}
		if stats.Name == name {
			w.mu.Unlock()
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
	list := w.snapshotLocked()
	w.mu.Unlock()

	return w.persist(list)
}

func (w *Wallets) DeleteWallet(address string) error {
	w.mu.Lock()
	newWallets := make([]string, 0, len(w.Wallets))
	for _, addr := range w.Wallets {
		if addr != address {
			newWallets = append(newWallets, addr)
		}
	}
	w.Wallets = newWallets
	delete(w.walletDataMap, address)

	list := w.snapshotLocked()
	w.mu.Unlock()

	return w.persist(list)
}

func (w *Wallets) RenameWallet(address, newName string) error {
	w.mu.Lock()
	for addr, stats := range w.walletDataMap {
		if addr != address && stats.Name == newName {
			w.mu.Unlock()
			return errors.New("wallet already exists")
		}
	}

	stats, ok := w.walletDataMap[address]
	if !ok {
		w.mu.Unlock()
		return nil
	}
	stats.Name = newName
	list := w.snapshotLocked()
	w.mu.Unlock()

	return w.persist(list)
}

func (w *Wallets) ToggleWalletMining(address string) bool {
	w.mu.Lock()
	stats, ok := w.walletDataMap[address]
	if !ok {
		w.mu.Unlock()
		return false
	}
	stats.Working = !stats.Working
	newState := stats.Working
	list := w.snapshotLocked()
	w.mu.Unlock()

	_ = w.persist(list)
	return newState
}

func (w *Wallets) SetAllMining(state bool) error {
	w.mu.Lock()
	for _, stats := range w.walletDataMap {
		stats.Working = state
	}
	list := w.snapshotLocked()
	w.mu.Unlock()

	return w.persist(list)
}

func (w *Wallets) UpdateWalletKey(address, privateKey string) error {
	w.mu.Lock()
	stats, ok := w.walletDataMap[address]
	if !ok {
		w.mu.Unlock()
		return nil
	}
	stats.PrivateKey = privateKey
	list := w.snapshotLocked()
	w.mu.Unlock()

	return w.persist(list)
}

func (w *Wallets) GetPrivateKey(address string) string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if stats, ok := w.walletDataMap[address]; ok {
		return stats.PrivateKey
	}
	return ""
}

func (w *Wallets) persist(list []types.WalletStats) error {
	w.storage.UpdateWallets(list)
	return w.storage.SaveStorage(w.storage.GetSessionPassword(), w.storage.GetStorage())
}
