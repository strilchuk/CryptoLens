package env

import (
	"CryptoLens_Backend/logger"
	"github.com/joho/godotenv"
	"os"
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.Log.Printf("Error loading .env file: %v\n", err)
	}
}

func GetDBUser() string {
	return os.Getenv("DB_USERNAME")
}

func GetDBName() string {
	return os.Getenv("DB_DATABASE")
}

func GetDBPass() string {
	return os.Getenv("DB_PASSWORD")
}

func GetDBHost() string {
	return os.Getenv("DB_HOST")
}

func GetToken() string {
	return os.Getenv("TOKEN")
}
