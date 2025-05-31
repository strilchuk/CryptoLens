package types

import (
	"CryptoLens_Backend/integration/bybit"
	"context"
)

// Strategy определяет интерфейс для торговых стратегий
type Strategy interface {
	// OnTicker вызывается при получении тикера
	OnTicker(ctx context.Context, ticker bybit.TickerMessage)
	// OnOrderBook вызывается при обновлении книги ордеров
	OnOrderBook(ctx context.Context, orderBook bybit.OrderBookMessage)
	// OnTrade вызывается при новой сделке
	OnTrade(ctx context.Context, trade bybit.TradeMessage)
	// OnOrder вызывается при обновлении ордера
	OnOrder(ctx context.Context, order bybit.OrderMessage)
	// OnExecution вызывается при исполнении ордера
	OnExecution(ctx context.Context, execution bybit.ExecutionMessage)
	// OnWallet вызывается при обновлении баланса
	OnWallet(ctx context.Context, wallet bybit.WalletMessage)
	// Start запускает стратегию
	Start(ctx context.Context)
	// Stop останавливает стратегию
	Stop(ctx context.Context)
} 