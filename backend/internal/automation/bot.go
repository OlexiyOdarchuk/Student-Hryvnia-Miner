package automation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/OlexiyOdarchuk/Student-Hryvnia-Miner/backend/config"
	"github.com/OlexiyOdarchuk/Student-Hryvnia-Miner/backend/types"

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

	RenameWallet(addr, newName string) error
	DeleteWalletBot(addr string) error
	ExportWalletJSONBot(addr string) (string, error)

	GetAutomation() config.AutomationConfig
	SaveAutomationBot(cfg config.AutomationConfig) error
	GetSubmitBufferSize() int
	SaveSubmitBufferSizeBot(size int) error
}

type convState int

const (
	convIdle convState = iota
	convAwaitNewName
	convAwaitImportJSON
	convAwaitWatchInput
	convAwaitRenameName
	convAwaitBlockTarget
	convAwaitSessionMinutes
	convAwaitProgressStep
	convAwaitScheduleStart
	convAwaitScheduleStop
)

type conversation struct {
	state convState
	addr  string
}

type Bot struct {
	ctrl    Controller
	bot     *tele.Bot
	chatID  int64
	stopped chan struct{}
	once    sync.Once

	convMu sync.Mutex
	conv   map[int64]*conversation

	addrMu  sync.Mutex
	addrMap map[string]string

	menu *tele.ReplyMarkup
}

const (
	btnStatusText   = "📊 Статус"
	btnWalletsText  = "💳 Гаманці"
	btnToggleText   = "▶️ / ⏸ Майнінг"
	btnAddText      = "➕ Гаманець"
	btnResetText    = "🔄 Reset"
	btnSettingsText = "⚙️ Налаштування"
	btnCancelText   = "✖️ Скасувати"
	btnNewWallet    = "🆕 Створити новий"
	btnImportJSON   = "📥 Імпортувати JSON"
	btnWatchWallet  = "👁 Спостережний"
)

func StartBot(ctx context.Context, token, chatID string, ctrl Controller) (*Bot, error) {
	if token == "" || chatID == "" {
		return nil, errors.New("token and chatID required")
	}
	id, err := strconv.ParseInt(strings.TrimSpace(chatID), 10, 64)
	if err != nil {
		return nil, errors.New("invalid chat id: " + err.Error())
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
		addrMap: make(map[string]string),
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

func (b *Bot) Send(message string) error {
	if b == nil || b.bot == nil {
		return errors.New("bot not initialized")
	}
	_, err := b.bot.Send(&tele.User{ID: b.chatID}, message)
	return err
}

func (b *Bot) buildMainMenu() {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true, IsPersistent: true}
	row1 := menu.Row(menu.Text(btnStatusText), menu.Text(btnWalletsText))
	row2 := menu.Row(menu.Text(btnToggleText), menu.Text(btnAddText))
	row3 := menu.Row(menu.Text(btnSettingsText), menu.Text(btnResetText))
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
	b.bot.Handle("/settings", b.handleSettings)
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
		conv.addr = ""
	}
}

func shortAddrID(addr string) string {
	sum := sha256.Sum256([]byte(addr))
	return hex.EncodeToString(sum[:6])
}

func (b *Bot) registerAddr(addr string) string {
	id := shortAddrID(addr)
	b.addrMu.Lock()
	b.addrMap[id] = addr
	b.addrMu.Unlock()
	return id
}

func (b *Bot) resolveAddr(id string) (string, bool) {
	b.addrMu.Lock()
	defer b.addrMu.Unlock()
	addr, ok := b.addrMap[id]
	return addr, ok
}

