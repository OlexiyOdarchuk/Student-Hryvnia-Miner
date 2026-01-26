package backend

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/argon2"
)

const StorageFile = "SHMinerSettings.bin"

type StorageData struct {
	Config  AppConfig     `json:"config"`
	Wallets []WalletStats `json:"wallets"`
}

var CurrentStorage StorageData
var sessionPassword string

func GetSessionPassword() string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()
	return sessionPassword
}

func ChangePassword(oldPass, newPass string) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

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
	defaultConfig := AppConfig{
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
		Wallets: []WalletStats{},
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
	Config.BaseURL = CurrentStorage.Config.BaseURL
	Config.ServerPort = CurrentStorage.Config.ServerPort
	Config.Difficulty = CurrentStorage.Config.Difficulty
	Config.MaxRetries = CurrentStorage.Config.MaxRetries

	if CurrentStorage.Config.HTTPTimeout > 0 {
		Config.HTTPTimeout = time.Duration(CurrentStorage.Config.HTTPTimeout) * time.Second
	} else {
		Config.HTTPTimeout = DefaultHTTPTimeout
	}

	if CurrentStorage.Config.RetryDelayMs > 0 {
		Config.RetryDelay = time.Duration(CurrentStorage.Config.RetryDelayMs) * time.Millisecond
	} else {
		Config.RetryDelay = DefaultRetryDelay
	}

	if CurrentStorage.Config.BalanceFreqS > 0 {
		Config.BalanceFreq = time.Duration(CurrentStorage.Config.BalanceFreqS) * time.Second
	} else {
		Config.BalanceFreq = DefaultBalanceUpdateFreq
	}

	dataMutex.Lock()
	defer dataMutex.Unlock()

	Wallets = []string{}
	walletDataMap = make(map[string]*WalletStats)

	for _, w := range CurrentStorage.Wallets {
		Wallets = append(Wallets, w.Address)
		newStat := w
		newStat.SessionMined = 0
		walletDataMap[w.Address] = &newStat
	}
}

func StorageExists() bool {
	_, err := os.Stat(StorageFile)
	return !os.IsNotExist(err)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
