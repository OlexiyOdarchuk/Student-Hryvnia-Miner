package config

import (
	"errors"
	"shminer/backend"
	"shminer/backend/internal/stats"
	"shminer/backend/types"
	"time"
)

type AppConfig struct {
	BaseURL      string        `json:"base_url"`
	ServerPort   string        `json:"server_port"`
	Difficulty   int           `json:"difficulty"`
	HTTPTimeout  uint32        `json:"http_timeout"`
	MaxRetries   uint16        `json:"max_retries"`
	RetryDelayMs uint32        `json:"retry_delay_ms"`
	BalanceFreqS time.Duration `json:"balance_freq_s"`
	Threads      uint8         `json:"threads"`
}

func (c *AppConfig) UpdateConfig(password string, newConf types.AppConfig) error {
	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	if password != backend.sessionPassword {
		return errors.New("Невірний пароль")
	}

	backend.CurrentStorage.Config = newConf

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

	return backend.SaveStorage(password, backend.CurrentStorage)
}
