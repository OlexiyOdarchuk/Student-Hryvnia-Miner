package main

import "time"

// Server configuration constants
const (
	DefaultServerPort        = ":8090"
	DefaultBaseURL           = "https://s-hryvnia-1.onrender.com"
	DefaultDifficulty        = 20
	DefaultHTTPTimeout       = 5 * time.Second
	DefaultMaxRetries        = 3
	DefaultRetryDelay        = 100 * time.Millisecond
	DefaultBalanceUpdateFreq = 5 * time.Second
	DefaultHashCacheTTL      = 300 * time.Millisecond
	DefaultSSEUpdateInterval = 200 * time.Millisecond
)

// HTTP client configuration
const (
	MaxIdleConnections      = 100
	MaxIdleConnsPerHost     = 10
	IdleConnTimeout         = 90 * time.Second
	ExponentialBackoffMaxMs = 30000
)

// Statistics and monitoring
const (
	HashrateHistorySize = 60
	LogRingBufferSize   = 100
	MaxTerminalLogs     = 200
	MinWalletAddressLen = 20
	MegahashDivisor     = 1000000
)

// Timings
const (
	MinerSleepInterval      = 170 * time.Millisecond
	EnvWatcherDebounce      = 100 * time.Millisecond
	SpeedMonitorInterval    = 1 * time.Second
	MaxExponentialBackoffMs = 30000
	EnvFileWatcherInitDelay = 100 * time.Millisecond
)

// UI Update intervals
const (
	LogHistoryMaxEntries = 100
	HashHistoryWindow    = 60
)
