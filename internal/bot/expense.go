package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"finance-tg-bot/internal/models"
)

func (b *Bot) startAddExpenseFlow(msg *tgbotapi.Message) {
	userID := msg.From.ID

	// Создаем новую сессию
	session := &UserSession{
		State: StateWaitingForAmount,
	}

	b.setUserSession(userID, session)

	// Запрашиваем сумму
	response := `💸 *Добавление новой траты*

💰 *Шаг 1 из 2:* Введите сумму траты:

• Только число (например: 500)
• Можно с копейками: 149.99 или 149,99
• Без валюты и других символов

❌ *Для отмены введите:* /cancel`

	b.sendMessage(msg.Chat.ID, response)
}

func (b *Bot) handleAmountInput(ctx context.Context, msg *tgbotapi.Message, user *models.User, session *UserSession) {
	// Проверяем на отмену
	if msg.IsCommand() && msg.Command() == "cancel" {
		b.handleCancel(msg)
		return
	}

	// Парсим сумму
	amount, err := parseAmount(msg.Text)
	if err != nil || amount <= 0 {
		b.sendMessage(msg.Chat.ID, "❌ *Неправильный формат суммы!*\n\nПожалуйста, введите положительное число.\n\n📝 *Примеры:*\n• 500\n• 149.99\n• 75,50\n\nПопробуйте еще раз:")
		return
	}

	// Сохраняем сумму
	session.Amount = amount
	session.State = StateWaitingForDescription
	b.setUserSession(user.TelegramID, session)

	// Запрашиваем описание
	response := fmt.Sprintf(`✅ *Шаг 1 завершен!* Сумма: *%.2f руб.*

📝 *Шаг 2 из 2:* Введите описание траты:

• Что купили: "кофе Starbucks"
• Где купили: "обед в кафе"  
• Детали: "продукты в Пятерочке"

💡 *Чем подробнее описание, тем точнее определится категория!*

❌ *Для отмены введите:* /cancel`, amount)

	b.sendMessage(msg.Chat.ID, response)
}

func (b *Bot) handleDescriptionInput(ctx context.Context, msg *tgbotapi.Message, user *models.User, session *UserSession) {
	// Проверяем на отмену
	if msg.IsCommand() && msg.Command() == "cancel" {
		b.handleCancel(msg)
		return
	}

	description := strings.TrimSpace(msg.Text)
	if len(description) < 2 {
		b.sendMessage(msg.Chat.ID, "❌ *Слишком короткое описание!*\n\nПожалуйста, введите более подробное описание (минимум 2 символа).\n\nПример: \"кофе в Starbucks на вынос\"")
		return
	}

	if len(description) > 500 {
		b.sendMessage(msg.Chat.ID, "❌ *Слишком длинное описание!*\n\nПожалуйста, сократите описание до 500 символов.")
		return
	}

	// Создаем запрос
	req := &models.CreateExpenseRequest{
		Amount:      session.Amount,
		Description: description,
	}

	// Создаем трату через сервис
	expense, err := b.expenseService.CreateExpense(ctx, req, user.TelegramID)
	if err != nil {
		b.logger.Error("Ошибка создания траты", "error", err)
		b.sendMessage(msg.Chat.ID, "❌ Ошибка при сохранении траты. Попробуйте еще раз: /add")
		b.clearUserSession(user.TelegramID)
		return
	}

	// Форматируем ответ с ПРАВИЛЬНОЙ датой
	currentTime := time.Now()
	if !expense.CreatedAt.IsZero() {
		currentTime = expense.CreatedAt
	}

	response := fmt.Sprintf(`🎉 *Трата успешно добавлена!*

📋 *Детали траты:*
💰 *Сумма:* %.2f руб.
📝 *Описание:* %s
🏷 *Категория:* %s
📅 *Дата:* %s

💡 *Категория определена автоматически на основе описания!*

🔄 *Что дальше?*
• Добавить еще трату: /add
• Посмотреть статистику: /stats
• Список всех трат: /list
• Помощь по командам: /help`,
		expense.Amount,
		expense.Description,
		expense.Category,
		currentTime.Format("02.01.2006 15:04"))

	// Сначала отправляем успешное сообщение
	b.sendMessage(msg.Chat.ID, response)

	// Очищаем сессию
	b.clearUserSession(user.TelegramID)
}
