package bot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"finance-tg-bot/internal/models"
	"finance-tg-bot/internal/services"
)

// Состояния пользователя
type UserState int

const (
	StateNone UserState = iota
	StateWaitingForAmount
	StateWaitingForDescription
	StateDeletingExpense
)

// Данные пользователя для добавления траты
type UserSession struct {
	State    UserState
	Amount   float64
	TempData string
	Page     int
}

type Bot struct {
	api            *tgbotapi.BotAPI
	expenseService *services.ExpenseService
	logger         *slog.Logger
	db             Database
	// Храним состояние пользователей
	userSessions map[int64]*UserSession
	sessionsMu   sync.RWMutex
}

type Database interface {
	CreateUser(ctx context.Context, user *models.User) error
	CreateExpense(ctx context.Context, expense *models.Expense) error
	GetUserExpenses(ctx context.Context, userID int64, limit, offset int) ([]models.Expense, error)
	GetUserStats(ctx context.Context, userID int64, period string) (*models.Stats, error)
	GetExpenseByID(ctx context.Context, expenseID string) (*models.Expense, error)
	DeleteExpense(ctx context.Context, expenseID string) error
}

func NewBot(token string, expenseService *services.ExpenseService, db Database, logger *slog.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота: %w", err)
	}

	api.Debug = false

	bot := &Bot{
		api:            api,
		expenseService: expenseService,
		db:             db,
		logger:         logger,
		userSessions:   make(map[int64]*UserSession),
	}

	return bot, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Бот запущен", "username", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			b.logger.Info("Бот остановлен")
			return nil

		case update := <-updates:
			if update.Message == nil && update.CallbackQuery == nil {
				continue
			}

			b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	if _, err := b.api.Send(msg); err != nil {
		b.logger.Error("Ошибка отправки сообщения", "error", err)
	}
}
