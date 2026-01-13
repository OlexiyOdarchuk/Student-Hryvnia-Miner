package main

type DashboardData struct {
	Hashrate     float64       `json:"hashrate"`
	TotalBlocks  int           `json:"total_blocks"`
	Uptime       string        `json:"uptime"`
	TotalBalance int           `json:"total_balance"`
	Wallets      []WalletStats `json:"wallets"`
	NewLogs      []LogEntry    `json:"new_logs"`
}

type WalletStats struct {
	Address       string `json:"address"`
	Short         string `json:"short"`
	SessionMined  int    `json:"session_mined"`
	ServerBalance int    `json:"server_balance"`
}

type LogEntry struct {
	ID      int64  `json:"id"`
	Time    string `json:"time"`
	Message string `json:"message"`
	Type    string `json:"type"` // "info", "success", "error"
}
