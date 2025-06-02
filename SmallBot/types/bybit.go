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
	CreateLimitOrder(ctx context.Context, symbol string, side string, qty string, price string) (*bybit.BybitOrderResponse, error)
	CancelOrder(ctx context.Context, symbol string, orderID string) (*bybit.BybitOrderResponse, error)
	CancelAllOrders(ctx context.Context, symbol string) (*bybit.BybitOrderResponse, error)
	IsOrderActive() bool
	SetOrderActive(active bool)
	SetLastOrderID(orderID string)
	GetLastOrderID() string
	SetWebSocketHandler(handler BybitWebSocketHandlerInterface)
}

type BybitWebSocketHandlerInterface interface {
	HandleMessage(ctx context.Context, msg bybit.WebSocketMessage)
	HandlePrivateMessage(ctx context.Context, msg bybit.WebSocketMessage)
}