func (b *Bot) handleHelp(c tele.Context) error {
	msg := "🤖 *S-UAH Miner Bot*\n\n" +
		"Використовуйте клавіатуру нижче або команди:\n" +
		"/status — поточна статистика\n" +
		"/wallets — список гаманців з діями\n" +
		"/mine — продовжити майнінг\n" +
		"/pause — призупинити майнінг\n" +
		"/addwallet — додати гаманець\n" +
		"/settings — налаштування автоматизації\n" +
		"/reset — скинути сесійні лічильники\n\n" +
		"Ви отримуватимете автоматичні сповіщення про події майнінгу, якщо увімкнено в налаштуваннях."
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

	var sb strings.Builder
	sb.WriteString("📊 *Статистика*\n\n")
	sb.WriteString("Стан: ")
	sb.WriteString(state)
	sb.WriteString("\nХешрейт: ")
	sb.WriteString(strconv.FormatFloat(data.Hashrate, 'f', 2, 64))
	sb.WriteString(" MH/s\nБаланс: ")
	sb.WriteString(strconv.FormatFloat(data.TotalBalance, 'f', 2, 64))
	sb.WriteString(" S-UAH\nАптайм: ")
	sb.WriteString(data.Uptime)
	sb.WriteString("\n\nЗараховано: ")
	sb.WriteString(strconv.FormatUint(uint64(data.SessionBlocks), 10))
	sb.WriteString("  (")
	sb.WriteString(strconv.FormatFloat(data.BlocksPerMin, 'f', 2, 64))
	sb.WriteString(" / хв)\nНамайнено: ")
	sb.WriteString(strconv.FormatUint(uint64(data.SessionFound), 10))
	sb.WriteString("  (")
	sb.WriteString(strconv.FormatFloat(data.FoundPerMin, 'f', 2, 64))
	sb.WriteString(" / хв)\nУ черзі: ")
	sb.WriteString(strconv.Itoa(data.SubmitQueueLen))
	sb.WriteString("\nВсього: ")
	sb.WriteString(strconv.FormatUint(uint64(data.LifetimeBlocks), 10))
	sb.WriteString("\n\nГаманці: ")
	sb.WriteString(strconv.Itoa(active))
	sb.WriteString(" активні / ")
	sb.WriteString(strconv.Itoa(len(data.Wallets)))
	sb.WriteString(" всього")

	return c.Send(sb.String(), tele.ModeMarkdown, b.menu)
}

func (b *Bot) handleWallets(c tele.Context) error {
	b.resetConv(c.Sender().ID)
	return c.Send("Гаманці (клік по рядку — увімкнути/вимкнути):", b.walletsMarkup())
}

func (b *Bot) walletsMarkup() *tele.ReplyMarkup {
	data := b.ctrl.GetDashboardData()
	markup := &tele.ReplyMarkup{}
	if len(data.Wallets) == 0 {
		return markup
	}
	rows := make([]tele.Row, 0, len(data.Wallets)*2)
	for _, w := range data.Wallets {
		icon := "⏸"
		if w.Working {
			icon = "🟢"
		}
		label := icon + " " + w.Name + " — " + strconv.FormatFloat(w.ServerBalance, 'f', 2, 64)
		id := b.registerAddr(w.Address)
		toggle := markup.Data(label, "tw", id)
		rename := markup.Data("✏ Rename", "rn", id)
		del := markup.Data("🗑 Delete", "dl", id)
		export := markup.Data("📤 Export", "ex", id)
		rows = append(rows, markup.Row(toggle), markup.Row(rename, del, export))
	}
	markup.Inline(rows...)
	return markup
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

func (b *Bot) handleSettings(c tele.Context) error {
	b.resetConv(c.Sender().ID)
	return c.Send(b.settingsText(), tele.ModeMarkdown, b.settingsMarkup())
}

func (b *Bot) settingsText() string {
	a := b.ctrl.GetAutomation()

	var sb strings.Builder
	sb.WriteString("⚙️ *Налаштування автоматизації*\n\n")
	sb.WriteString("📅 Розклад: ")
	if a.ScheduleEnabled {
		sb.WriteString("увімкнено")
	} else {
		sb.WriteString("вимкнено")
	}
	start := a.ScheduleStart
	if start == "" {
		start = "—"
	}
	stop := a.ScheduleStop
	if stop == "" {
		stop = "—"
	}
	sb.WriteString(" (")
	sb.WriteString(start)
	sb.WriteString(" → ")
	sb.WriteString(stop)
	sb.WriteString(")\n🎯 Ціль блоків: ")
	sb.WriteString(strconv.FormatUint(uint64(a.BlockTarget), 10))
	sb.WriteString("\n⏱ Таймер сесії: ")
	sb.WriteString(strconv.FormatUint(uint64(a.SessionMinutes), 10))
	sb.WriteString(" хв\n🔔 Крок прогресу: ")
	if a.ProgressNotifyStep > 0 {
		sb.WriteString("кожні ")
		sb.WriteString(strconv.FormatUint(uint64(a.ProgressNotifyStep), 10))
		sb.WriteString(" блоків")
	} else {
		sb.WriteString("вимкнено")
	}
	sb.WriteString("\n\n🔔 Сповіщення:")
	sb.WriteString("\n  старт: ")
	sb.WriteString(onOff(a.NotifyOnStart))
	sb.WriteString("\n  стоп: ")
	sb.WriteString(onOff(a.NotifyOnStop))
	sb.WriteString("\n  ціль: ")
	sb.WriteString(onOff(a.NotifyOnTarget))
	return sb.String()
}

func onOff(v bool) string {
	if v {
		return "✅"
	}
	return "❌"
}

func (b *Bot) settingsMarkup() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("🔔 Сповіщення", "s", "notif")),
		markup.Row(markup.Data("📅 Розклад", "s", "sched")),
		markup.Row(
			markup.Data("🎯 Ціль блоків", "s", "bt"),
			markup.Data("⏱ Таймер", "s", "sm"),
		),
		markup.Row(markup.Data("🔔 Крок прогресу", "s", "pn")),
	)
	return markup
}

