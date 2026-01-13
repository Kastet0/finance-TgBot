package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"finance-tg-bot/internal/models"
)

func (b *Bot) handleListExpenses(ctx context.Context, msg *tgbotapi.Message, user *models.User, page int) {
	if b.db == nil {
		b.sendMessage(msg.Chat.ID, "❌ База данных временно не доступна. Попробуйте позже.")
		return
	}

	limit := 5 // Показываем по 5 трат на странице
	offset := page * limit

	expenses, err := b.db.GetUserExpenses(ctx, user.TelegramID, limit, offset)
	if err != nil {
		b.logger.Error("Ошибка получения трат", "error", err)
		b.sendMessage(msg.Chat.ID, "❌ Ошибка при получении списка трат. Попробуйте позже.")
		return
	}

	if len(expenses) == 0 {
		if page == 0 {
			b.sendMessage(msg.Chat.ID, "📭 *У вас пока нет трат.*\n\nДобавьте первую трату командой: /add")
		} else {
			b.sendMessage(msg.Chat.ID, "📭 *Больше нет трат для отображения.*")
		}
		return
	}

	// Формируем сообщение со списком
	response := "📝 *Ваши траты:*\n\n"
	total := 0.0

	for i, expense := range expenses {
		total += expense.Amount
		dateStr := expense.CreatedAt.Format("02.01 15:04")
		if expense.CreatedAt.IsZero() {
			dateStr = "не указана"
		}

		// Форматируем каждую трату с номером
		response += fmt.Sprintf("%d. 💰 *%.2f руб.*\n   📝 %s\n   🏷 %s\n   📅 %s\n\n",
			offset+i+1,
			expense.Amount,
			expense.Description,
			expense.Category,
			dateStr)
	}

	response += fmt.Sprintf("💰 *Итого на странице: %.2f руб.*\n\n", total)

	// Создаем клавиатуру с кнопками удаления
	var keyboardRows [][]tgbotapi.InlineKeyboardButton

	for i, expense := range expenses {
		// Создаем кнопку удаления для каждой траты
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("🗑️ Удалить трату %d", offset+i+1),
			fmt.Sprintf("delete_%s", expense.ID.String()),
		)
		keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	// Добавляем кнопки навигации если нужно
	if page > 0 || len(expenses) == limit {
		var navButtons []tgbotapi.InlineKeyboardButton

		if page > 0 {
			navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(
				"◀️ Назад",
				fmt.Sprintf("list_%d", page-1),
			))
		}

		if len(expenses) == limit {
			navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(
				"Вперед ▶️",
				fmt.Sprintf("list_%d", page+1),
			))
		}

		keyboardRows = append(keyboardRows, navButtons)
	}

	// Кнопка для отмены
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Закрыть", "close_list"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	// Отправляем сообщение с клавиатурой
	message := tgbotapi.NewMessage(msg.Chat.ID, response)
	message.ParseMode = "Markdown"
	message.ReplyMarkup = keyboard

	if _, err := b.api.Send(message); err != nil {
		b.logger.Error("Ошибка отправки сообщения", "error", err)
	}
}

func (b *Bot) handleTodayStats(ctx context.Context, msg *tgbotapi.Message, user *models.User) {
	if b.db == nil {
		b.sendMessage(msg.Chat.ID, "❌ База данных временно не доступна. Попробуйте позже.")
		return
	}

	stats, err := b.db.GetUserStats(ctx, user.TelegramID, "today")
	if err != nil {
		b.logger.Error("Ошибка получения статистики", "error", err)
		b.sendMessage(msg.Chat.ID, "❌ Ошибка при получении статистики. Попробуйте позже.")
		return
	}

	response := formatStats(stats, "сегодня")
	b.sendMessage(msg.Chat.ID, response)
}

func (b *Bot) handleStats(ctx context.Context, msg *tgbotapi.Message, user *models.User, period string) {
	if b.db == nil {
		b.sendMessage(msg.Chat.ID, "❌ База данных временно не доступна. Попробуйте позже.")
		return
	}

	if period == "" {
		period = "all"
	}

	// Преобразуем русские названия в английские
	switch strings.ToLower(period) {
	case "сегодня", "today":
		period = "today"
	case "неделя", "week":
		period = "week"
	case "месяц", "month":
		period = "month"
	default:
		period = "all"
	}

	var periodName string
	switch period {
	case "today":
		periodName = "сегодня"
	case "week":
		periodName = "неделю"
	case "month":
		periodName = "месяц"
	default:
		periodName = "всё время"
	}

	stats, err := b.db.GetUserStats(ctx, user.TelegramID, period)
	if err != nil {
		b.logger.Error("Ошибка получения статистики", "error", err)
		b.sendMessage(msg.Chat.ID, "❌ Ошибка при получении статистики. Попробуйте позже.")
		return
	}

	response := formatStats(stats, periodName)
	b.sendMessage(msg.Chat.ID, response)
}
