package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Telegram TelegramConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Logger   LoggerConfig
}

type AppConfig struct {
	Name    string
	Version string
	Env     string
	Debug   bool
}

type TelegramConfig struct {
	Token   string
	Timeout time.Duration
	Debug   bool
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type LoggerConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	// Устанавливаем значения по умолчанию
	viper.SetDefault("app.name", "FinanceBot")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.debug", true)

	viper.SetDefault("telegram.timeout", 60)
	viper.SetDefault("telegram.debug", false)

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.name", "finance_bot")
	viper.SetDefault("database.ssl_mode", "disable")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")

	// Читаем из .env файла
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Автоматически читаем переменные окружения
	viper.AutomaticEnv()

	// Пробуем прочитать .env файл
	if err := viper.ReadInConfig(); err != nil {
		// Если файла нет, используем переменные окружения
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("ошибка чтения конфига: %w", err)
		}
	}

	var cfg Config

	// Заполняем конфигурацию
	cfg.App.Name = viper.GetString("app.name")
	cfg.App.Version = viper.GetString("app.version")
	cfg.App.Env = viper.GetString("app.env")
	cfg.App.Debug = viper.GetBool("app.debug")

	cfg.Telegram.Token = viper.GetString("telegram_token")
	if cfg.Telegram.Token == "" {
		return nil, fmt.Errorf("telegram_token не установлен")
	}
	cfg.Telegram.Timeout = viper.GetDuration("telegram.timeout")
	cfg.Telegram.Debug = viper.GetBool("telegram.debug")

	cfg.Database.Host = viper.GetString("db_host")
	cfg.Database.Port = viper.GetInt("db_port")
	cfg.Database.User = viper.GetString("db_user")
	cfg.Database.Password = viper.GetString("db_password")
	cfg.Database.Name = viper.GetString("db_name")
	cfg.Database.SSLMode = viper.GetString("db_ssl_mode")

	cfg.Redis.Host = viper.GetString("redis_host")
	cfg.Redis.Port = viper.GetInt("redis_port")
	cfg.Redis.Password = viper.GetString("redis_password")
	cfg.Redis.DB = viper.GetInt("redis_db")

	cfg.Logger.Level = viper.GetString("log_level")
	cfg.Logger.Format = viper.GetString("log_format")

	return &cfg, nil
}

// Helper функция для получения порта из строки
func GetPort() int {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return 8080
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 8080
	}

	return port
}
