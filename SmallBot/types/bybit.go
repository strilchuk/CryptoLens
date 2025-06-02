package types

import (
	"SmallBot/integration/bybit"
	"context"
)

type BybitServiceInterface interface {
	GetWalletBalance(ctx context.Context, token string) (*bybit.BybitWalletBalance, error)
	GetFeeRate(ctx context.Context, token string, category string, symbol string, baseCoin string) (*bybit.BybitFeeRateResponse, error)
	StartWebSocket(ctx context.Context)
	StartPrivateWebSocket(ctx context.Context)
}

type BybitWebSocketHandlerInterface interface {
	HandleMessage(ctx context.Context, msg bybit.WebSocketMessage)
	HandlePrivateMessage(ctx context.Context, msg bybit.WebSocketMessage)
}