func (b *Bot) notifMarkup() *tele.ReplyMarkup {
	a := b.ctrl.GetAutomation()
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data(onOff(a.NotifyOnStart)+" Старт майнінгу", "ns", "start")),
		markup.Row(markup.Data(onOff(a.NotifyOnStop)+" Зупинка майнінгу", "ns", "stop")),
		markup.Row(markup.Data(onOff(a.NotifyOnTarget)+" Досягнення цілі", "ns", "target")),
		markup.Row(markup.Data("← Назад", "s", "back")),
	)
	return markup
}

func (b *Bot) schedMarkup() *tele.ReplyMarkup {
	a := b.ctrl.GetAutomation()
	start := a.ScheduleStart
	if start == "" {
		start = "—"
	}
	stop := a.ScheduleStop
	if stop == "" {
		stop = "—"
	}
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data(onOff(a.ScheduleEnabled)+" Увімкнути розклад", "sh", "en")),
		markup.Row(markup.Data("⏰ Старт: "+start, "sh", "st")),
		markup.Row(markup.Data("⏰ Стоп: "+stop, "sh", "sp")),
		markup.Row(markup.Data("← Назад", "s", "back")),
	)
	return markup
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
	case btnSettingsText:
		return b.handleSettings(c)
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
		return c.Send("✅ Новий гаманець `"+name+"` створено.\nАдреса:\n`"+addr+"`",
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

	case convAwaitRenameName:
		addr := conv.addr
		name := text
		b.resetConv(c.Sender().ID)
		if name == "" || addr == "" {
			return c.Send("Нове ім'я не може бути порожнім.", b.menu)
		}
		if err := b.ctrl.RenameWallet(addr, name); err != nil {
			return c.Send("❌ "+err.Error(), b.menu)
		}
		return c.Send("✅ Гаманець перейменовано на `"+name+"`.", tele.ModeMarkdown, b.menu)

	case convAwaitBlockTarget:
		b.resetConv(c.Sender().ID)
		n, err := strconv.ParseUint(text, 10, 32)
		if err != nil {
			return c.Send("❌ Очікується ціле число ≥ 0.", b.menu)
		}
		cfg := b.ctrl.GetAutomation()
		cfg.BlockTarget = uint32(n)
		if err := b.ctrl.SaveAutomationBot(cfg); err != nil {
			return c.Send("❌ "+err.Error(), b.menu)
		}
		return c.Send("✅ Ціль блоків оновлено.\n\n"+b.settingsText(), tele.ModeMarkdown, b.settingsMarkup())

	case convAwaitSessionMinutes:
		b.resetConv(c.Sender().ID)
		n, err := strconv.ParseUint(text, 10, 32)
		if err != nil {
			return c.Send("❌ Очікується ціле число ≥ 0.", b.menu)
		}
		cfg := b.ctrl.GetAutomation()
		cfg.SessionMinutes = uint32(n)
		if err := b.ctrl.SaveAutomationBot(cfg); err != nil {
			return c.Send("❌ "+err.Error(), b.menu)
		}
		return c.Send("✅ Таймер сесії оновлено.\n\n"+b.settingsText(), tele.ModeMarkdown, b.settingsMarkup())

	case convAwaitProgressStep:
		b.resetConv(c.Sender().ID)
		n, err := strconv.ParseUint(text, 10, 32)
		if err != nil {
			return c.Send("❌ Очікується ціле число ≥ 0.", b.menu)
		}
		cfg := b.ctrl.GetAutomation()
		cfg.ProgressNotifyStep = uint32(n)
		if err := b.ctrl.SaveAutomationBot(cfg); err != nil {
			return c.Send("❌ "+err.Error(), b.menu)
		}
		return c.Send("✅ Крок прогресу оновлено.\n\n"+b.settingsText(), tele.ModeMarkdown, b.settingsMarkup())

	case convAwaitScheduleStart:
		b.resetConv(c.Sender().ID)
		if _, ok := parseHHMM(text); !ok {
			return c.Send("❌ Формат HH:MM (00:00 – 23:59).", b.menu)
		}
		cfg := b.ctrl.GetAutomation()
		cfg.ScheduleStart = text
		if err := b.ctrl.SaveAutomationBot(cfg); err != nil {
			return c.Send("❌ "+err.Error(), b.menu)
		}
		return c.Send("✅ Час старту оновлено.\n\n"+b.settingsText(), tele.ModeMarkdown, b.schedMarkup())

	case convAwaitScheduleStop:
		b.resetConv(c.Sender().ID)
		if _, ok := parseHHMM(text); !ok {
			return c.Send("❌ Формат HH:MM (00:00 – 23:59).", b.menu)
		}
		cfg := b.ctrl.GetAutomation()
		cfg.ScheduleStop = text
		if err := b.ctrl.SaveAutomationBot(cfg); err != nil {
			return c.Send("❌ "+err.Error(), b.menu)
		}
		return c.Send("✅ Час зупинки оновлено.\n\n"+b.settingsText(), tele.ModeMarkdown, b.schedMarkup())
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
		return b.cbToggleWallet(c, payload)
	case "rn":
		return b.cbRenameWallet(c, payload)
	case "dl":
		return b.cbDeleteAsk(c, payload)
	case "dlc":
		return b.cbDeleteConfirm(c, payload)
	case "dlx":
		return b.cbDeleteCancel(c)
	case "ex":
		return b.cbExportWallet(c, payload)
	case "s":
		return b.cbSettings(c, payload)
	case "ns":
		return b.cbNotifToggle(c, payload)
	case "sh":
		return b.cbScheduleAction(c, payload)
	}
	return c.Respond(&tele.CallbackResponse{Text: "невідома дія"})
}

