package bot

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/supercharged09/crypto-rates/internal/model"
	"github.com/supercharged09/crypto-rates/internal/repository"
	"github.com/supercharged09/crypto-rates/internal/service"
)

// CryptoBot - бот для отображения курсов
type CryptoBot struct {
	api              *tgbotapi.BotAPI
	rateSvc          *service.RateService
	subRepo          *repository.SubscriptionRepository
	analyticsService *service.AnalyticsService
	chartService     *service.ChartService
}

// NewCryptoBot - создание нового бота
func NewCryptoBot(
	token string,
	rateSvc *service.RateService,
	subRepo *repository.SubscriptionRepository,
	analyticsService *service.AnalyticsService,
	chartService *service.ChartService,
) (*CryptoBot, error) {
	// Создаём HTTP клиент с кастомным DNS (Google DNS)
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
				Resolver: &net.Resolver{
					PreferGo: true,
					Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
						d := net.Dialer{Timeout: 5 * time.Second}
						// Используем Google DNS (8.8.8.8) и Cloudflare DNS (1.1.1.1)
						return d.DialContext(ctx, network, "8.8.8.8:53")
					},
				},
			}).DialContext,
			TLSHandshakeTimeout: 15 * time.Second,
		},
	}

	api, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = false

	return &CryptoBot{
		api:              api,
		rateSvc:          rateSvc,
		subRepo:          subRepo,
		analyticsService: analyticsService,
		chartService:     chartService,
	}, nil
}

// Start запуск бота в режиме long polling
func (b *CryptoBot) Start(ctx context.Context) {
	log.Printf("Telegram bot started as @%s", b.api.Self.UserName)

	//конфиг получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//получение канала с обновлениями
	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			//обработка callback от inline нопок
			if update.CallbackQuery != nil {
				b.handleCallback(update.CallbackQuery)
				continue
			}

			//обработка сообщений
			if update.Message != nil {
				b.handleMessage(update.Message)
			}

		case <-ctx.Done():
			log.Println("Bot stopped")
			b.api.StopReceivingUpdates()
			return
		}
	}
}

// handleCallback обрабатывает нажатия на инлайн кнопки
func (b *CryptoBot) handleCallback(callback *tgbotapi.CallbackQuery) {
	//отслеживание активности
	if callback.Message != nil {
		b.analyticsService.TrackCallbackActivity(callback)
	}
	//убираем "часики" на кнопке
	ack := tgbotapi.NewCallback(callback.ID, "")
	if _, err := b.api.Request(ack); err != nil {
		log.Printf("ERROR: failed to ack callback: %v", err)
	}

	data := callback.Data
	chatID := callback.Message.Chat.ID
	switch data {
	case "rates_all":
		text, err := b.getAllRatesText()
		if err != nil {
			b.reply(chatID, "❌ Ошибка получения курсов")
			return
		}
		b.replyWithInline(chatID, text, getRatesKeyboard())

	case "rate_btc", "rate_eth", "rate_usdt", "rate_bnb", "rate_xrp", "rate_sol":
		// Извлекаем символ из callback: rate_btc → btc
		symbol := strings.TrimPrefix(data, "rate_")
		info := model.GetCryptoInfo(symbol)
		if info == nil {
			b.reply(chatID, "❌ Валюта не найдена")
			return
		}

		text, err := b.getRateText(info.ID)
		if err != nil {
			b.reply(chatID, fmt.Sprintf("❌ Ошибка: %v", err))
			return
		}
		b.replyWithInline(chatID, text, getRatesKeyboard())

	case "auto_10":
		sub := model.Subscription{
			ChatID:         chatID,
			Cryptocurrency: "all",
			IntervalMin:    10,
			IsActive:       true,
		}
		if err := b.subRepo.Upsert(context.Background(), sub); err != nil {
			b.reply(chatID, "❌ Ошибка сохранения подписки")
			return
		}
		b.reply(chatID, "✅ Авто-рассылка каждые 10 мин.\n/stop_auto для отключения")

	case "auto_30":
		sub := model.Subscription{
			ChatID:         chatID,
			Cryptocurrency: "all",
			IntervalMin:    30,
			IsActive:       true,
		}
		if err := b.subRepo.Upsert(context.Background(), sub); err != nil {
			b.reply(chatID, "❌ Ошибка сохранения подписки")
			return
		}
		b.reply(chatID, "✅ Авто-рассылка каждые 30 мин.\n/stop_auto для отключения")

	case "auto_stop":
		if err := b.subRepo.Deactivate(context.Background(), chatID, "all"); err != nil {
			b.reply(chatID, "❌ Ошибка отключения")
			return
		}
		b.reply(chatID, "✅ Авто-рассылка отключена")

	case "chart_btc":
		b.sendChartHTML(chatID, "bitcoin", "Bitcoin")
	case "chart_eth":
		b.sendChartHTML(chatID, "ethereum", "Ethereum")
	case "chart_usdt":
		b.sendChartHTML(chatID, "tether", "Tether")
	case "chart_bnb":
		b.sendChartHTML(chatID, "binancecoin", "BNB")
	case "chart_xrp":
		b.sendChartHTML(chatID, "ripple", "XRP")
	case "chart_sol":
		b.sendChartHTML(chatID, "solana", "Solana")
	case "start":
		b.cmdStart(callback.Message)
	}
}

