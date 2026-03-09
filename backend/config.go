package backend

import (
	"errors"
	"shminer/backend/internal/stats"
	"shminer/backend/types"
	"time"
)

var Config struct {
	BaseURL     string
	ServerPort  string
	Difficulty  int
	HTTPTimeout time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	BalanceFreq time.Duration
	Threads     int
}

func UpdateConfig(password string, newConf types.AppConfig) error {
	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	if password != sessionPassword {
		return errors.New("Невірний пароль")
	}

	CurrentStorage.Config = newConf

	Config.BaseURL = newConf.BaseURL
	Config.ServerPort = newConf.ServerPort
	Config.Difficulty = newConf.Difficulty
	Config.MaxRetries = newConf.MaxRetries
	Config.Threads = newConf.Threads

	if newConf.HTTPTimeout > 0 {
		Config.HTTPTimeout = time.Duration(newConf.HTTPTimeout) * time.Second
	} else {
		Config.HTTPTimeout = DefaultHTTPTimeout
	}

	if newConf.RetryDelayMs > 0 {
		Config.RetryDelay = time.Duration(newConf.RetryDelayMs) * time.Millisecond
	} else {
		Config.RetryDelay = DefaultRetryDelay
	}

	if newConf.BalanceFreqS > 0 {
		Config.BalanceFreq = time.Duration(newConf.BalanceFreqS) * time.Second
	} else {
		Config.BalanceFreq = DefaultBalanceUpdateFreq
	}

	return SaveStorage(password, CurrentStorage)
}
