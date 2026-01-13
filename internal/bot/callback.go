package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"finance-tg-bot/internal/models"
)

func (b *Bot) handleCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	// Отвечаем на callback (убираем часики)
	b.api.Send(tgbotapi.NewCallback(callback.ID, ""))

	data := callback.Data
	chatID := callback.Message.Chat.ID
	userID := callback.From.ID

	b.logger.Debug("Получен callback", "user_id", userID, "data", data)

	// Создаем пользователя для обработки
	user := b.createUserFromCallback(callback)

	// Сохраняем пользователя
	go b.saveUser(ctx, user)

	// Обрабатываем callback
	switch {
	case data == "close_list":
		b.handleCloseList(callback)
	case strings.HasPrefix(data, "delete_"):
		expenseID := strings.TrimPrefix(data, "delete_")
		b.handleDeleteCallback(ctx, callback, user, expenseID)
	case strings.HasPrefix(data, "list_"):
		pageStr := strings.TrimPrefix(data, "list_")
		page, _ := strconv.Atoi(pageStr)

		// Удаляем старое сообщение
		b.api.Send(tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID))

		// Показываем новую страницу
		b.handleListExpenses(ctx, &tgbotapi.Message{
			Chat:      callback.Message.Chat,
			From:      callback.From,
			MessageID: callback.Message.MessageID,
		}, user, page)
	}
}

func (b *Bot) createUserFromCallback(callback *tgbotapi.CallbackQuery) *models.User {
	return &models.User{
		TelegramID:   callback.From.ID,
		Username:     callback.From.UserName,
		FirstName:    callback.From.FirstName,
		LastName:     callback.From.LastName,
		LanguageCode: callback.From.LanguageCode,
		Currency:     "RUB",
	}
}

func (b *Bot) handleCloseList(callback *tgbotapi.CallbackQuery) {
	// Удаляем сообщение со списком
	b.api.Send(tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID))

	// Отправляем подтверждение
	b.sendMessage(callback.Message.Chat.ID, "✅ Список закрыт.")
}

func (b *Bot) handleDeleteCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, user *models.User, expenseID string) {
	// Проверяем, что трата принадлежит пользователю
	expense, err := b.db.GetExpenseByID(ctx, expenseID)
	if err != nil {
		b.logger.Error("Ошибка получения траты", "error", err)
		b.sendMessage(callback.Message.Chat.ID, "❌ Ошибка при получении данных о трате.")
		return
	}

	if expense == nil {
		b.sendMessage(callback.Message.Chat.ID, "❌ Трата не найдена.")
		return
	}

	if expense.UserID != user.TelegramID {
		b.sendMessage(callback.Message.Chat.ID, "❌ Вы не можете удалить чужую трату.")
		return
	}

	// Удаляем сообщение со списком
	b.api.Send(tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID))

	// Создаем сессию удаления
	session := &UserSession{
		State:    StateDeletingExpense,
		TempData: expenseID,
	}
	b.setUserSession(user.TelegramID, session)

	// Отправляем сообщение с подтверждением
	response := fmt.Sprintf(`🗑️ *Подтверждение удаления*

Вы действительно хотите удалить эту трату?

📋 *Детали траты:*
💰 Сумма: *%.2f руб.*
📝 Описание: *%s*
🏷 Категория: *%s*
📅 Дата: *%s*

⚠️ *Внимание:* Это действие нельзя отменить!

✅ Для подтверждения введите: *ДА*
❌ Для отмены введите: *НЕТ* или /cancel`,
		expense.Amount,
		expense.Description,
		expense.Category,
		expense.CreatedAt.Format("02.01.2006 15:04"))

	b.sendMessage(callback.Message.Chat.ID, response)
}

func (b *Bot) handleDeleteConfirmation(ctx context.Context, msg *tgbotapi.Message, user *models.User, session *UserSession) {
	text := strings.ToLower(strings.TrimSpace(msg.Text))

	switch text {
	case "да", "yes", "удалить", "подтверждаю":
		// Удаляем трату
		expenseID := session.TempData
		err := b.db.DeleteExpense(ctx, expenseID)

		if err != nil {
			b.logger.Error("Ошибка удаления траты", "error", err)
			b.sendMessage(msg.Chat.ID, "❌ Ошибка при удалении траты. Попробуйте позже.")
		} else {
			b.sendMessage(msg.Chat.ID, "✅ *Трата успешно удалена!*\n\n🗑️ Трата была удалена из вашей истории.")
		}

		// Очищаем сессию
		b.clearUserSession(user.TelegramID)

	case "нет", "no", "отмена", "отменить":
		b.sendMessage(msg.Chat.ID, "❌ *Удаление отменено.*\n\nТрата не была удалена.")
		b.clearUserSession(user.TelegramID)

	default:
		b.sendMessage(msg.Chat.ID, "❌ *Не понял ответ.*\n\nПожалуйста, введите:\n• *ДА* - для подтверждения удаления\n• *НЕТ* - для отмены\n\nИли /cancel для выхода.")
	}
}
