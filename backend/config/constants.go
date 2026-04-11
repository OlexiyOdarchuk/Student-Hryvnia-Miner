package config

import "time"

const (
	DefaultServerPort        = ":8080"
	DefaultBaseURL           = "https://s-hryvnia-1.onrender.com"
	DefaultDifficulty        = 20
	DefaultHTTPTimeout       = 5 * time.Second
	DefaultMaxRetries        = 3
	DefaultRetryDelay        = 100 * time.Millisecond
	DefaultBalanceUpdateFreq = 5 * time.Second
	DefaultBlockCheckFreqMs  = 5000
)

const (
	ExponentialBackoffMaxMs = 30000
)
