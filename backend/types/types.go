package types

import (
	"shminer/backend/config"
	"sync"
	"sync/atomic"
	"time"
)

type DashboardData struct {
	Hashrate       float64       `json:"hashrate"`
	SessionBlocks  uint32        `json:"session_blocks"`
	LifetimeBlocks uint32        `json:"lifetime_blocks"`
	Uptime         string        `json:"uptime"`
	TotalBalance   float64       `json:"total_balance"`
	Wallets        []WalletStats `json:"wallets"`
	NewLogs        []LogEntry    `json:"new_logs"`
}

type WalletStats struct {
	Address       string  `json:"address"`
	PrivateKey    string  `json:"private_key,omitempty"`
	Name          string  `json:"name"`
	SessionMined  uint32  `json:"session_mined"`
	TotalMined    uint32  `json:"total_mined"`
	ServerBalance float64 `json:"server_balance"`
	Working       bool    `json:"working"`
}

type LogEntry struct {
	ID      int64  `json:"id"`
	Time    string `json:"time"`
	Message string `json:"message"`
	Type    string `json:"type"`
}
type Stats struct {
	StartTime         time.Time     `json:"start_time"`
	GlobalHashrate    atomic.Value  `json:"global_hashrate"`
	HashrateHistory   [60]float64   `json:"hashrate_history"`
	HashrateHistPos   int           `json:"hashrate_hist_pos"`
	HashrateHistMutex sync.Mutex    `json:"hashrate_hist_mutex"`
	HashCount         atomic.Uint32 `json:"hash_count"`
	SessionMined      uint32        `json:"session_mined"`
}

type StorageData struct {
	Config  *config.AppConfig
	Wallets []WalletStats
}
