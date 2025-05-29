package types

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/models"
	"context"
	"net/http"
)

// BybitServiceInterface определяет интерфейс для сервиса Bybit
type BybitServiceInterface interface {
	GetWalletBalance(ctx context.Context, token string) (*bybit.BybitWalletBalance, error)
	GetFeeRate(ctx context.Context, token string, category string, symbol string, baseCoin string) (*bybit.BybitFeeRateResponse, error)
	GetInstruments(ctx context.Context, category string) ([]models.BybitInstrument, error)
	StartInstrumentsUpdate(ctx context.Context)
	StartWebSocket(ctx context.Context)
	StartBackgroundTasks(ctx context.Context)
}

// BybitHandlerInterface определяет интерфейс для обработчика Bybit
type BybitHandlerInterface interface {
	GetWalletBalance(w http.ResponseWriter, r *http.Request)
	GetFeeRate(w http.ResponseWriter, r *http.Request)
	GetInstruments(w http.ResponseWriter, r *http.Request)
}
