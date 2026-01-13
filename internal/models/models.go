package models

import (
	"time"

	"github.com/google/uuid"
)

// Expense - модель траты
type Expense struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Amount      float64   `json:"amount" db:"amount"`
	Currency    string    `json:"currency" db:"currency"`
	Category    string    `json:"category" db:"category"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// User - модель пользователя
type User struct {
	ID           int64     `json:"id" db:"id"`
	TelegramID   int64     `json:"telegram_id" db:"telegram_id"`
	Username     string    `json:"username" db:"username"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	LanguageCode string    `json:"language_code" db:"language_code"`
	Currency     string    `json:"currency" db:"currency"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Category - модель категории
type Category struct {
	ID       int64    `json:"id" db:"id"`
	Name     string   `json:"name" db:"name"`
	Emoji    string   `json:"emoji" db:"emoji"`
	Keywords []string `json:"keywords" db:"keywords"`
}

// Статистика
type Stats struct {
	TotalAmount   float64            `json:"total_amount"`
	AverageAmount float64            `json:"average_amount"`
	Count         int                `json:"count"`
	ByCategory    map[string]float64 `json:"by_category"`
}

// DTO для создания траты
type CreateExpenseRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Description string  `json:"description" validate:"required,min=3"`
	Category    string  `json:"category,omitempty"`
}

// Ответ бота
type BotResponse struct {
	Text      string     `json:"text"`
	Keyboard  [][]string `json:"keyboard,omitempty"`
	ParseMode string     `json:"parse_mode,omitempty"`
}
