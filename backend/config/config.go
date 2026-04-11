package config

type AppConfig struct {
	BaseURL      string `json:"base_url"`
	ServerPort   string `json:"server_port"`
	Difficulty   int    `json:"difficulty"`
	HTTPTimeout  int    `json:"http_timeout"`
	RetryDelayMs int    `json:"retry_delay_ms"`
	BalanceFreqS int    `json:"balance_freq_s"`
	MaxRetries   uint16 `json:"max_retries"`
	Threads      uint8  `json:"threads"`
}

var Config = AppConfig{
	BaseURL:      DefaultBaseURL,
	ServerPort:   DefaultServerPort,
	Difficulty:   DefaultDifficulty,
	HTTPTimeout:  int(DefaultHTTPTimeout.Seconds()),
	MaxRetries:   DefaultMaxRetries,
	RetryDelayMs: int(DefaultRetryDelay.Milliseconds()),
	BalanceFreqS: int(DefaultBalanceUpdateFreq.Seconds()),
	Threads:      4,
}

func (c *AppConfig) Update(newConf AppConfig) {
	if newConf.BaseURL != "" {
		c.BaseURL = newConf.BaseURL
	}

	if newConf.ServerPort != "" {
		c.ServerPort = newConf.ServerPort
	}

	if newConf.Difficulty > 0 {
		c.Difficulty = newConf.Difficulty
	}

	if newConf.HTTPTimeout > 0 {
		c.HTTPTimeout = newConf.HTTPTimeout
	} else {
		c.HTTPTimeout = int(DefaultHTTPTimeout.Seconds())
	}

	if newConf.MaxRetries > 0 {
		c.MaxRetries = newConf.MaxRetries
	}

	if newConf.RetryDelayMs > 0 {
		c.RetryDelayMs = newConf.RetryDelayMs
	} else {
		c.RetryDelayMs = int(DefaultRetryDelay.Milliseconds())
	}

	if newConf.BalanceFreqS > 0 {
		c.BalanceFreqS = newConf.BalanceFreqS
	} else {
		c.BalanceFreqS = int(DefaultBalanceUpdateFreq.Seconds())
	}

	if newConf.Threads > 0 {
		c.Threads = newConf.Threads
	}
}
