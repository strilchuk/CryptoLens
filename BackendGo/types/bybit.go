package types

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/models"
	"context"
	"github.com/shopspring/decimal"
	"net/http"
)

// BybitServiceInterface определяет интерфейс для сервиса Bybit
type BybitServiceInterface interface {
	GetWalletBalance(ctx context.Context, token string) (*bybit.BybitWalletBalance, error)
	GetFeeRate(ctx context.Context, token string, category string, symbol string, baseCoin string) (*bybit.BybitFeeRateResponse, error)
	GetInstruments(ctx context.Context, category string) ([]models.BybitInstrument, error)
	StartInstrumentsUpdate(ctx context.Context)
	StartWebSocket(ctx context.Context)
	StartPrivateWebSocket(ctx context.Context)
	GetStrategyManager() StrategyManagerInterface
	GetUserStrategyService() UserStrategyServiceInterface
}

// BybitHandlerInterface определяет интерфейс для обработчика Bybit
type BybitHandlerInterface interface {
	GetWalletBalance(w http.ResponseWriter, r *http.Request)
	GetFeeRate(w http.ResponseWriter, r *http.Request)
	GetInstruments(w http.ResponseWriter, r *http.Request)
}

// BybitWebSocketHandlerInterface определяет интерфейс для обработчика WebSocket сообщений
type BybitWebSocketHandlerInterface interface {
	HandleMessage(ctx context.Context, msg bybit.WebSocketMessage)
	HandlePrivateMessage(ctx context.Context, msg bybit.WebSocketMessage, userID string)
}

// StrategyManagerInterface определяет интерфейс для менеджера стратегий
type StrategyManagerInterface interface {
	AddStrategy(userID string, strategy Strategy)
	RemoveStrategy(userID string, strategy Strategy)
	UpdateUserInstruments(ctx context.Context, userID string) error
	HandleTicker(ctx context.Context, ticker bybit.TickerMessage)
	HandleOrderBook(ctx context.Context, orderBook bybit.OrderBookMessage)
	HandleTrade(ctx context.Context, trade bybit.TradeMessage)
	HandleOrder(ctx context.Context, order bybit.OrderMessage)
	HandleExecution(ctx context.Context, execution bybit.ExecutionMessage)
	HandleWallet(ctx context.Context, wallet bybit.WalletMessage)
	Start(ctx context.Context)
	Stop(ctx context.Context)
	GetStrategies(userID string) []Strategy
	GetStrategiesInfo() map[string][]string
	GetTicker(ctx context.Context, symbol string) (*bybit.TickerMessage, error)
	GetTickerHistory(ctx context.Context, symbol string, limit int64) ([]bybit.TickerMessage, error)
	GetOrderBook(ctx context.Context, symbol string) (*bybit.OrderBookMessage, error)
	GetOrderBookHistory(ctx context.Context, symbol string, limit int64) ([]bybit.OrderBookMessage, error)
	GetOrderBookSpread(ctx context.Context, symbol string) (decimal.Decimal, error)
	GetPublicTrades(ctx context.Context, symbol string, limit int64) ([]bybit.TradeMessage, error)
	GetPrivateOrder(ctx context.Context, userID, orderID string) (*bybit.OrderMessage, error)
	GetPrivateExecution(ctx context.Context, userID, execID string) (*bybit.ExecutionMessage, error)
	GetPrivateWallet(ctx context.Context, userID string) (*bybit.WalletMessage, error)
	GetWalletBalance(ctx context.Context, userID string) (*bybit.BybitWalletBalance, error)
}

// BybitAccountRepositoryInterface определяет методы для работы с аккаунтами Bybit
type BybitAccountRepositoryInterface interface {
	GetActiveAccountByUserID(ctx context.Context, userID string) (*bybit.BybitAccount, error)
	GetActiveAccounts(ctx context.Context) ([]bybit.BybitAccount, error)
	CreateAccount(ctx context.Context, userID string, apiKey, apiSecret, accountType string) (*bybit.BybitAccount, error)
	UpdateAccount(ctx context.Context, userID string, apiKey, apiSecret, accountType string, isActive bool) error
	DeleteAccount(ctx context.Context, userID string) error
}
