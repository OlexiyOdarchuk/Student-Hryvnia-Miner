package config

import (
	"testing"
)

func TestAppConfig_Update(t *testing.T) {
	config := AppConfig{
		MinerID:          "old-id",
		BaseURL:          DefaultBaseURL,
		ServerPort:       DefaultServerPort,
		Difficulty:       DefaultDifficulty,
		HTTPTimeout:      int(DefaultHTTPTimeout.Seconds()),
		MaxRetries:       DefaultMaxRetries,
		RetryDelayMs:     int(DefaultRetryDelay.Milliseconds()),
		BalanceFreqS:     int(DefaultBalanceUpdateFreq.Seconds()),
		BlockCheckFreqMs: DefaultBlockCheckFreqMs,
		Threads:          4,
	}

	newConf := AppConfig{
		MinerID:          "new-id",
		BaseURL:          "http://localhost:8080",
		ServerPort:       ":9090",
		Difficulty:       10,
		HTTPTimeout:      10,
		MaxRetries:       5,
		RetryDelayMs:     200,
		BalanceFreqS:     10,
		BlockCheckFreqMs: 1000,
		Threads:          8,
	}

	config.Update(newConf)

	if config.MinerID != newConf.MinerID {
		t.Errorf("Expected %s, got %s", newConf.MinerID, config.MinerID)
	}
	if config.BaseURL != newConf.BaseURL {
		t.Errorf("Expected %s, got %s", newConf.BaseURL, config.BaseURL)
	}
	if config.ServerPort != newConf.ServerPort {
		t.Errorf("Expected %s, got %s", newConf.ServerPort, config.ServerPort)
	}
	if config.Difficulty != newConf.Difficulty {
		t.Errorf("Expected %d, got %d", newConf.Difficulty, config.Difficulty)
	}
	if config.HTTPTimeout != newConf.HTTPTimeout {
		t.Errorf("Expected %d, got %d", newConf.HTTPTimeout, config.HTTPTimeout)
	}
	if config.MaxRetries != newConf.MaxRetries {
		t.Errorf("Expected %d, got %d", newConf.MaxRetries, config.MaxRetries)
	}
	if config.RetryDelayMs != newConf.RetryDelayMs {
		t.Errorf("Expected %d, got %d", newConf.RetryDelayMs, config.RetryDelayMs)
	}
	if config.BalanceFreqS != newConf.BalanceFreqS {
		t.Errorf("Expected %d, got %d", newConf.BalanceFreqS, config.BalanceFreqS)
	}
	if config.BlockCheckFreqMs != newConf.BlockCheckFreqMs {
		t.Errorf("Expected %d, got %d", newConf.BlockCheckFreqMs, config.BlockCheckFreqMs)
	}
	if config.Threads != newConf.Threads {
		t.Errorf("Expected %d, got %d", newConf.Threads, config.Threads)
	}
}

func TestAppConfig_Update_ZeroValues(t *testing.T) {
	config := AppConfig{
		MinerID:          "test-id",
		BaseURL:          "http://test",
		ServerPort:       ":9090",
		Difficulty:       10,
		HTTPTimeout:      10,
		MaxRetries:       5,
		RetryDelayMs:     200,
		BalanceFreqS:     10,
		BlockCheckFreqMs: 1000,
		Threads:          8,
	}

	newConf := AppConfig{
		MinerID:          "",
		BaseURL:          "",
		ServerPort:       "",
		Difficulty:       0,
		HTTPTimeout:      0,
		MaxRetries:       0,
		RetryDelayMs:     0,
		BalanceFreqS:     0,
		BlockCheckFreqMs: 0,
		Threads:          0,
	}

	config.Update(newConf)

	if config.MinerID != "test-id" {
		t.Errorf("Expected miner id not to change, got %s", config.MinerID)
	}
	if config.BaseURL != "http://test" {
		t.Errorf("Expected base url not to change, got %s", config.BaseURL)
	}
	if config.ServerPort != ":9090" {
		t.Errorf("Expected port not to change, got %s", config.ServerPort)
	}
	if config.Difficulty != 10 {
		t.Errorf("Expected difficulty not to change, got %d", config.Difficulty)
	}
	if config.HTTPTimeout != int(DefaultHTTPTimeout.Seconds()) {
		t.Errorf("Expected HTTPTimeout to be default, got %d", config.HTTPTimeout)
	}
	if config.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries not to change, got %d", config.MaxRetries)
	}
	if config.RetryDelayMs != int(DefaultRetryDelay.Milliseconds()) {
		t.Errorf("Expected RetryDelayMs to be default, got %d", config.RetryDelayMs)
	}
	if config.BalanceFreqS != int(DefaultBalanceUpdateFreq.Seconds()) {
		t.Errorf("Expected BalanceFreqS to be default, got %d", config.BalanceFreqS)
	}
	if config.BlockCheckFreqMs != DefaultBlockCheckFreqMs {
		t.Errorf("Expected BlockCheckFreqMs to be default, got %d", config.BlockCheckFreqMs)
	}
	if config.Threads != 8 {
		t.Errorf("Expected Threads not to change, got %d", config.Threads)
	}
}
