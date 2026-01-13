package services

import (
	"context"
	"fmt"
	"strings"

	"finance-tg-bot/internal/models"

	"github.com/google/uuid"
)

type ExpenseRepository interface {
	CreateExpense(ctx context.Context, expense *models.Expense) error
	GetUserExpenses(ctx context.Context, userID int64, limit, offset int) ([]models.Expense, error)
	GetUserStats(ctx context.Context, userID int64, period string) (*models.Stats, error)
}

type CategoryDetector interface {
	DetectCategory(description string) (string, float64)
}

type ExpenseService struct {
	repo     ExpenseRepository
	detector CategoryDetector
}

func NewExpenseService(repo ExpenseRepository, detector CategoryDetector) *ExpenseService {
	return &ExpenseService{
		repo:     repo,
		detector: detector,
	}
}

// CreateExpense создает новую трату
func (s *ExpenseService) CreateExpense(ctx context.Context, req *models.CreateExpenseRequest, userID int64) (*models.Expense, error) {
	// Определяем категорию автоматически
	category, _ := s.detector.DetectCategory(req.Description)

	// Если пользователь указал категорию, используем её
	if req.Category != "" {
		category = req.Category
	}

	expense := &models.Expense{
		ID:          uuid.New(),
		UserID:      userID,
		Amount:      req.Amount,
		Currency:    "RUB", // По умолчанию
		Category:    category,
		Description: req.Description,
	}

	if err := s.repo.CreateExpense(ctx, expense); err != nil {
		return nil, fmt.Errorf("ошибка сохранения траты: %w", err)
	}

	return expense, nil
}

// GetExpenses возвращает траты пользователя
func (s *ExpenseService) GetExpenses(ctx context.Context, userID int64, page, limit int) ([]models.Expense, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit
	return s.repo.GetUserExpenses(ctx, userID, limit, offset)
}

// GetStats возвращает статистику
func (s *ExpenseService) GetStats(ctx context.Context, userID int64, period string) (*models.Stats, error) {
	return s.repo.GetUserStats(ctx, userID, period)
}

// FormatStats форматирует статистику для Telegram
func (s *ExpenseService) FormatStats(stats *models.Stats, period string) string {
	if stats.Count == 0 {
		return "📭 У вас пока нет трат за этот период."
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📊 *Статистика за %s:*\n\n", period))
	sb.WriteString(fmt.Sprintf("💰 Всего потрачено: *%.2f руб.*\n", stats.TotalAmount))
	sb.WriteString(fmt.Sprintf("📈 Средний чек: *%.2f руб.*\n", stats.AverageAmount))
	sb.WriteString(fmt.Sprintf("📝 Количество трат: *%d*\n\n", stats.Count))

	if len(stats.ByCategory) > 0 {
		sb.WriteString("*Распределение по категориям:*\n")
		for category, amount := range stats.ByCategory {
			percentage := (amount / stats.TotalAmount) * 100
			sb.WriteString(fmt.Sprintf("• %s: %.2f руб. (%.1f%%)\n",
				category, amount, percentage))
		}
	}

	// Добавляем совет
	sb.WriteString("\n💡 *Совет:* ")
	if stats.AverageAmount > 1000 {
		sb.WriteString("Попробуйте сократить ежедневные траты на 20%")
	} else if stats.AverageAmount > 500 {
		sb.WriteString("Хороший результат! Можно еще лучше!")
	} else {
		sb.WriteString("Отлично! Вы хорошо контролируете расходы!")
	}

	return sb.String()
}

// FormatExpense форматирует трату для отображения
func (s *ExpenseService) FormatExpense(expense models.Expense) string {
	return fmt.Sprintf(
		"💰 *%.2f руб.*\n"+
			"📝 %s\n"+
			"🏷 %s\n"+
			"📅 %s",
		expense.Amount,
		expense.Description,
		expense.Category,
		expense.CreatedAt.Format("02.01.2006 15:04"),
	)
}

// FormatExpensesList форматирует список трат
func (s *ExpenseService) FormatExpensesList(expenses []models.Expense) string {
	if len(expenses) == 0 {
		return "📭 У вас пока нет трат."
	}

	var sb strings.Builder
	sb.WriteString("📝 *Последние траты:*\n\n")

	total := 0.0
	for i, exp := range expenses {
		sb.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, s.FormatExpense(exp)))
		total += exp.Amount
	}

	sb.WriteString(fmt.Sprintf("💸 *Итого: %.2f руб.*", total))
	return sb.String()
}
