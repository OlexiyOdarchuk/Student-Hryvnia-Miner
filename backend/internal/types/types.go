package types

type DashboardData struct {
	Hashrate       float64       `json:"hashrate"`
	SessionBlocks  int           `json:"session_blocks"`
	LifetimeBlocks int           `json:"lifetime_blocks"`
	Uptime         string        `json:"uptime"`
	TotalBalance   float64       `json:"total_balance"`
	Wallets        []WalletStats `json:"wallets"`
	NewLogs        []LogEntry    `json:"new_logs"`
}

type WalletStats struct {
	Address       string  `json:"address"`
	PrivateKey    string  `json:"private_key,omitempty"`
	Name          string  `json:"name"`
	SessionMined  int     `json:"session_mined"`
	TotalMined    int     `json:"total_mined"`
	ServerBalance float64 `json:"server_balance"`
	Working       bool    `json:"working"`
}

type LogEntry struct {
	ID      int64  `json:"id"`
	Time    string `json:"time"`
	Message string `json:"message"`
	Type    string `json:"type"` // "info", "success", "error"
}

type AppConfig struct {
	BaseURL      string `json:"base_url"`
	ServerPort   string `json:"server_port"`
	Difficulty   int    `json:"difficulty"`
	HTTPTimeout  int    `json:"http_timeout"`
	MaxRetries   int    `json:"max_retries"`
	RetryDelayMs int    `json:"retry_delay_ms"`
	BalanceFreqS int    `json:"balance_freq_s"`
	Threads      int    `json:"threads"`
}
