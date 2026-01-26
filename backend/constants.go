package backend

import "time"

const (
	DefaultServerPort        = ":8080"
	DefaultBaseURL           = "https://s-hryvnia-1.onrender.com"
	DefaultDifficulty        = 20
	DefaultHTTPTimeout       = 5 * time.Second
	DefaultMaxRetries        = 3
	DefaultRetryDelay        = 100 * time.Millisecond
	DefaultBalanceUpdateFreq = 5 * time.Second
	DefaultHashCacheTTL      = 300 * time.Millisecond
)

const (
	MaxIdleConnections      = 100
	MaxIdleConnsPerHost     = 10
	IdleConnTimeout         = 90 * time.Second
	ExponentialBackoffMaxMs = 30000
)

const (
	HashrateHistorySize = 60
	MegahashDivisor     = 1000000
)

const (
	MinerSleepInterval = 170 * time.Millisecond
)