func (b *Bot) cbToggleWallet(c tele.Context, payload string) error {
	addr, ok := b.resolveAddr(payload)
	if !ok {
		return c.Respond(&tele.CallbackResponse{Text: "гаманець не знайдено (оновіть список)"})
	}
	working := b.ctrl.ToggleWallet(addr)
	state := "⏸ вимкнено"
	if working {
		state = "🟢 увімкнено"
	}
	if err := c.Edit("Гаманці (клік по рядку — увімкнути/вимкнути):", b.walletsMarkup()); err != nil {
		slog.Debug("refresh wallets markup failed", "err", err)
	}
	short := addr
	if len(addr) > 10 {
		short = addr[:10] + "…"
	}
	return c.Respond(&tele.CallbackResponse{Text: state + " " + short})
}

func (b *Bot) cbRenameWallet(c tele.Context, payload string) error {
	addr, ok := b.resolveAddr(payload)
	if !ok {
		return c.Respond(&tele.CallbackResponse{Text: "гаманець не знайдено"})
	}
	conv := b.getConv(c.Sender().ID)
	b.convMu.Lock()
	conv.state = convAwaitRenameName
	conv.addr = addr
	b.convMu.Unlock()
	if err := c.Respond(&tele.CallbackResponse{}); err != nil {
		slog.Debug("callback respond failed", "err", err)
	}
	return c.Send("Введіть нове ім'я для гаманця:", b.cancelMarkup())
}

func (b *Bot) cbDeleteAsk(c tele.Context, payload string) error {
	addr, ok := b.resolveAddr(payload)
	if !ok {
		return c.Respond(&tele.CallbackResponse{Text: "гаманець не знайдено"})
	}
	short := addr
	if len(addr) > 16 {
		short = addr[:16] + "…"
	}
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(
		markup.Data("✅ Так, видалити", "dlc", payload),
		markup.Data("✖ Скасувати", "dlx", "_"),
	))
	if err := c.Respond(&tele.CallbackResponse{}); err != nil {
		slog.Debug("callback respond failed", "err", err)
	}
	return c.Send("Видалити гаманець `"+short+"`?", tele.ModeMarkdown, markup)
}

func (b *Bot) cbDeleteConfirm(c tele.Context, payload string) error {
	addr, ok := b.resolveAddr(payload)
	if !ok {
		return c.Respond(&tele.CallbackResponse{Text: "гаманець не знайдено"})
	}
	if err := b.ctrl.DeleteWalletBot(addr); err != nil {
		return c.Respond(&tele.CallbackResponse{Text: err.Error()})
	}
	if err := c.Edit("🗑 Гаманець видалено."); err != nil {
		slog.Debug("edit after delete failed", "err", err)
	}
	return c.Respond(&tele.CallbackResponse{Text: "видалено"})
}

