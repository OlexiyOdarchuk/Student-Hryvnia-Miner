package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"os"
	"shminer/backend/internal/wallets"
	"shminer/backend/types"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

const StorageFile = "SHMinerSettings.bin"

type Storage struct {
	currentStorage  types.StorageData
	sessionPassword string
	mu              sync.RWMutex
}

func (s *Storage) GetSessionPassword() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionPassword
}

func ChangePassword(oldPass, newPass string) error {
	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	if oldPass != sessionPassword {
		return errors.New("Старий пароль невірний")
	}

	err := SaveStorage(newPass, CurrentStorage)
	if err != nil {
		return err
	}

	sessionPassword = newPass
	return nil
}

func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
}

func InitStorage(password string) error {
	defaultConfig := types.AppConfig{
		BaseURL:      DefaultBaseURL,
		ServerPort:   DefaultServerPort,
		Difficulty:   DefaultDifficulty,
		HTTPTimeout:  int(DefaultHTTPTimeout.Seconds()),
		MaxRetries:   DefaultMaxRetries,
		RetryDelayMs: int(DefaultRetryDelay.Milliseconds()),
		BalanceFreqS: int(DefaultBalanceUpdateFreq.Seconds()),
	}

	data := StorageData{
		Config:  defaultConfig,
		Wallets: []types.WalletStats{},
	}

	CurrentStorage = data
	sessionPassword = password
	applyLoadedData()

	return SaveStorage(password, data)
}

func SaveStorage(password string, data StorageData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	key := DeriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, jsonData, nil)

	finalData := append(salt, ciphertext...)

	return os.WriteFile(StorageFile, finalData, 0644)
}

func LoadStorage(password string) error {
	fileData, err := os.ReadFile(StorageFile)
	if os.IsNotExist(err) {
		return errors.New("not_found")
	}
	if err != nil {
		return err
	}

	if len(fileData) < 16 {
		return errors.New("invalid file format")
	}

	salt := fileData[:16]
	ciphertext := fileData[16:]

	key := DeriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return errors.New("ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return errors.New("Невірний пароль")
	}

	var data StorageData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return err
	}

	CurrentStorage = data
	sessionPassword = password

	applyLoadedData()

	return nil
}

func applyLoadedData() {
	config.Config.BaseURL = CurrentStorage.Config.BaseURL
	config.Config.ServerPort = CurrentStorage.Config.ServerPort
	config.Config.Difficulty = CurrentStorage.Config.Difficulty
	config.Config.MaxRetries = CurrentStorage.Config.MaxRetries

	if CurrentStorage.Config.HTTPTimeout > 0 {
		config.Config.HTTPTimeout = time.Duration(CurrentStorage.Config.HTTPTimeout) * time.Second
	} else {
		config.Config.HTTPTimeout = DefaultHTTPTimeout
	}

	if CurrentStorage.Config.RetryDelayMs > 0 {
		config.Config.RetryDelay = time.Duration(CurrentStorage.Config.RetryDelayMs) * time.Millisecond
	} else {
		config.Config.RetryDelay = DefaultRetryDelay
	}

	if CurrentStorage.Config.BalanceFreqS > 0 {
		config.Config.BalanceFreq = time.Duration(CurrentStorage.Config.BalanceFreqS) * time.Second
	} else {
		config.Config.BalanceFreq = DefaultBalanceUpdateFreq
	}

	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	wallets.Wallets = []string{}
	stats.walletDataMap = make(map[string]*types.WalletStats)

	for _, w := range CurrentStorage.Wallets {
		wallets.Wallets = append(wallets.Wallets, w.Address)
		newStat := w
		newStat.SessionMined = 0
		stats.walletDataMap[w.Address] = &newStat
	}
}

func StorageExists() bool {
	_, err := os.Stat(StorageFile)
	return !os.IsNotExist(err)
}

func (s *Storage) GetStorage() types.StorageData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentStorage
}

func (s *Storage) UpdateWallets(newWallets []types.WalletStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentStorage.Wallets = newWallets
}
