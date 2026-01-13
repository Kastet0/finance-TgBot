package bot

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"finance-tg-bot/internal/models"
)

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	// Обрабатываем callback-запросы
	if update.CallbackQuery != nil {
		b.handleCallback(ctx, update.CallbackQuery)
		return
	}

	msg := update.Message
	if msg == nil {
		return
	}

	b.handleMessage(ctx, msg)
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	b.logger.Debug("Получено сообщение",
		"user_id", userID,
		"username", msg.From.UserName,
		"text", msg.Text)

	// Создаем пользователя
	user := b.createUserFromMessage(msg)

	// Сохраняем пользователя в БД
	go b.saveUser(ctx, user)

	// Проверяем состояние пользователя
	session, hasSession := b.getUserSession(userID)

	if hasSession && session.State != StateNone {
		// Пользователь в процессе добавления/удаления траты
		b.handleUserSession(ctx, msg, user, session)
		return
	}

	// Обычная обработка команд
	if msg.IsCommand() {
		b.handleCommand(ctx, msg, user)
	} else {
		b.handleText(ctx, msg, user)
	}
}

func (b *Bot) createUserFromMessage(msg *tgbotapi.Message) *models.User {
	return &models.User{
		TelegramID:   msg.From.ID,
		Username:     msg.From.UserName,
		FirstName:    msg.From.FirstName,
		LastName:     msg.From.LastName,
		LanguageCode: msg.From.LanguageCode,
		Currency:     "RUB",
	}
}

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message, user *models.User) {
	cmd := msg.Command()
	args := msg.CommandArguments()

	switch cmd {
	case "start":
		b.handleStart(msg, user)
	case "help":
		b.handleHelp(msg)
	case "add":
		b.startAddExpenseFlow(msg)
	case "list":
		b.handleListExpenses(ctx, msg, user, 0)
	case "delete":
		b.handleDeleteCommand(ctx, msg, user, args)
	case "today":
		b.handleTodayStats(ctx, msg, user)
	case "stats":
		b.handleStats(ctx, msg, user, args)
	case "categories":
		b.handleCategories(msg)
	case "clear":
		b.handleClear(msg, user)
	case "cancel":
		b.handleCancel(msg)
	default:
		// Проверяем, может быть это callback от кнопки
		if strings.HasPrefix(cmd, "delete_") {
			expenseID := strings.TrimPrefix(cmd, "delete_")
			callback := &tgbotapi.CallbackQuery{
				From:    msg.From,
				Message: msg,
				Data:    "delete_" + expenseID,
			}
			b.handleDeleteCallback(ctx, callback, user, expenseID)
		} else if strings.HasPrefix(cmd, "list_") {
			// Обработка пагинации списка
			pageStr := strings.TrimPrefix(cmd, "list_")
			page, _ := strconv.Atoi(pageStr)
			b.handleListExpenses(ctx, msg, user, page)
		} else {
			b.sendMessage(msg.Chat.ID, "❌ Неизвестная команда. Введите /help для списка команд.")
		}
	}
}
