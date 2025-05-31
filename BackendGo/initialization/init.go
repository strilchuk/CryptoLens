package initialization

import (
	"CryptoLens_Backend/db"
	"CryptoLens_Backend/integration/redis"
	"CryptoLens_Backend/logger"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"log"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	DB *sql.DB
)

// Initialize выполняет инициализацию всех компонентов приложения
func Initialize() {
	initLogger()
	initDB()
	applyMigrations()
	initRedis()
}

// initDB инициализирует подключение к базе данных
func initDB() {
	var err error
	DB, err = db.InitDB()
	if err != nil {
		log.Fatal(err)
	}
}

// initLogger инициализирует систему логирования
func initLogger() {
	err := logger.Init("logs/app.log")
	if err != nil {
		log.Fatal(err)
	}
}

// applyMigrations применяет миграции базы данных
func applyMigrations() {
	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
}

func initRedis() {
	if err := redis.Init(); err != nil {
		log.Fatal(err)
	}
}
