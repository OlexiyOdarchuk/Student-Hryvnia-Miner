package config

type AppConfig struct {
	MinerID          string           `json:"miner_id"`
	TelegramHandle   string           `json:"telegram_handle"`
	BaseURL          string           `json:"base_url"`
	ServerPort       string           `json:"server_port"`
	Difficulty       int              `json:"difficulty"`
	HTTPTimeout      int              `json:"http_timeout"`
	RetryDelayMs     int              `json:"retry_delay_ms"`
	BalanceFreqS     int              `json:"balance_freq_s"`
	BlockCheckFreqMs int              `json:"block_check_freq_ms"`
	SubmitBufferSize int              `json:"submit_buffer_size"`
	MaxRetries       uint16           `json:"max_retries"`
	Threads          uint8            `json:"threads"`
	Automation       AutomationConfig `json:"automation"`
}

type AutomationConfig struct {
	TelegramBotToken   string `json:"telegram_bot_token"`
	TelegramChatID     string `json:"telegram_chat_id"`
	ScheduleStart      string `json:"schedule_start"`
	ScheduleStop       string `json:"schedule_stop"`
	BlockTarget        uint32 `json:"block_target"`
	SessionMinutes     uint32 `json:"session_minutes"`
	ProgressNotifyStep uint32 `json:"progress_notify_step"`
	ScheduleEnabled    bool   `json:"schedule_enabled"`
	NotifyOnStop       bool   `json:"notify_on_stop"`
	NotifyOnTarget     bool   `json:"notify_on_target"`
	NotifyOnStart      bool   `json:"notify_on_start"`
}

var Config = AppConfig{
	BaseURL:          DefaultBaseURL,
	ServerPort:       DefaultServerPort,
	Difficulty:       DefaultDifficulty,
	HTTPTimeout:      int(DefaultHTTPTimeout.Seconds()),
	MaxRetries:       DefaultMaxRetries,
	RetryDelayMs:     int(DefaultRetryDelay.Milliseconds()),
	BalanceFreqS:     int(DefaultBalanceUpdateFreq.Seconds()),
	BlockCheckFreqMs: DefaultBlockCheckFreqMs,
	SubmitBufferSize: DefaultSubmitBufferSize,
	Threads:          4,
}

func (c *AppConfig) Update(newConf AppConfig) {
	c.Automation = newConf.Automation

	if newConf.MinerID != "" {
		c.MinerID = newConf.MinerID
	}

	if newConf.TelegramHandle != "" {
		c.TelegramHandle = newConf.TelegramHandle
	}

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

	if newConf.BlockCheckFreqMs > 0 {
		c.BlockCheckFreqMs = newConf.BlockCheckFreqMs
	} else {
		c.BlockCheckFreqMs = DefaultBlockCheckFreqMs
	}

	if newConf.SubmitBufferSize > 0 {
		c.SubmitBufferSize = newConf.SubmitBufferSize
	} else {
		c.SubmitBufferSize = DefaultSubmitBufferSize
	}

	if newConf.Threads > 0 {
		c.Threads = newConf.Threads
	}
}
