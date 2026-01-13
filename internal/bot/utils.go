package bot

import (
	"context"
	"fmt"
	"strings"

	"finance-tg-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// parseAmount парсит сумму из строки
func parseAmount(s string) (float64, error) {
	// Убираем запятые, пробелы и лишние символы
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSpace(s)

	// Убираем валюту и другие символы в конце
	currencySymbols := []string{"руб", "р", "RUB", "₽", "$", "€", "£", "¥", "usd", "eur"}
	for _, symbol := range currencySymbols {
		s = strings.TrimSuffix(strings.ToLower(s), strings.ToLower(symbol))
	}

	// Парсим число
	var amount float64
	_, err := fmt.Sscanf(s, "%f", &amount)
	if err != nil {
		// Пробуем убрать все нецифровые символы кроме точки
		var cleanStr strings.Builder
		for _, r := range s {
			if (r >= '0' && r <= '9') || r == '.' || r == ',' {
				if r == ',' {
					cleanStr.WriteRune('.')
				} else {
					cleanStr.WriteRune(r)
				}
			}
		}
		_, err = fmt.Sscanf(cleanStr.String(), "%f", &amount)
	}

	return amount, err
}

// formatStats форматирует статистику для вывода
func formatStats(stats *models.Stats, period string) string {
	if stats == nil || stats.Count == 0 {
		return fmt.Sprintf("📭 *У вас пока нет трат за %s.*\n\nДобавьте первую трату командой: /add", period)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 *Статистика за %s:*\n\n", period))
	sb.WriteString(fmt.Sprintf("💰 *Всего потрачено:* %.2f руб.\n", stats.TotalAmount))
	sb.WriteString(fmt.Sprintf("📈 *Средний чек:* %.2f руб.\n", stats.AverageAmount))
	sb.WriteString(fmt.Sprintf("📝 *Количество трат:* %d\n\n", stats.Count))

	if len(stats.ByCategory) > 0 {
		sb.WriteString("*📋 Распределение по категориям:*\n")
		for category, amount := range stats.ByCategory {
			percentage := (amount / stats.TotalAmount) * 100
			emoji := getCategoryEmoji(category)
			sb.WriteString(fmt.Sprintf("%s *%s:* %.2f руб. (%.1f%%)\n",
				emoji, category, amount, percentage))
		}
		sb.WriteString("\n")
	}

	// Добавляем совет на основе статистики
	sb.WriteString("💡 *Совет по экономии:*\n")
	if stats.AverageAmount > 1500 {
		sb.WriteString("Вы тратите довольно много. Попробуйте:\n• Сократить ежедневные траты на 30%\n• Готовить дома вместо кафе\n• Использовать публичный транспорт")
	} else if stats.AverageAmount > 800 {
		sb.WriteString("Хороший результат! Можно лучше:\n• Сократить ненужные мелкие покупки\n• Планировать бюджет на неделю\n• Откладывать 10% от доходов")
	} else if stats.AverageAmount > 300 {
		sb.WriteString("Отлично! Вы контролируете расходы.\n• Продолжайте в том же духе!\n• Создайте финансовую подушку\n• Инвестируйте сэкономленные деньги")
	} else {
		sb.WriteString("Превосходно! Вы образец экономии.\n• Подумайте об инвестициях\n• Поставьте финансовые цели\n• Помогите другим научиться экономить")
	}

	sb.WriteString(fmt.Sprintf("\n\n🔄 *Добавить новую трату:* /add"))

	return sb.String()
}

// getCategoryEmoji возвращает эмодзи для категории
func getCategoryEmoji(category string) string {
	switch {
	case strings.Contains(category, "Еда"):
		return "🍔"
	case strings.Contains(category, "Продукты"):
		return "🛒"
	case strings.Contains(category, "Транспорт"):
		return "🚗"
	case strings.Contains(category, "Коммуналка"):
		return "🏠"
	case strings.Contains(category, "Развлечения"):
		return "🎬"
	case strings.Contains(category, "Здоровье"):
		return "🏥"
	case strings.Contains(category, "Одежда"):
		return "👕"
	case strings.Contains(category, "Образование"):
		return "📚"
	case strings.Contains(category, "Подарки"):
		return "🎁"
	default:
		return "📦"
	}
}

// handleDeleteCommand обрабатывает команду /delete
func (b *Bot) handleDeleteCommand(ctx context.Context, msg *tgbotapi.Message, user *models.User, args string) {
	if args == "" {
		// Показываем список для удаления
		b.handleListExpenses(ctx, msg, user, 0)
		return
	}

	// Пытаемся удалить по ID из аргументов
	expenseID := strings.TrimSpace(args)

	// Создаем фейковый callback для обработки
	callback := &tgbotapi.CallbackQuery{
		From:    msg.From,
		Message: msg,
		Data:    "delete_" + expenseID,
	}

	b.handleDeleteCallback(ctx, callback, user, expenseID)
}
