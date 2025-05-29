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

func GetServerPort() string {
	return os.Getenv("SERVER_PORT")
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

func GetRedisHost() string {
	return os.Getenv("REDIS_HOST")
}

func GetRedisPassword() string {
	return os.Getenv("REDIS_PASSWORD")
}

func GetRedisPortLocal() string {
	return os.Getenv("REDIS_PORT_LOCAL")
}

func GetBybitApiUrl() string {
	return os.Getenv("BYBIT_API_URL")
}

func GetBybitApiTestUrl() string {
	return os.Getenv("BYBIT_API_TEST_URL")
}

func GetBybitWsUrl() string {
	return os.Getenv("BYBIT_WS_URL")
}

func GetBybitWsTestUrl() string {
	return os.Getenv("BYBIT_WS_TEST_URL")
}

func GetBybitRecvWindow() string {
	return os.Getenv("BYBIT_RECV_WINDOW")
}

func GetBybitInstrumentsUpdateInterval() string {
	return os.Getenv("BYBIT_INSTRUMENTS_UPDATE_INTERVAL")
}

func GetBybitApiMode() string {
	return os.Getenv("BYBIT_API_MODE")
}

func GetJWTSecret() string {
	return os.Getenv("JWT_SECRET")
}
