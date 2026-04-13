package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"shminer/backend/config"
	"shminer/backend/types"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type Storage struct {
	mu              sync.RWMutex
	currentStorage  types.StorageData
	sessionPassword string
}

func NewDriver() *Storage {
	return &Storage{}
}

func getSettingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return config.ConfigFileName, nil
	}

	appDir := filepath.Join(configDir, config.AppName)

	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, config.ConfigFileName), nil
}

func (s *Storage) GetConfigFilePath() string {
	path, err := getSettingsPath()
	if err != nil {
		return "Невідомий шлях"
	}
	return path
}

func (s *Storage) TryAutoLogin() (bool, error) {
	if !s.CheckExists() {
		return false, errors.New("not_found")
	}

	err := s.LoadStorage("")

	if err == nil {
		slog.Info("🔓 Автологін (Пароль не встановлено)")
		return true, nil
	}

	if err.Error() == "Невірний пароль" {
		return false, nil
	}

	return false, err
}

func (s *Storage) GetSessionPassword() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionPassword
}

func (s *Storage) GetStorage() types.StorageData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentStorage
}

func (s *Storage) UpdateWallets(newWallets []types.WalletStats) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range newWallets {
		newWallets[i].SessionMined = 0
	}

	s.currentStorage.Wallets = newWallets
}

func (s *Storage) ChangePassword(oldPass, newPass string) error {
	if oldPass != s.sessionPassword {
		return errors.New("Старий пароль невірний")
	}

	if err := s.SaveStorage(newPass, s.currentStorage); err != nil {
		return err
	}

	s.sessionPassword = newPass
	return nil
}

func (s *Storage) DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
}

func (s *Storage) InitStorage(password string) error {
	minerID := uuid.New().String()
	conf := config.Config
	conf.MinerID = minerID

	data := types.StorageData{
		Config:  conf,
		Wallets: []types.WalletStats{},
	}

	s.currentStorage = data
	s.sessionPassword = password
	s.applyLoadedData()

	err := s.SaveStorage(password, data)
	if err == nil {
		slog.Info("🆕 New miner registered", "id", minerID)
	}
	return err
}

func (s *Storage) SaveStorage(password string, data types.StorageData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentStorage = data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	key := s.DeriveKey(password, salt)
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

	configPath, err := getSettingsPath()
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, finalData, 0644)
}

func (s *Storage) LoadStorage(password string) error {
	configPath, err := getSettingsPath()
	fileData, err := os.ReadFile(configPath)
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

	key := s.DeriveKey(password, salt)
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

	var data types.StorageData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return err
	}

	s.currentStorage = data
	s.sessionPassword = password

	s.applyLoadedData()

	return nil
}

func (s *Storage) PersistConfig(password string) error {
	s.currentStorage.Config = config.Config
	return s.SaveStorage(password, s.currentStorage)
}

func (s *Storage) applyLoadedData() {
	if s.currentStorage.Config.MinerID == "" {
		s.currentStorage.Config.MinerID = uuid.New().String()
	}
	config.Config = s.currentStorage.Config
}

func (s *Storage) CheckExists() bool {
	configPath, err := getSettingsPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(configPath)
	return !os.IsNotExist(err)
}
