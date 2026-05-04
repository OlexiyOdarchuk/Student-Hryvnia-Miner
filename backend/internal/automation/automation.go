package automation

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OlexiyOdarchuk/Student-Hryvnia-Miner/backend/config"
	"github.com/OlexiyOdarchuk/Student-Hryvnia-Miner/backend/types"

	tele "gopkg.in/telebot.v3"
)

type DashboardProvider interface {
	GetDashboardData() types.DashboardData
}

type Engine struct {
	ctrl           Controller
	dash           DashboardProvider
	bot            *Bot
	botMu          sync.Mutex
	sessionStart   atomic.Int64
	lastMilestone  atomic.Uint32
	notifiedTarget atomic.Bool
	scheduleInit   atomic.Bool
	lastScheduleOn atomic.Bool
}

func New(ctrl Controller, dash DashboardProvider) *Engine {
	e := &Engine{
		ctrl: ctrl,
		dash: dash,
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
	bot, err := StartBot(ctx, auto.TelegramBotToken, auto.TelegramChatID, e.ctrl)
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

func (e *Engine) activeBot() *Bot {
	e.botMu.Lock()
	defer e.botMu.Unlock()
	return e.bot
}

func (e *Engine) ResetSession() {
	e.sessionStart.Store(time.Now().Unix())
	e.notifiedTarget.Store(false)
	e.lastMilestone.Store(0)
}

func (e *Engine) checkRules() {
	auto := config.Config.Automation
	data := e.dash.GetDashboardData()
	mining := e.ctrl.IsMining()

	if auto.ProgressNotifyStep > 0 {
		step := auto.ProgressNotifyStep
		current := data.SessionBlocks / step
		if current > e.lastMilestone.Load() {
			e.lastMilestone.Store(current)
			reached := uint64(current) * uint64(step)
			e.sendTelegram("🔔 Зараховано " + strconv.FormatUint(reached, 10) + " блоків у сесії")
		}
	}

	if auto.BlockTarget > 0 && mining && data.SessionBlocks >= auto.BlockTarget {
		if !e.notifiedTarget.Swap(true) {
			msg := "🎯 Досягнуто цілі: зараховано " + strconv.FormatUint(uint64(data.SessionBlocks), 10) + " блоків. Зупиняю майнінг."
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
			msg := "⏱️ Таймер сесії " + strconv.FormatUint(uint64(auto.SessionMinutes), 10) + " хв вичерпано. Зупиняю майнінг."
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

func (e *Engine) notifyIfEnabled(enabled func(config.AutomationConfig) bool, message string) {
	if enabled(config.Config.Automation) {
		e.sendTelegram(message)
	}
}

func (e *Engine) SendTestMessage() error {
	auto := config.Config.Automation
	if auto.TelegramBotToken == "" || auto.TelegramChatID == "" {
		return errors.New("bot token and chat ID are required")
	}
	return sendOneShot(auto.TelegramBotToken, auto.TelegramChatID, "✅ Тестове повідомлення від S-UAH Miner")
}

func (e *Engine) SendTestWithConfig(token, chatID string) error {
	if token == "" || chatID == "" {
		return errors.New("token and chat_id required")
	}
	return sendOneShot(token, chatID, "✅ Перевірка з'єднання з S-UAH Miner")
}

func (e *Engine) sendTelegram(message string) {
	bot := e.activeBot()
	if bot == nil {
		return
	}
	if err := bot.Send(message); err != nil {
		slog.Debug("Telegram send failed", "err", err)
	}
}

func sendOneShot(token, chatID, message string) error {
	id, err := strconv.ParseInt(strings.TrimSpace(chatID), 10, 64)
	if err != nil {
		return errors.New("invalid chat id: " + err.Error())
	}
	tb, err := tele.NewBot(tele.Settings{Token: token})
	if err != nil {
		return err
	}
	_, err = tb.Send(&tele.User{ID: id}, message)
	return err
}

func (e *Engine) ResolveChatID(token, chatID string) (string, error) {
	if token == "" || chatID == "" {
		return "", errors.New("token and chat_id required")
	}
	tb, err := tele.NewBot(tele.Settings{Token: token})
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(chatID)
	if id, err := strconv.ParseInt(name, 10, 64); err == nil {
		chat, err := tb.ChatByID(id)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(chat.ID, 10), nil
	}
	chat, err := tb.ChatByUsername(name)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(chat.ID, 10), nil
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
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, false
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}