// handleMessage - обработка входящих сообщений
func (b *CryptoBot) handleMessage(msg *tgbotapi.Message) {
	//отслеживание активности пользователя
	b.analyticsService.TrackActivity(msg)
	//обработка только команд, начинающихся с /
	if !msg.IsCommand() {
		return //игнорирование обычного текста
	}

	log.Printf("Command: %s from chat %d (args: %q)", msg.Command(), msg.Chat.ID, msg.CommandArguments())

	switch {
	case msg.Command() == "start":
		b.cmdStart(msg)
	case msg.Command() == "rates", strings.HasPrefix(msg.Command(), "rates_"):
		b.cmdRates(msg)
	case msg.Command() == "start_auto":
		b.cmdStartAuto(msg)
	case msg.Command() == "stop_auto":
		b.cmdStopAuto(msg)
	default:
		b.replyWithInline(msg.Chat.ID, "Неизвестная команда", getMainKeyboard())
	}
}

// cmdStart - обработчик /start
func (b *CryptoBot) cmdStart(msg *tgbotapi.Message) {
	text := fmt.Sprintf(
		"👋 Привет! Я бот для отслеживания курсов криптовалют.\n\n" +
			"⚡ *Команды:*\n" +
			"/rates — все курсы\n" +
			"/rates\\_btc — курс BTC\n" + // ← \_ вместо _ т.к. в markdown режиме телега использует символ _ для курсива
			"/rates\\_eth — курс ETH\n" +
			"/rates\\_usdt — курс USDT\n" +
			"/rates\\_bnb — курс BNB\n" +
			"/rates\\_xrp — курс XRP\n" +
			"/rates\\_sol — курс SOL\n" +
			"/rates\\_dogecoin — курс Dogecoin\n" +
			"/rates\\_cardano — курс Cardano\n" +
			"и любая другая монета с CoinGecko\n" +
			"/start\\_auto 10 — авто-рассылка\n" +
			"/stop\\_auto — отключить рассылку\n\n" +
			"Для монет с дефисом используйте:\n" +
			"/rates shiba-inu\n" +
			"/rates bitcoin-cash\n\n" +
			"Или используй кнопки 👇",
	)
	b.replyWithInline(msg.Chat.ID, text, getMainKeyboard())
}

// cmdRates — обработчик /rates и /rates_btc, /rates_eth и т.д.
func (b *CryptoBot) cmdRates(msg *tgbotapi.Message) {
	command := msg.Command()

	// Если команда вида /rates_btc — извлекаем символ из команды
	var cryptoQuery string
	if strings.HasPrefix(command, "rates_") {
		cryptoQuery = strings.TrimPrefix(command, "rates_")
	} else {
		// Обычная /rates — смотрим аргументы после команды
		cryptoQuery = strings.TrimSpace(msg.CommandArguments())
	}

	// Если нет конкретной валюты — показываем все курсы
	if cryptoQuery == "" {
		text, err := b.getAllRatesText()
		if err != nil {
			b.reply(msg.Chat.ID, "❌ Ошибка получения курсов")
			return
		}
		b.replyWithInline(msg.Chat.ID, text, getRatesKeyboard())
		return
	}

	// Сначала проверяем в списке поддерживаемых
	info := model.GetCryptoInfo(cryptoQuery)
	if info != nil {
		text, err := b.getRateText(info.ID)
		if err != nil {
			b.reply(msg.Chat.ID, fmt.Sprintf("❌ Ошибка получения курса %s", info.Name))
			return
		}
		b.replyWithInline(msg.Chat.ID, text, getRatesKeyboard())
		return
	}

	// Если нет в списке — пробуем получить с CoinGecko
	text, err := b.getAnyRateText(cryptoQuery)
	if err != nil {
		b.reply(msg.Chat.ID, fmt.Sprintf("❌ Не удалось получить курс %s\nПроверьте правильность ID на coingecko.com", cryptoQuery))
		return
	}
	b.replyWithInline(msg.Chat.ID, text, getMainKeyboard())
}

