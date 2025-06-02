package bybit

import (
	"context"
	"time"
)

// Client интерфейс для работы с API Bybit
type Client interface {
	// GetWalletBalance получает баланс кошелька
	GetWalletBalance(ctx context.Context) (*BybitWalletBalance, error)

	// GetInstruments получает список доступных для торговли пар
	GetInstruments(ctx context.Context, category string) (*BybitInstrumentsResponse, error)

	// GetTickers получает текущие котировки
	GetTickers(ctx context.Context, category string, symbol *string) (*BybitTickersResponse, error)

	// GetKlines получает исторические свечи
	GetKlines(
		ctx context.Context,
		category string,
		symbol string,
		interval string,
		limit int,
		start *time.Time,
		end *time.Time,
	) (*BybitKlinesResponse, error)

	// GetTrades получает исторические сделки
	GetTrades(
		ctx context.Context,
		category string,
		symbol string,
		limit int,
		orderID *string,
	) (*BybitTradesResponse, error)

	// CreateOrder создает ордер
	CreateOrder(
		ctx context.Context,
		symbol string,
		side string,
		orderType string,
		qty string,
		price *string,
		timeInForce string,
		orderLinkID *string,
	) (*BybitOrderResponse, error)

	// AmendOrder изменяет ордер
	AmendOrder(
		ctx context.Context,
		symbol string,
		orderID string,
		price *string,
		qty *string,
	) (*BybitOrderResponse, error)

	// CancelOrder отменяет ордер
	CancelOrder(
		ctx context.Context,
		symbol string,
		orderID string,
	) (*BybitOrderResponse, error)

	// CancelAllOrders отменяет все ордера
	CancelAllOrders(
		ctx context.Context,
		symbol string,
	) (*BybitOrderResponse, error)

	// GetOpenOrders получает открытые ордера
	GetOpenOrders(
		ctx context.Context,
		symbol string,
		orderID *string,
		limit int,
	) (*BybitOrderListResponse, error)

	// GetFeeRate получает ставки комиссии
	GetFeeRate(
		ctx context.Context,
		category string,
		symbol *string,
		baseCoin *string,
	) (*BybitFeeRateResponse, error)
}
