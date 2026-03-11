package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"os"
	"shminer/backend/config"
	"shminer/backend/types"
	"sync"

	"golang.org/x/crypto/argon2"
)

const StorageFile = "SHMinerSettings.bin"

type StorageData struct {
	Config  config.AppConfig    `json:"config"`
	Wallets []types.WalletStats `json:"wallets"`
}

var (
	CurrentStorage  StorageData
	sessionPassword string
)

type Storage struct {
	mu sync.RWMutex
}

func NewDriver() *Storage {
	return &Storage{}
}

func (s *Storage) GetSessionPassword() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return sessionPassword
}

func (s *Storage) SaveStorage(password string, data StorageData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	CurrentStorage = data
	return SaveStorage(password, data)
}

func (s *Storage) GetStorage() StorageData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return CurrentStorage
}

func (s *Storage) UpdateWallets(newWallets []types.WalletStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	CurrentStorage.Wallets = newWallets
}

func GetStorage() StorageData {
	return CurrentStorage
}

func GetSessionPassword() string {
	return sessionPassword
}

func ChangePassword(oldPass, newPass string) error {
	if oldPass != sessionPassword {
		return errors.New("Старий пароль невірний")
	}

	if err := SaveStorage(newPass, CurrentStorage); err != nil {
		return err
	}

	sessionPassword = newPass
	return nil
}

func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
}

func InitStorage(password string) error {
	data := StorageData{
		Config:  config.Config,
		Wallets: []types.WalletStats{},
	}

	CurrentStorage = data
	sessionPassword = password
	applyLoadedData()

	return SaveStorage(password, data)
}

func SaveStorage(password string, data StorageData) error {
	CurrentStorage = data
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

func PersistConfig(password string) error {
	CurrentStorage.Config = config.Config
	return SaveStorage(password, CurrentStorage)
}

func applyLoadedData() {
	config.Config = CurrentStorage.Config
}

func StorageExists() bool {
	_, err := os.Stat(StorageFile)
	return !os.IsNotExist(err)
}
