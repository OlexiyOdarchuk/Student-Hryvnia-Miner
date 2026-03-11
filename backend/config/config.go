package config

import (
	"errors"
	"shminer/backend/types"
	"time"
)

type Storage interface {
	SaveStorage(password string, data types.StorageData) error
}

type AppConfig struct {
	BaseURL      string        `json:"base_url"`
	ServerPort   string        `json:"server_port"`
	Difficulty   int           `json:"difficulty"`
	HTTPTimeout  time.Duration `json:"http_timeout"`
	MaxRetries   uint16        `json:"max_retries"`
	RetryDelayMs time.Duration `json:"retry_delay_ms"`
	BalanceFreqS time.Duration `json:"balance_freq_s"`
	Threads      uint8         `json:"threads"`
	storage      Storage
}

func (c *AppConfig) UpdateConfig(password string, newConf AppConfig) error {
	stats.dataMutex.Lock()
	defer stats.dataMutex.Unlock()

	if password != backend.sessionPassword {
		return errors.New("Невірний пароль")
	}

	storage.CurrentStorage.Config = newConf

	c.BaseURL = newConf.BaseURL
	c.ServerPort = newConf.ServerPort
	c.Difficulty = newConf.Difficulty
	c.MaxRetries = newConf.MaxRetries
	c.Threads = newConf.Threads

	if newConf.HTTPTimeout > 0 {
		c.HTTPTimeout = newConf.HTTPTimeout
	} else {
		c.HTTPTimeout = DefaultHTTPTimeout
	}

	if newConf.RetryDelayMs > 0 {
		c.RetryDelayMs = newConf.RetryDelayMs
	} else {
		c.RetryDelayMs = DefaultRetryDelay
	}

	if newConf.BalanceFreqS > 0 {
		c.BalanceFreqS = newConf.BalanceFreqS
	} else {
		c.BalanceFreqS = DefaultBalanceUpdateFreq
	}

	return c.storage.SaveStorage(password, c.storage.CurrentStorage)
}
