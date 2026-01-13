package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"finance-tg-bot/internal/models"
)

func (b *Bot) handleStart(msg *tgbotapi.Message, user *models.User) {
	response := fmt.Sprintf(`👋 *Привет, %s!*

Я твой финансовый помощник 🤖

📌 *Что я умею:*
• 💰 Вести учет расходов в удобном диалоговом режиме
• 🏷 Автоматически определять категории трат
• 📊 Показывать подробную статистику
• 💡 Давать советы по экономии

🚀 *Попробуйте прямо сейчас:*
1. Введите команду */add*
2. Следуйте простым подсказкам бота
3. Добавьте свою первую трату за 30 секунд!

📋 *Все команды:* /help

*Удачи в управлении финансами!* 💰`, user.FirstName)

	b.sendMessage(msg.Chat.ID, response)
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	helpText := `📋 *Список всех команд:*

🚀 *Основные команды:*
/start - Начать работу с ботом
/help - Показать это сообщение
/cancel - Отменить текущее действие

💰 *Учет расходов:*
/add - Добавить новую трату (пошаговый режим)
/list - Показать траты с возможностью удаления
/delete [ID] - Удалить конкретную трату
/today - Траты за сегодня

📊 *Статистика и аналитика:*
/stats - Общая статистика расходов
/stats сегодня - Статистика за сегодня
/stats неделя - Статистика за неделю
/stats месяц - Статистика за месяц

🏷 *Категории:*
/categories - Показать все категории трат

🗑️ *Как удалить трату:*
1. Введите команду */list*
2. Нажмите кнопку "🗑️ Удалить" под нужной тратой
3. Подтвердите удаление

💡 *Пример:*
/list → нажимаете "🗑️ Удалить трату 1" → "ДА"

⚡ *Быстрый старт:* просто введите /add и следуйте инструкциям!`

	b.sendMessage(msg.Chat.ID, helpText)
}

func (b *Bot) handleCategories(msg *tgbotapi.Message) {
	categories := []string{
		"🍔 *Еда* - кафе, рестораны, фастфуд, доставка еды",
		"🛒 *Продукты* - супермаркеты, продукты питания",
		"🚗 *Транспорт* - такси, метро, автобус, бензин",
		"🏠 *Коммуналка* - квартплата, электричество, вода, интернет",
		"🎬 *Развлечения* - кино, концерты, театры, музеи",
		"🏥 *Здоровье* - аптеки, лекарства, врачи, спортзал",
		"👕 *Одежда* - магазины одежды, обуви, аксессуары",
		"📚 *Образование* - курсы, книги, обучение",
		"🎁 *Подарки* - подарки, сувениры",
		"📦 *Прочее* - все остальные расходы",
	}

	var sb strings.Builder
	sb.WriteString("🏷 *Доступные категории трат:*\n\n")

	for i, category := range categories {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, category))
	}

	sb.WriteString("\n💡 *Как это работает:*\n")
	sb.WriteString("Бот автоматически определяет категорию по ключевым словам в описания.\n")
	sb.WriteString("Например: \"кофе Starbucks\" → 🍔 Еда\n")
	sb.WriteString("\n✨ *Совет:* Чем подробнее описание, тем точнее категория!")

	b.sendMessage(msg.Chat.ID, sb.String())
}

func (b *Bot) handleClear(msg *tgbotapi.Message, user *models.User) {
	response := `🗑️ *Очистка данных*

⚠️ *Внимание:* Эта функция пока в разработке.

📋 *Что можно сделать сейчас:*
• Просмотреть траты: /list
• Посмотреть статистику: /stats
• Добавить новую трату: /add

🛠 *Для полной очистки данных* обратитесь к администратору.`

	b.sendMessage(msg.Chat.ID, response)
}

func (b *Bot) handleText(ctx context.Context, msg *tgbotapi.Message, user *models.User) {
	userID := msg.From.ID

	// Проверяем состояние пользователя
	session, hasSession := b.getUserSession(userID)

	// Если пользователь в состоянии удаления
	if hasSession && session.State == StateDeletingExpense {
		b.handleDeleteConfirmation(ctx, msg, user, session)
		return
	}

	// Остальная логика обработки текста...
	text := strings.ToLower(msg.Text)

	var response string
	switch {
	case strings.Contains(text, "привет") || strings.Contains(text, "здравствуй") || strings.Contains(text, "хай"):
		response = fmt.Sprintf("Привет, %s! 😊\n\nГотов помочь с учетом финансов! Введите /help для списка команд.", user.FirstName)
	case strings.Contains(text, "спасибо") || strings.Contains(text, "благодар"):
		response = "Всегда рад помочь! 🙏\n\nЕсли есть вопросы - пишите!"
	case strings.Contains(text, "как дела") || strings.Contains(text, "как ты") || strings.Contains(text, "как жизнь"):
		response = "У меня всё отлично! Готов помогать вам с финансами! 💰\n\nХотите добавить трату? Введите /add"
	case strings.Contains(text, "пока") || strings.Contains(text, "до свидания") || strings.Contains(text, "досвидос"):
		response = "До свидания! Возвращайтесь для учета расходов! 📊\n\nНе забудьте записать сегодняшние траты! 😉"
	case strings.Contains(text, "сколько потратил") || strings.Contains(text, "статистик"):
		response = "📊 Чтобы увидеть статистику трат, введите команду /stats\n\nИли /stats сегодня - для статистики за сегодня"
	case strings.Contains(text, "добавить трату") || strings.Contains(text, "добавить расход"):
		response = "💰 Чтобы добавить новую трату, введите команду /add\n\nЯ проведу вас по простым шагам!"
	case strings.Contains(text, "категории") || strings.Contains(text, "категория"):
		response = "🏷 Чтобы посмотреть все категории трат, введите команду /categories"
	default:
		response = `💡 *Я понимаю команды.* Вот что я умею:

💰 *Учет расходов:*
• /add - добавить новую трату (пошагово)
• /list - посмотреть и удалить траты
• /delete [ID] - удалить трату по ID
• /stats - статистика расходов

📋 *Помощь:*
• /help - все команды
• /start - начать работу

🚀 *Попробуйте:* просто введите /add и следуйте инструкциям!`
	}

	b.sendMessage(msg.Chat.ID, response)
}
