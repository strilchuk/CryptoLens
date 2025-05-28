package db

import (
	"CryptoLens_Backend/env"
	"CryptoLens_Backend/logger"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

var (
	dbUser          string
	dbName          string
	dbPass          string
	dbHost          string
	dbConnectionStr string
)

func init() {
	env.Init()
	dbUser = env.GetDBUser()
	dbName = env.GetDBName()
	dbPass = env.GetDBPass()
	dbHost = env.GetDBHost()
}

func InitDB() (*sql.DB, error) {
	dbConnectionStr = fmt.Sprintf(
		"user=%s dbname=%s sslmode=disable password=%s host=%s port=5432", dbUser, dbName, dbPass, dbHost,
	)

	db, err := sql.Open("postgres", dbConnectionStr)
	if err != nil {
		logger.Log.Printf("Error connecting to database: %v", err)
		return nil, fmt.Errorf("Error connecting to database: %v", err)
	}

	if err := migrateDB(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrateDB(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Log.Printf("Could not create migrate instance: %v", err)
		return fmt.Errorf("Could not create migrate instance: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver,
	)
	if err != nil {
		logger.Log.Printf("Could not create migration: %v", err)
		return fmt.Errorf("Could not create migration: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log.Printf("Could not run migration: %v", err)
		return fmt.Errorf("Could not run migration: %v", err)
	}
	return nil
}
