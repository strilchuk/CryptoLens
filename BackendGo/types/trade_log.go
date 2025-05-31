package types

import (
	"CryptoLens_Backend/integration/bybit"
	"context"
)

// TradeLogRepositoryInterface определяет методы для работы с логами торговли
type TradeLogRepositoryInterface interface {
	SaveExecution(ctx context.Context, userID string, exec bybit.ExecutionMessage) error
} 