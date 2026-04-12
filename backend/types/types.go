package types

import (
	"shminer/backend/config"
	"sync"
	"sync/atomic"
	"time"
)

type DashboardData struct {
	NewLogs        []LogEntry    `json:"new_logs"`
	Wallets        []WalletStats `json:"wallets"`
	Hashrate       float64       `json:"hashrate"`
	TotalBalance   float64       `json:"total_balance"`
	SessionBlocks  uint32        `json:"session_blocks"`
	LifetimeBlocks uint32        `json:"lifetime_blocks"`
	Uptime         string        `json:"uptime"`
}

type WalletStats struct {
	ServerBalance float64 `json:"server_balance"`
	Address       string  `json:"address"`
	PrivateKey    string  `json:"private_key,omitempty"`
	Name          string  `json:"name"`
	SessionMined  uint32  `json:"session_mined"`
	TotalMined    uint32  `json:"total_mined"`
	Working       bool    `json:"working"`
}

type LogEntry struct {
	ID      int64  `json:"id"`
	Time    string `json:"time"`
	Message string `json:"message"`
	Type    string `json:"type"`
}
type Stats struct {
	HashrateHistory   [60]float64   `json:"hashrate_history"`
	StartTime         time.Time     `json:"start_time"`
	GlobalHashrate    atomic.Value  `json:"global_hashrate"`
	HashrateHistMutex sync.Mutex    `json:"hashrate_hist_mutex"`
	HashCount         atomic.Uint32 `json:"hash_count"`
	HashrateHistPos   int           `json:"hashrate_hist_pos"`
	SessionMined      uint32        `json:"session_mined"`
}

type StorageData struct {
	Wallets []WalletStats    `json:"wallets"`
	Config  config.AppConfig `json:"config"`
}

type TxPayload struct {
	From      string `json:"from"`
	Signature string `json:"signature,omitempty"`
	To        string `json:"to"`
	Amount    int    `json:"amount"`
	Fee       int    `json:"fee"`
}
