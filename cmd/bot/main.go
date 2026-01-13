package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"finance-tg-bot/internal/bot"
	"finance-tg-bot/internal/config"
	"finance-tg-bot/internal/database"
	"finance-tg-bot/internal/services"
)

func main() {
	// 1. Настраиваем логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("🚀 Запуск Finance Telegram Bot")

	// 2. Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		slog.Error("❌ Ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}

	slog.Info("✅ Конфигурация загружена")

	// 3. Настраиваем graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 4. Подключаемся к базе данных
	slog.Info("🔗 Подключение к PostgreSQL...")
	db, err := database.NewPostgres(cfg)
	if err != nil {
		slog.Error("❌ Ошибка подключения к БД", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("✅ Подключение к БД успешно")

	// 5. Инициализируем сервисы
	slog.Info("⚙️ Инициализация сервисов...")

	// Детектор категорий
	categoryDetector := services.NewSimpleCategoryDetector()

	// Сервис трат
	expenseService := services.NewExpenseService(db, categoryDetector)

	// 6. Создаем бота
	slog.Info("🤖 Создание Telegram бота...")
	telegramBot, err := bot.NewBot(
		cfg.Telegram.Token,
		expenseService,
		db,
		logger,
	)
	if err != nil {
		slog.Error("❌ Ошибка создания бота", "error", err)
		os.Exit(1)
	}

	slog.Info("✅ Бот создан успешно")

	// 7. Запускаем бота
	slog.Info("▶️ Запуск бота...")

	if err := telegramBot.Start(ctx); err != nil {
		slog.Error("❌ Бот завершился с ошибкой", "error", err)
		os.Exit(1)
	}

	slog.Info("👋 Бот завершил работу")
}
