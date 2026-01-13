package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func main() {
	// Загружаем конфиг
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Предупреждение: не удалось прочитать .env файл: %v", err)
	}

	// Подключаемся к БД
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("DB_HOST"),
		viper.GetString("DB_PORT"),
		viper.GetString("DB_USER"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_NAME"),
		viper.GetString("DB_SSL_MODE"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	// Проверяем соединение
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Ошибка ping БД:", err)
	}

	log.Println("✅ Подключение к БД успешно")

	// Читаем файлы миграций
	migrationDir := "migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		log.Fatal("Ошибка чтения директории миграций:", err)
	}

	// Применяем миграции по порядку
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			applyMigration(db, filepath.Join(migrationDir, file.Name()))
		}
	}

	log.Println("✅ Все миграции применены успешно")
}

func applyMigration(db *sql.DB, filePath string) {
	log.Printf("Применяем миграцию: %s", filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла %s: %v", filePath, err)
	}

	// Выполняем SQL
	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatalf("Ошибка выполнения миграции %s: %v", filePath, err)
	}

	log.Printf("✅ Миграция %s применена успешно", filePath)
}
