package bot

import (
	"context"
	"time"

	"finance-tg-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) saveUser(ctx context.Context, user *models.User) {
	// Пытаемся сохранить пользователя
	if b.db != nil {
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		if err := b.db.CreateUser(ctx, user); err != nil {
			b.logger.Error("Ошибка сохранения пользователя", "error", err)
		}
	}
}

func (b *Bot) getUserSession(userID int64) (*UserSession, bool) {
	b.sessionsMu.RLock()
	session, exists := b.userSessions[userID]
	b.sessionsMu.RUnlock()
	return session, exists
}

func (b *Bot) setUserSession(userID int64, session *UserSession) {
	b.sessionsMu.Lock()
	b.userSessions[userID] = session
	b.sessionsMu.Unlock()
}

func (b *Bot) clearUserSession(userID int64) {
	b.sessionsMu.Lock()
	delete(b.userSessions, userID)
	b.sessionsMu.Unlock()
}

func (b *Bot) handleUserSession(ctx context.Context, msg *tgbotapi.Message, user *models.User, session *UserSession) {
	switch session.State {
	case StateWaitingForAmount:
		b.handleAmountInput(ctx, msg, user, session)
	case StateWaitingForDescription:
		b.handleDescriptionInput(ctx, msg, user, session)
	case StateDeletingExpense:
		b.handleDeleteConfirmation(ctx, msg, user, session)
	}
}

func (b *Bot) handleCancel(msg *tgbotapi.Message) {
	userID := msg.From.ID

	session, hasSession := b.getUserSession(userID)

	if hasSession && session.State != StateNone {
		var message string
		switch session.State {
		case StateWaitingForAmount, StateWaitingForDescription:
			message = "❌ *Добавление траты отменено.*\n\nВы можете начать заново командой: /add"
		case StateDeletingExpense:
			message = "❌ *Удаление траты отменено.*\n\nТрата не была удалена."
		}

		// Очищаем сессию
		b.clearUserSession(userID)

		// Отправляем сообщение об отмене
		b.sendMessage(msg.Chat.ID, message)
	} else {
		b.sendMessage(msg.Chat.ID, "ℹ️ *Нечего отменять.*\n\nВы не находитесь в процессе добавления или удаления траты.")
	}
}