func (b *Bot) cbDeleteCancel(c tele.Context) error {
	if err := c.Edit("Скасовано."); err != nil {
		slog.Debug("edit cancel failed", "err", err)
	}
	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) cbExportWallet(c tele.Context, payload string) error {
	addr, ok := b.resolveAddr(payload)
	if !ok {
		return c.Respond(&tele.CallbackResponse{Text: "гаманець не знайдено"})
	}
	data, err := b.ctrl.ExportWalletJSONBot(addr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: err.Error()})
	}
	if err := c.Respond(&tele.CallbackResponse{}); err != nil {
		slog.Debug("callback respond failed", "err", err)
	}
	return c.Send("📤 Експорт гаманця (зберігайте в надійному місці):\n```\n"+data+"\n```",
		tele.ModeMarkdown)
}

func (b *Bot) cbSettings(c tele.Context, payload string) error {
	switch payload {
	case "notif":
		if err := c.Edit(b.settingsText(), tele.ModeMarkdown, b.notifMarkup()); err != nil {
			slog.Debug("edit notif menu failed", "err", err)
		}
	case "sched":
		if err := c.Edit(b.settingsText(), tele.ModeMarkdown, b.schedMarkup()); err != nil {
			slog.Debug("edit sched menu failed", "err", err)
		}
	case "back":
		if err := c.Edit(b.settingsText(), tele.ModeMarkdown, b.settingsMarkup()); err != nil {
			slog.Debug("edit settings menu failed", "err", err)
		}
	case "bt":
		b.getConv(c.Sender().ID).state = convAwaitBlockTarget
		if err := c.Respond(&tele.CallbackResponse{}); err != nil {
			slog.Debug("callback respond failed", "err", err)
		}
		return c.Send("Введіть ціль блоків (0 — вимкнути правило):", b.cancelMarkup())
	case "sm":
		b.getConv(c.Sender().ID).state = convAwaitSessionMinutes
		if err := c.Respond(&tele.CallbackResponse{}); err != nil {
			slog.Debug("callback respond failed", "err", err)
		}
		return c.Send("Введіть тривалість сесії у хвилинах (0 — без таймера):", b.cancelMarkup())
	case "pn":
		b.getConv(c.Sender().ID).state = convAwaitProgressStep
		if err := c.Respond(&tele.CallbackResponse{}); err != nil {
			slog.Debug("callback respond failed", "err", err)
		}
		return c.Send("Введіть крок прогресу в блоках (0 — вимкнути). Напр. 200 або 1000:", b.cancelMarkup())
	}
	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) cbNotifToggle(c tele.Context, payload string) error {
	cfg := b.ctrl.GetAutomation()
	switch payload {
	case "start":
		cfg.NotifyOnStart = !cfg.NotifyOnStart
	case "stop":
		cfg.NotifyOnStop = !cfg.NotifyOnStop
	case "target":
		cfg.NotifyOnTarget = !cfg.NotifyOnTarget
	default:
		return c.Respond(&tele.CallbackResponse{Text: "невідомо"})
	}
	if err := b.ctrl.SaveAutomationBot(cfg); err != nil {
		return c.Respond(&tele.CallbackResponse{Text: err.Error()})
	}
	if err := c.Edit(b.settingsText(), tele.ModeMarkdown, b.notifMarkup()); err != nil {
		slog.Debug("edit after notif toggle failed", "err", err)
	}
	return c.Respond(&tele.CallbackResponse{})
}

func (b *Bot) cbScheduleAction(c tele.Context, payload string) error {
	switch payload {
	case "en":
		cfg := b.ctrl.GetAutomation()
		cfg.ScheduleEnabled = !cfg.ScheduleEnabled
		if err := b.ctrl.SaveAutomationBot(cfg); err != nil {
			return c.Respond(&tele.CallbackResponse{Text: err.Error()})
		}
		if err := c.Edit(b.settingsText(), tele.ModeMarkdown, b.schedMarkup()); err != nil {
			slog.Debug("edit after schedule toggle failed", "err", err)
		}
	case "st":
		b.getConv(c.Sender().ID).state = convAwaitScheduleStart
		if err := c.Respond(&tele.CallbackResponse{}); err != nil {
			slog.Debug("callback respond failed", "err", err)
		}
		return c.Send("Введіть час старту у форматі HH:MM:", b.cancelMarkup())
	case "sp":
		b.getConv(c.Sender().ID).state = convAwaitScheduleStop
		if err := c.Respond(&tele.CallbackResponse{}); err != nil {
			slog.Debug("callback respond failed", "err", err)
		}
		return c.Send("Введіть час зупинки у форматі HH:MM:", b.cancelMarkup())
	}
	return c.Respond(&tele.CallbackResponse{})
}
