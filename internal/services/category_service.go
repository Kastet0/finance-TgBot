package services

import (
	"strings"
)

// SimpleCategoryDetector - простой детектор категорий по ключевым словам
type SimpleCategoryDetector struct {
	categories map[string][]string
}

func NewSimpleCategoryDetector() *SimpleCategoryDetector {
	return &SimpleCategoryDetector{
		categories: map[string][]string{
			"🍔 Еда":         {"кафе", "ресторан", "макд", "бургер", "пицца", "суши", "кофе", "starbucks", "шаурма"},
			"🛒 Продукты":    {"пятерочка", "магнит", "ашан", "лента", "продукты", "молоко", "хлеб", "колбаса"},
			"🚗 Транспорт":   {"такси", "метро", "автобус", "бензин", "заправка", "парковка", "yandex", "uber"},
			"🏠 Коммуналка":  {"квартплата", "электричество", "вода", "газ", "интернет", "телефон", "тв"},
			"🎬 Развлечения": {"кино", "концерт", "театр", "музей", "парк", "аттракцион", "билет"},
			"🏥 Здоровье":    {"аптека", "врач", "больница", "лекарства", "таблетки", "витамины", "спортзал"},
			"👕 Одежда":      {"одежда", "обувь", "магазин", "шопинг", "футболка", "джинсы", "куртка"},
			"📚 Образование": {"курсы", "книги", "учебник", "обучение", "семинар", "вебинар"},
			"🎁 Подарки":     {"подарок", "сюрприз", "день рождения", "новый год", "поздравление"},
		},
	}
}

func (d *SimpleCategoryDetector) DetectCategory(description string) (string, float64) {
	desc := strings.ToLower(description)

	// Ищем ключевые слова
	for category, keywords := range d.categories {
		for _, keyword := range keywords {
			if strings.Contains(desc, keyword) {
				return category, 0.9 // Высокая уверенность
			}
		}
	}

	// Если не нашли, возвращаем "Прочее"
	return "📦 Прочее", 0.1
}

// GetCategories возвращает список доступных категорий
func (d *SimpleCategoryDetector) GetCategories() []string {
	categories := make([]string, 0, len(d.categories))
	for category := range d.categories {
		categories = append(categories, category)
	}
	return append(categories, "📦 Прочее")
}

// SuggestCategories предлагает категории по описанию
func (d *SimpleCategoryDetector) SuggestCategories(description string) []string {
	desc := strings.ToLower(description)
	suggestions := make([]string, 0)

	for category, keywords := range d.categories {
		for _, keyword := range keywords {
			if strings.Contains(desc, keyword) && !contains(suggestions, category) {
				suggestions = append(suggestions, category)
			}
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "📦 Прочее")
	}

	return suggestions
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
