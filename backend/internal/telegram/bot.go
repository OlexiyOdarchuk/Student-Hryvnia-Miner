package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"shminer/backend/types"
	"strconv"
	"strings"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
)

type Controller interface {
	GetDashboardData() types.DashboardData
	SetMining(state bool)
	IsMining() bool
	ToggleWallet(addr string) bool
	AddWallet(name, addr, priv string) error
	ImportWalletJSON(jsonContent string) error
	GenerateWallet(name string) (string, error)
	ResetSession()
}

type convState int

const (
	convIdle convState = iota
	convAwaitNewName
	convAwaitImportJSON
	convAwaitWatchInput
)

type conversation struct {
	state convState
}

type Bot struct {
	ctrl    Controller
	bot     *tele.Bot
	chatID  int64
	stopped chan struct{}
	once    sync.Once

	convMu sync.Mutex
	conv   map[int64]*conversation

	menu *tele.ReplyMarkup
}

const (
	btnStatusText   = "📊 Статус"
	btnWalletsText  = "💳 Гаманці"
	btnToggleText   = "▶️ / ⏸ Майнінг"
	btnAddText      = "➕ Гаманець"
	btnResetText    = "🔄 Reset"
	btnCancelText   = "✖️ Скасувати"
	btnNewWallet    = "🆕 Створити новий"
	btnImportJSON   = "📥 Імпортувати JSON"
	btnWatchWallet  = "👁 Спостережний"
)

func Start(ctx context.Context, token, chatID string, ctrl Controller) (*Bot, error) {
	if token == "" || chatID == "" {
		return nil, fmt.Errorf("token and chatID required")
	}
	id, err := strconv.ParseInt(strings.TrimSpace(chatID), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chat id: %w", err)
	}

	tb, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 5 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	b := &Bot{
		ctrl:    ctrl,
		bot:     tb,
		chatID:  id,
		stopped: make(chan struct{}),
		conv:    make(map[int64]*conversation),
	}
	b.buildMainMenu()
	b.registerHandlers()

	go func() {
		tb.Start()
		close(b.stopped)
	}()

	go func() {
		<-ctx.Done()
		b.Stop()
	}()

	slog.Info("🤖 Telegram bot started")
	return b, nil
}

func (b *Bot) Stop() {
	b.once.Do(func() {
		b.bot.Stop()
	})
}

func (b *Bot) buildMainMenu() {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true, IsPersistent: true}
	row1 := menu.Row(menu.Text(btnStatusText), menu.Text(btnWalletsText))
	row2 := menu.Row(menu.Text(btnToggleText), menu.Text(btnAddText))
	row3 := menu.Row(menu.Text(btnResetText))
	menu.Reply(row1, row2, row3)
	b.menu = menu
}

func (b *Bot) cancelMarkup() *tele.ReplyMarkup {
	mk := &tele.ReplyMarkup{ResizeKeyboard: true, IsPersistent: true}
	mk.Reply(mk.Row(mk.Text(btnCancelText)))
	return mk
}

func (b *Bot) addMenuMarkup() *tele.ReplyMarkup {
	mk := &tele.ReplyMarkup{ResizeKeyboard: true, IsPersistent: true}
	mk.Reply(
		mk.Row(mk.Text(btnNewWallet)),
		mk.Row(mk.Text(btnImportJSON)),
		mk.Row(mk.Text(btnWatchWallet)),
		mk.Row(mk.Text(btnCancelText)),
	)
	return mk
}

func (b *Bot) registerHandlers() {
	b.bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if !b.authorized(c) {
				return b.reject(c)
			}
			return next(c)
		}
	})

	b.bot.Handle("/start", b.handleHelp)
	b.bot.Handle("/help", b.handleHelp)
	b.bot.Handle("/status", b.handleStatus)
	b.bot.Handle("/wallets", b.handleWallets)
	b.bot.Handle("/mine", b.handleStart)
	b.bot.Handle("/pause", b.handlePause)
	b.bot.Handle("/addwallet", b.handleAddMenu)
	b.bot.Handle("/reset", b.handleReset)

	b.bot.Handle(tele.OnText, b.handleText)
	b.bot.Handle(tele.OnCallback, b.handleCallback)
}

func (b *Bot) authorized(c tele.Context) bool {
	return c.Sender() != nil && c.Sender().ID == b.chatID
}

func (b *Bot) reject(c tele.Context) error {
	return c.Send("⛔ Доступ заборонено. Ваш ID не відповідає налаштованому Chat ID.")
}