// getAnyRateText получает текст для произвольной монеты
func (b *CryptoBot) getAnyRateText(cryptoID string) (string, error) {
	stats, err := b.rateSvc.GetAnyRate(context.Background(), cryptoID)
	if err != nil {
		return "", err
	}

	// Форматируем вручную, так как монета не в SupportedCryptos
	return fmt.Sprintf(
		"🪙 *%s*\n"+
			"💵 Цена: $%s\n"+
			"📉 Мин за 24ч: $%.2f\n"+
			"📈 Макс за 24ч: $%.2f\n"+
			"🕐 Обновлено: %s",
		cryptoID,
		formatPrice(stats.CurrentPrice),
		stats.MinPrice24h,
		stats.MaxPrice24h,
		stats.LastUpdated.Format("15:04:05"),
	), nil
}

// cmdStartAuto - обработчик /start-auto {minutes}
func (b *CryptoBot) cmdStartAuto(msg *tgbotapi.Message) {
	args := strings.TrimSpace(msg.CommandArguments())
	if args == "" {
		b.reply(msg.Chat.ID, "Укажите интервал в минутах.\nПример: /start_auto 10")
		return
	}

	minutes, err := strconv.Atoi(args)
	if err != nil || minutes < 1 || minutes > 1440 {
		b.reply(msg.Chat.ID, "Укажите число от 1 до 1440 (минут)")
		return
	}

	sub := model.Subscription{
		ChatID:         msg.Chat.ID,
		Cryptocurrency: "all",
		IntervalMin:    minutes,
		IsActive:       true,
	}

	if err := b.subRepo.Upsert(context.Background(), sub); err != nil {
		log.Printf("ERROR: failed to save subscription: %v", err)
		b.reply(msg.Chat.ID, "Ошибка сохранения подписки")
		return
	}

	b.reply(msg.Chat.ID, fmt.Sprintf(
		"Авто-рассылка включена!\nИнтервал: каждые %d мин.\n\nДля отключения: /stop_auto",
		minutes,
	))
}

// cmdStopAuto обработчик stop-auto
func (b *CryptoBot) cmdStopAuto(msg *tgbotapi.Message) {
	if err := b.subRepo.Deactivate(context.Background(), msg.Chat.ID, "all"); err != nil {
		log.Printf("ERROR: failed to deactivate subscription: %v", err)
		b.reply(msg.Chat.ID, "Ошибка отключения подписки")
		return
	}

	b.reply(msg.Chat.ID, "Авто рассылка отключена")
}

// reply - вспомогательный метод отправки простого сообщения
func (b *CryptoBot) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("ERROR: failed to send message: %v", err)
	}
}

// replyWithKeyboard - отправка сообщения с клавиатуры
func (b *CryptoBot) replyWithKeyboard(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("ERROR: failed to send message: %v", err)
	}
}

// getAllRatesText - формирование текста со всеми курсами
func (b *CryptoBot) getAllRatesText() (string, error) {
	var parts []string

	for _, crypto := range model.SupportedCryptos {
		stats, err := b.rateSvc.GetRateStats(context.Background(), crypto.ID)
		if err != nil {
			log.Printf("WARNING: failed to get %s stats: %v", crypto.ID, err)
			continue
		}
		parts = append(parts, formatRateMessage(stats))
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("no rates available")
	}

	return strings.Join(parts, "\n\n"), nil
}

// getRateText формирование текста для одной валюты
func (b *CryptoBot) getRateText(query string) (string, error) {
	info := model.GetCryptoInfo(query)
	if info == nil {
		return "", fmt.Errorf("неподдерживаемая валюта: %s", query)
	}

	stats, err := b.rateSvc.GetRateStats(context.Background(), info.ID)
	if err != nil {
		return "", err
	}
	return formatRateMessage(stats), nil
}

// SendRateToChat отправит курс в конкретный чат. Для авто рассылки
func (b *CryptoBot) SendRateToChat(chatID int64) {
	text, err := b.getAllRatesText()
	if err != nil {
		log.Printf("ERROR: failed to get rates for chat %d: %v", chatID, err)
		return
	}
	b.reply(chatID, text)
}

