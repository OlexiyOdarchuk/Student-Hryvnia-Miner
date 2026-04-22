package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"shminer/backend/config"
	"shminer/backend/internal/telegram"
	"shminer/backend/types"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type DashboardProvider interface {
	GetDashboardData() types.DashboardData
}

type Controller = telegram.Controller

type Engine struct {
	ctrl           Controller
	dash           DashboardProvider
	httpClient     *http.Client
	bot            *telegram.Bot
	botMu          sync.Mutex
	sessionStart   atomic.Int64
	lastErrorSent  atomic.Int64
	notifiedTarget atomic.Bool
	scheduleInit   atomic.Bool
	lastScheduleOn atomic.Bool
}

func New(ctrl Controller, dash DashboardProvider) *Engine {
	e := &Engine{
		ctrl:       ctrl,
		dash:       dash,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	e.sessionStart.Store(time.Now().Unix())
	return e
}

func (e *Engine) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	e.ResetSession()
	e.notifyIfEnabled(func(a config.AutomationConfig) bool { return a.NotifyOnStart },
		"🚀 Майнінг запущено")

	e.startBot(ctx)

	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			e.notifyIfEnabled(func(a config.AutomationConfig) bool { return a.NotifyOnStop },
				"🛑 Майнінг зупинено")
			e.stopBot()
			return
		case <-tick.C:
			e.checkRules()
		}
	}
}

func (e *Engine) startBot(ctx context.Context) {
	auto := config.Config.Automation
	if auto.TelegramBotToken == "" || auto.TelegramChatID == "" {
		return
	}
	e.botMu.Lock()
	defer e.botMu.Unlock()
	if e.bot != nil {
		return
	}
	bot, err := telegram.Start(ctx, auto.TelegramBotToken, auto.TelegramChatID, e.ctrl)
	if err != nil {
		slog.Debug("Telegram bot not started", "err", err)
		return
	}
	e.bot = bot
}

func (e *Engine) stopBot() {
	e.botMu.Lock()
	defer e.botMu.Unlock()
	if e.bot != nil {
		e.bot.Stop()
		e.bot = nil
	}
}

func (e *Engine) ResetSession() {
	e.sessionStart.Store(time.Now().Unix())
	e.notifiedTarget.Store(false)
}

func (e *Engine) checkRules() {
	auto := config.Config.Automation
	data := e.dash.GetDashboardData()
	mining := e.ctrl.IsMining()

	if auto.BlockTarget > 0 && mining && data.SessionBlocks >= auto.BlockTarget {
		if !e.notifiedTarget.Swap(true) {
			msg := fmt.Sprintf("🎯 Досягнуто цілі: зараховано %d блоків. Зупиняю майнінг.", data.SessionBlocks)
			slog.Info(msg)
			if auto.NotifyOnTarget {
				e.sendTelegram(msg)
			}
			e.ctrl.SetMining(false)
			return
		}
	}

	if auto.SessionMinutes > 0 && mining {
		start := time.Unix(e.sessionStart.Load(), 0)
		if time.Since(start) >= time.Duration(auto.SessionMinutes)*time.Minute {
			msg := fmt.Sprintf("⏱️ Таймер сесії %d хв вичерпано. Зупиняю майнінг.", auto.SessionMinutes)
			slog.Info(msg)
			if auto.NotifyOnStop {
				e.sendTelegram(msg)
			}
			e.ctrl.SetMining(false)
			return
		}
	}

	if auto.ScheduleEnabled && auto.ScheduleStart != "" && auto.ScheduleStop != "" {
		shouldBeOn := withinSchedule(time.Now(), auto.ScheduleStart, auto.ScheduleStop)
		hadPrev := e.scheduleInit.Swap(true)
		prev := e.lastScheduleOn.Swap(shouldBeOn)
		if !hadPrev {
			prev = shouldBeOn
		}
		if shouldBeOn && !mining {
			slog.Info("📅 Розклад: час запускати майнінг")
			if auto.NotifyOnStart {
				e.sendTelegram("📅 За розкладом: майнінг запускається")
			}
			e.ResetSession()
			e.ctrl.SetMining(true)
		} else if !shouldBeOn && mining && prev {
			slog.Info("📅 Розклад: час зупиняти майнінг")
			if auto.NotifyOnStop {
				e.sendTelegram("📅 За розкладом: майнінг зупиняється")
			}
			e.ctrl.SetMining(false)
		}
	}
}

func (e *Engine) NotifyError(message string) {
	auto := config.Config.Automation
	if !auto.NotifyOnError {
		return
	}
	now := time.Now().Unix()
	last := e.lastErrorSent.Load()
	if now-last < 60 {
		return
	}
	if !e.lastErrorSent.CompareAndSwap(last, now) {
		return
	}
	e.sendTelegram("⚠️ " + message)
}

func (e *Engine) notifyIfEnabled(enabled func(config.AutomationConfig) bool, message string) {
	if enabled(config.Config.Automation) {
		e.sendTelegram(message)
	}
}

func (e *Engine) SendTestMessage() error {
	auto := config.Config.Automation
	if auto.TelegramBotToken == "" || auto.TelegramChatID == "" {
		return fmt.Errorf("bot token and chat ID are required")
	}
	return e.doSendTelegram(auto.TelegramBotToken, auto.TelegramChatID, "✅ Тестове повідомлення від S-UAH Miner")
}

func (e *Engine) SendTestWithConfig(token, chatID string) error {
	if token == "" || chatID == "" {
		return fmt.Errorf("token and chat_id required")
	}
	return e.doSendTelegram(token, chatID, "✅ Перевірка з'єднання з S-UAH Miner")
}

func (e *Engine) sendTelegram(message string) {
	auto := config.Config.Automation
	if auto.TelegramBotToken == "" || auto.TelegramChatID == "" {
		return
	}
	if err := e.doSendTelegram(auto.TelegramBotToken, auto.TelegramChatID, message); err != nil {
		slog.Debug("Telegram send failed", "err", err)
	}
}

func (e *Engine) doSendTelegram(token, chatID, message string) error {
	endpoint := "https://api.telegram.org/bot" + token + "/sendMessage"
	payload := map[string]string{
		"chat_id": chatID,
		"text":    message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram api: %d", resp.StatusCode)
	}
	return nil
}

func (e *Engine) ResolveChatID(token, chatID string) (string, error) {
	endpoint := "https://api.telegram.org/bot" + token + "/getChat?chat_id=" + url.QueryEscape(chatID)
	resp, err := e.httpClient.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		Ok     bool `json:"ok"`
		Result struct {
			ID int64 `json:"id"`
		} `json:"result"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if !out.Ok {
		return "", fmt.Errorf("%s", out.Description)
	}
	return fmt.Sprintf("%d", out.Result.ID), nil
}

func withinSchedule(now time.Time, start, stop string) bool {
	startMin, okA := parseHHMM(start)
	stopMin, okB := parseHHMM(stop)
	if !okA || !okB || startMin == stopMin {
		return false
	}
	cur := now.Hour()*60 + now.Minute()
	if startMin < stopMin {
		return cur >= startMin && cur < stopMin
	}
	return cur >= startMin || cur < stopMin
}

func parseHHMM(s string) (int, bool) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, false
	}
	var h, m int
	if _, err := fmt.Sscanf(parts[0], "%d", &h); err != nil {
		return 0, false
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &m); err != nil {
		return 0, false
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}