func (b *Bot) getConv(userID int64) *conversation {
	b.convMu.Lock()
	defer b.convMu.Unlock()
	conv, ok := b.conv[userID]
	if !ok {
		conv = &conversation{state: convIdle}
		b.conv[userID] = conv
	}
	return conv
}

func (b *Bot) resetConv(userID int64) {
	b.convMu.Lock()
	defer b.convMu.Unlock()
	if conv, ok := b.conv[userID]; ok {
		conv.state = convIdle
	}
}

func (b *Bot) handleHelp(c tele.Context) error {
	msg := `🤖 *S-UAH Miner Bot*

Використовуйте клавіатуру нижче або команди:
/status — поточна статистика
/wallets — список гаманців з кнопками керування
/mine — продовжити майнінг
/pause — призупинити майнінг
/addwallet — додати гаманець (новий, імпорт JSON або спостережний)
/reset — скинути сесійні лічильники

Ви отримуватимете автоматичні сповіщення про події майнінгу, якщо увімкнено в налаштуваннях.`
	return c.Send(msg, tele.ModeMarkdown, b.menu)
}

func (b *Bot) handleStatus(c tele.Context) error {
	b.resetConv(c.Sender().ID)
	data := b.ctrl.GetDashboardData()
	mining := b.ctrl.IsMining()
	active := 0
	for _, w := range data.Wallets {
		if w.Working {
			active++
		}
	}
	state := "🟢 активний"
	if !mining {
		state = "⏸ призупинено"
	}
	msg := fmt.Sprintf(
		`📊 *Статистика*

Стан: %s
Хешрейт: %.2f MH/s
Баланс: %.2f S-UAH
Аптайм: %s

Зараховано: %d  (%.2f / хв)
Намайнено: %d  (%.2f / хв)
У черзі: %d
Всього: %d

Гаманці: %d активні / %d всього`,
		state,
		data.Hashrate,
		data.TotalBalance,
		data.Uptime,
		data.SessionBlocks, data.BlocksPerMin,
		data.SessionFound, data.FoundPerMin,
		data.SubmitQueueLen,
		data.LifetimeBlocks,
		active, len(data.Wallets),
	)
	return c.Send(msg, tele.ModeMarkdown, b.menu)
}

func (b *Bot) handleWallets(c tele.Context) error {
	b.resetConv(c.Sender().ID)
	data := b.ctrl.GetDashboardData()
	if len(data.Wallets) == 0 {
		return c.Send("Немає гаманців. Використайте ➕ Гаманець щоб додати.", b.menu)
	}

	markup := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0, len(data.Wallets))
	for _, w := range data.Wallets {
		icon := "⏸"
		if w.Working {
			icon = "🟢"
		}
		label := fmt.Sprintf("%s %s — %.2f", icon, w.Name, w.ServerBalance)
		btn := markup.Data(label, "tw", w.Address)
		rows = append(rows, markup.Row(btn))
	}
	markup.Inline(rows...)
	return c.Send("Гаманці (клік — увімкнути/вимкнути):", markup)
}

func (b *Bot) handleStart(c tele.Context) error {
	b.resetConv(c.Sender().ID)
	b.ctrl.SetMining(true)
	return c.Send("▶ Майнінг продовжено.", b.menu)
}

func (b *Bot) handlePause(c tele.Context) error {
	b.resetConv(c.Sender().ID)
	b.ctrl.SetMining(false)
	return c.Send("⏸ Майнінг призупинено.", b.menu)
}

func (b *Bot) handleToggleMining(c tele.Context) error {
	b.resetConv(c.Sender().ID)
	if b.ctrl.IsMining() {
		b.ctrl.SetMining(false)
		return c.Send("⏸ Майнінг призупинено.", b.menu)
	}
	b.ctrl.SetMining(true)
	return c.Send("▶ Майнінг запущено.", b.menu)
}

func (b *Bot) handleReset(c tele.Context) error {
	b.resetConv(c.Sender().ID)
	b.ctrl.ResetSession()
	return c.Send("🔄 Сесійні лічильники скинуто.", b.menu)
}

func (b *Bot) handleAddMenu(c tele.Context) error {
	b.resetConv(c.Sender().ID)
	return c.Send("Оберіть спосіб додавання гаманця:", b.addMenuMarkup())
}

