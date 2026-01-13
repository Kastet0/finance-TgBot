package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"finance-tg-bot/internal/config"
	"finance-tg-bot/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(cfg *config.Config) (*Postgres, error) {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("не удалось проверить соединение: %w", err)
	}

	log.Println("✅ Подключение к PostgreSQL успешно")

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

// User методы
func (p *Postgres) CreateUser(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (
            telegram_id, username, first_name, last_name, 
            language_code, currency, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (telegram_id) DO UPDATE SET
            username = EXCLUDED.username,
            first_name = EXCLUDED.first_name,
            last_name = EXCLUDED.last_name,
            language_code = EXCLUDED.language_code,
            updated_at = EXCLUDED.updated_at
        RETURNING id
    `

	now := time.Now()
	err := p.pool.QueryRow(ctx, query,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.LanguageCode,
		user.Currency,
		now,
		now,
	).Scan(&user.ID)

	return err
}

func (p *Postgres) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	var user models.User
	query := `
        SELECT id, telegram_id, username, first_name, last_name, 
               language_code, currency, created_at, updated_at
        FROM users
        WHERE telegram_id = $1
    `

	err := p.pool.QueryRow(ctx, query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.LanguageCode,
		&user.Currency,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// Expense методы
func (p *Postgres) CreateExpense(ctx context.Context, expense *models.Expense) error {
	query := `
        INSERT INTO expenses (
            id, user_id, amount, currency, category, 
            description, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	now := time.Now()
	_, err := p.pool.Exec(ctx, query,
		expense.ID,
		expense.UserID,
		expense.Amount,
		expense.Currency,
		expense.Category,
		expense.Description,
		now,
		now,
	)

	return err
}

func (p *Postgres) GetUserExpenses(ctx context.Context, userID int64, limit, offset int) ([]models.Expense, error) {
	query := `
        SELECT id, user_id, amount, currency, category, 
               description, created_at, updated_at
        FROM expenses
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := p.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var exp models.Expense
		if err := rows.Scan(
			&exp.ID,
			&exp.UserID,
			&exp.Amount,
			&exp.Currency,
			&exp.Category,
			&exp.Description,
			&exp.CreatedAt,
			&exp.UpdatedAt,
		); err != nil {
			return nil, err
		}
		expenses = append(expenses, exp)
	}

	return expenses, nil
}

// Статистика
func (p *Postgres) GetUserStats(ctx context.Context, userID int64, period string) (*models.Stats, error) {
	var startDate time.Time
	now := time.Now()

	switch period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	default:
		startDate = now.AddDate(-100, 0, 0) // Все время
	}

	// Общая сумма и количество
	query := `
        SELECT COALESCE(SUM(amount), 0), COUNT(*)
        FROM expenses
        WHERE user_id = $1 AND created_at >= $2
    `

	stats := &models.Stats{
		ByCategory: make(map[string]float64),
	}

	err := p.pool.QueryRow(ctx, query, userID, startDate).Scan(
		&stats.TotalAmount,
		&stats.Count,
	)
	if err != nil {
		return nil, err
	}

	if stats.Count > 0 {
		stats.AverageAmount = stats.TotalAmount / float64(stats.Count)
	}

	// Сумма по категориям
	catQuery := `
        SELECT category, COALESCE(SUM(amount), 0)
        FROM expenses
        WHERE user_id = $1 AND created_at >= $2 AND category IS NOT NULL
        GROUP BY category
    `

	rows, err := p.pool.Query(ctx, catQuery, userID, startDate)
	if err != nil {
		return stats, nil // Возвращаем статистику без категорий
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var amount float64
		if err := rows.Scan(&category, &amount); err != nil {
			continue
		}
		stats.ByCategory[category] = amount
	}

	return stats, nil
}

func (p *Postgres) GetExpenseByID(ctx context.Context, expenseID string) (*models.Expense, error) {
	// Парсим UUID из строки
	id, err := uuid.Parse(expenseID)
	if err != nil {
		return nil, fmt.Errorf("неверный формат ID траты: %w", err)
	}

	query := `
		SELECT id, user_id, amount, currency, category, 
		       description, created_at, updated_at
		FROM expenses
		WHERE id = $1
	`

	var expense models.Expense
	err = p.pool.QueryRow(ctx, query, id).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.Amount,
		&expense.Currency,
		&expense.Category,
		&expense.Description,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Трата не найдена
		}
		return nil, fmt.Errorf("ошибка получения траты: %w", err)
	}

	return &expense, nil
}

// DeleteExpense удаляет трату по ID
func (p *Postgres) DeleteExpense(ctx context.Context, expenseID string) error {
	// Парсим UUID из строки
	id, err := uuid.Parse(expenseID)
	if err != nil {
		return fmt.Errorf("неверный формат ID траты: %w", err)
	}

	query := `DELETE FROM expenses WHERE id = $1`

	result, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления траты: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("трата с ID %s не найдена", expenseID)
	}

	log.Printf("Трата удалена: ID=%s", expenseID)
	return nil
}
