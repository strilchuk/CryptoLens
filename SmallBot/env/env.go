package env

import (
	"SmallBot/logger"
	"github.com/joho/godotenv"
	"os"
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.Log.Printf("Error loading .env file: %v\n", err)
	}
}

func GetServerPort() string {
	return os.Getenv("SERVER_PORT")
}

func GetBybitApiUrl() string {
	if os.Getenv("BYBIT_API_MODE") == "test" {
		return os.Getenv("BYBIT_API_TEST_URL")
	} else {
		return os.Getenv("BYBIT_API_URL")
	}
}

func GetBybitWsUrl() string {
	if os.Getenv("BYBIT_API_MODE") == "test" {
		return os.Getenv("BYBIT_WS_TEST_URL")
	} else {
		return os.Getenv("BYBIT_WS_URL")
	}
}

func GetBybitApiToken() string {
	if os.Getenv("BYBIT_API_MODE") == "test" {
		return os.Getenv("BYBIT_API_TOKEN_TEST")
	} else {
		return os.Getenv("BYBIT_API_TOKEN")
	}
}

func GetBybitApiSecret() string {
	if os.Getenv("BYBIT_API_MODE") == "test" {
		return os.Getenv("BYBIT_API_SECRET_TEST")
	} else {
		return os.Getenv("BYBIT_API_SECRET")
	}
}

func GetBybitRecvWindow() string {
	return os.Getenv("BYBIT_RECV_WINDOW")
}

func GetBybitApiMode() string {
	return os.Getenv("BYBIT_API_MODE")
}

func GetDebug() string {
	return os.Getenv("DEBUG")
}

func GetSymbol() string {
	return os.Getenv("SYMBOL")
}
