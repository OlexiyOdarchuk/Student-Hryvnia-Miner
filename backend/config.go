package backend

import (
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

func LoadConfig() {
	Config.BaseURL = DefaultBaseURL
	Config.ServerPort = DefaultServerPort
	Config.Difficulty = DefaultDifficulty
	Config.HTTPTimeout = DefaultHTTPTimeout
	Config.MaxRetries = DefaultMaxRetries
	Config.RetryDelay = DefaultRetryDelay
	Config.BalanceFreq = DefaultBalanceUpdateFreq
	Config.Threads = 0 // 0 means auto/max
}

func UpdateConfig(password string, newConf AppConfig) error {

	err := LoadStorage(password)
	if err != nil {
		return err
	}

	CurrentStorage.Config = newConf

	applyLoadedData()

	return SaveStorage(password, CurrentStorage)
}