func (b *Bot) handleText(c tele.Context) error {
	text := strings.TrimSpace(c.Text())

	switch text {
	case btnStatusText:
		return b.handleStatus(c)
	case btnWalletsText:
		return b.handleWallets(c)
	case btnToggleText:
		return b.handleToggleMining(c)
	case btnResetText:
		return b.handleReset(c)
	case btnAddText:
		return b.handleAddMenu(c)
	case btnCancelText:
		b.resetConv(c.Sender().ID)
		return c.Send("Скасовано.", b.menu)
	case btnNewWallet:
		b.getConv(c.Sender().ID).state = convAwaitNewName
		return c.Send("Введіть ім'я нового гаманця:", b.cancelMarkup())
	case btnImportJSON:
		b.getConv(c.Sender().ID).state = convAwaitImportJSON
		return c.Send(
			"Надішліть JSON гаманця у форматі, як експортує програма, напр.:\n"+
				"`{\"name\":\"...\",\"pub\":\"...\",\"priv\":\"...\"}`",
			tele.ModeMarkdown, b.cancelMarkup())
	case btnWatchWallet:
		b.getConv(c.Sender().ID).state = convAwaitWatchInput
		return c.Send(
			"Надішліть ім'я та адресу одним повідомленням: `<ім'я> <адреса>`",
			tele.ModeMarkdown, b.cancelMarkup())
	}

	conv := b.getConv(c.Sender().ID)
	switch conv.state {
	case convAwaitNewName:
		name := text
		if name == "" {
			return c.Send("Ім'я не може бути порожнім. Спробуйте ще раз:", b.cancelMarkup())
		}
		addr, err := b.ctrl.GenerateWallet(name)
		b.resetConv(c.Sender().ID)
		if err != nil {
			return c.Send("❌ "+err.Error(), b.menu)
		}
		return c.Send(fmt.Sprintf("✅ Новий гаманець `%s` створено.\nАдреса:\n`%s`", name, addr),
			tele.ModeMarkdown, b.menu)

	case convAwaitImportJSON:
		jsonContent := strings.TrimSpace(c.Text())
		err := b.ctrl.ImportWalletJSON(jsonContent)
		b.resetConv(c.Sender().ID)
		if err != nil {
			return c.Send("❌ "+err.Error(), b.menu)
		}
		return c.Send("✅ Гаманець імпортовано з JSON.", b.menu)

	case convAwaitWatchInput:
		parts := strings.Fields(text)
		if len(parts) < 2 {
			return c.Send("Формат: `<ім'я> <адреса>`. Спробуйте ще раз:",
				tele.ModeMarkdown, b.cancelMarkup())
		}
		name := parts[0]
		addr := parts[1]
		err := b.ctrl.AddWallet(name, addr, "")
		b.resetConv(c.Sender().ID)
		if err != nil {
			return c.Send("❌ "+err.Error(), b.menu)
		}
		return c.Send("✅ Спостережний гаманець додано (без приватного ключа).", b.menu)
	}

	return nil
}

func (b *Bot) handleCallback(c tele.Context) error {
	cb := c.Callback()
	if cb == nil {
		return nil
	}
	raw := strings.TrimPrefix(strings.TrimSpace(cb.Data), "\f")
	parts := strings.SplitN(raw, "|", 2)
	if len(parts) != 2 {
		return c.Respond(&tele.CallbackResponse{Text: "невідома дія"})
	}
	action, payload := parts[0], parts[1]
	switch action {
	case "tw":
		working := b.ctrl.ToggleWallet(payload)
		state := "⏸ вимкнено"
		if working {
			state = "🟢 увімкнено"
		}
		if err := b.refreshWalletsMarkup(c); err != nil {
			slog.Debug("refresh wallets markup failed", "err", err)
		}
		short := payload
		if len(payload) > 10 {
			short = payload[:10] + "…"
		}
		return c.Respond(&tele.CallbackResponse{Text: state + " " + short})
	}
	return c.Respond(&tele.CallbackResponse{Text: "невідома дія"})
}

func (b *Bot) refreshWalletsMarkup(c tele.Context) error {
	data := b.ctrl.GetDashboardData()
	if len(data.Wallets) == 0 {
		return nil
	}
	markup := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0, len(data.Wallets))
	for _, w := range data.Wallets {
		icon := "⏸"
		if w.Working {
			icon = "🟢"
		}
		label := fmt.Sprintf("%s %s — %.2f", icon, w.Name, w.ServerBalance)
		btn := markup.Data(label, "tw", w.Address)
		rows = append(rows, markup.Row(btn))
	}
	markup.Inline(rows...)
	return c.Edit("Гаманці (клік — увімкнути/вимкнути):", markup)
}