// formatRateMessage форматирует статистику в читаемый текст
func formatRateMessage(stats *model.RateStats) string {
	info := model.GetCryptoInfo(stats.Cryptocurrency)
	if info == nil {
		//fallback для данных без информации, но этого не должно случиться
		return fmt.Sprintf("*%s*\nЦена: $%s", stats.Cryptocurrency, formatPrice(stats.CurrentPrice))
	}

	changeSign := "+"
	if stats.ChangePercent1h < 0 {
		changeSign = ""
	}

	return fmt.Sprintf(
		"%s *%s*\n"+
			"💵 Цена: $%s\n"+
			"📉 Минимально за 24ч: $%.2f\n"+
			"📈 Максимально за 24ч: $%.2f\n"+
			"🕐 Измениние за час: %s%.2f%%",
		info.Emoji,
		info.Name,
		formatPrice(stats.CurrentPrice),
		stats.MinPrice24h,
		stats.MaxPrice24h,
		changeSign,
		stats.ChangePercent1h,
	)
}

// formatPrice форматирует цену с разделителями тысяч
func formatPrice(price float64) string {
	parts := strings.Split(fmt.Sprintf("%.2f", price), ".")
	intPart := parts[0]
	decPart := parts[1]

	//добавляем запятые для тысяч
	n := len(intPart)
	if n > 3 {
		var result strings.Builder
		for i, c := range intPart {
			if i > 0 && (n-i)%3 == 0 {
				result.WriteRune(',')
			}
			result.WriteRune(c)
		}
		intPart = result.String()
	}

	return intPart + "." + decPart
}

// replyWithInline — отправка сообщения с inline-клавиатурой
func (b *CryptoBot) replyWithInline(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("ERROR: failed to send message: %v", err)
	}
}

// getMainKeyboard — главная клавиатура
func getMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Все курсы", "rates_all"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟠 BTC", "rate_btc"),
			tgbotapi.NewInlineKeyboardButtonData("🔷 ETH", "rate_eth"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟢 USDT", "rate_usdt"),
			tgbotapi.NewInlineKeyboardButtonData("🟡 BNB", "rate_bnb"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔵 XRP", "rate_xrp"),
			tgbotapi.NewInlineKeyboardButtonData("🟣 SOL", "rate_sol"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 График BTC", "chart_btc"),
			tgbotapi.NewInlineKeyboardButtonData("📈 График ETH", "chart_eth"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏰ Авто 10 мин", "auto_10"),
			tgbotapi.NewInlineKeyboardButtonData("⏰ Авто 30 мин", "auto_30"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛑 Стоп", "auto_stop"),
		),
	)
}

// getRatesKeyboard — клавиатура после показа курсов
func getRatesKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Обновить", "rates_all"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟠 BTC", "rate_btc"),
			tgbotapi.NewInlineKeyboardButtonData("🔷 ETH", "rate_eth"),
			tgbotapi.NewInlineKeyboardButtonData("🟢 USDT", "rate_usdt"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟡 BNB", "rate_bnb"),
			tgbotapi.NewInlineKeyboardButtonData("🔵 XRP", "rate_xrp"),
			tgbotapi.NewInlineKeyboardButtonData("🟣 SOL", "rate_sol"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 График BTC", "chart_btc"),
			tgbotapi.NewInlineKeyboardButtonData("📈 График ETH", "chart_eth"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 График SOL", "chart_sol"),
			tgbotapi.NewInlineKeyboardButtonData("📈 График XRP", "chart_xrp"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏰ Авто 10 мин", "auto_10"),
			tgbotapi.NewInlineKeyboardButtonData("🛑 Стоп", "auto_stop"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главная", "start"),
		),
	)
}

// sendChartHTML отправит график как TML файл
func (b *CryptoBot) sendChartHTML(chatID int64, cryptoID string, cryptoName string) {
	html, err := b.chartService.GenerateChartHTML(context.Background(), cryptoID)
	if err != nil {
		log.Printf("ERROR: failed to generate chart: %v", err)
		b.reply(chatID, fmt.Sprintf("Ошибка генерации графика для %s", cryptoName))
		return
	}

	//создание временного файла
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("chart_%s_*.html", cryptoID))
	if err != nil {
		log.Printf("ERROR: failed to create temp file: %v", err)
		b.reply(chatID, "Ошибка создания файла")
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(html)); err != nil {
		log.Printf("ERROR: failed to write temp file: %v", err)
		b.reply(chatID, "Ошибка записи файла")
		return
	}
	tmpFile.Close()

	//отправляется как документ
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(tmpFile.Name()))
	doc.Caption = fmt.Sprintf("📈 %s / USD — 24 часа", cryptoName)
	if _, err := b.api.Send(doc); err != nil {
		log.Printf("ERROR: failed to send chart: %v", err)
		b.reply(chatID, "Ошибка отправки графика")
	}
}
