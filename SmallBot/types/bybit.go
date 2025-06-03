package types

import (
	"SmallBot/integration/bybit"
	"context"
	"github.com/shopspring/decimal"
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

	// Новые методы для стратегии
	GetUSDTBalance(ctx context.Context) (decimal.Decimal, error)
	GetBTCBalance(ctx context.Context) (decimal.Decimal, error)
	GetVolatility(ctx context.Context, symbol string) (decimal.Decimal, error)
	GetTradingFee(ctx context.Context, symbol string) (decimal.Decimal, error)
	CalculateOrderPrices(
		ctx context.Context,
		symbol string,
		currentPrice decimal.Decimal,
		volatility decimal.Decimal,
		fee decimal.Decimal,
		entryOffsetPercent decimal.Decimal,
		profitMultiplier decimal.Decimal,
	) (buyPrice, sellPrice decimal.Decimal, err error)
	CalculateOrderSize(
		ctx context.Context,
		symbol string,
		balance decimal.Decimal,
		percent decimal.Decimal,
		currentPrice decimal.Decimal,
	) (decimal.Decimal, error)

	SetSellOrderID(orderID string)
	SetBuyOrderID(orderID string)
	GetSellOrderID() string
	GetBuyOrderID() string
}

type BybitWebSocketHandlerInterface interface {
	HandleMessage(ctx context.Context, msg bybit.WebSocketMessage)
	HandlePrivateMessage(ctx context.Context, msg bybit.WebSocketMessage)
}
