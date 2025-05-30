package trading

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/logger"
	"context"
)

// TestStrategy - тестовая стратегия для проверки работы системы
type TestStrategy struct {
	userID string
}

// NewTestStrategy создает новую тестовую стратегию
func NewTestStrategy(userID string) *TestStrategy {
	return &TestStrategy{
		userID: userID,
	}
}

// OnTicker обрабатывает обновление тикера
func (s *TestStrategy) OnTicker(ctx context.Context, ticker bybit.TickerMessage) {
	logger.LogInfo("TestStrategy [%s] получил тикер: %s, цена: %s", s.userID, ticker.Symbol, ticker.LastPrice)
}

// OnOrderBook обрабатывает обновление стакана
func (s *TestStrategy) OnOrderBook(ctx context.Context, orderBook bybit.OrderBookMessage) {
	logger.LogInfo("TestStrategy [%s] получил книгу ордеров: %s, биды: %d, аски: %d", 
		s.userID, orderBook.Symbol, len(orderBook.Bids), len(orderBook.Asks))
}

// OnTrade обрабатывает публичную сделку
func (s *TestStrategy) OnTrade(ctx context.Context, trade bybit.TradeMessage) {
	logger.LogInfo("TestStrategy [%s] получил сделку: %s, цена: %s, объем: %s, сторона: %s", 
		s.userID, trade.Symbol, trade.Price, trade.Volume, trade.Side)
}

// OnOrder обрабатывает обновление ордера
func (s *TestStrategy) OnOrder(ctx context.Context, order bybit.OrderMessage) {
	logger.LogInfo("TestStrategy [%s] получил ордер: %s, статус: %s", 
		s.userID, order.OrderID, order.OrderStatus)
}

// OnExecution обрабатывает исполнение ордера
func (s *TestStrategy) OnExecution(ctx context.Context, execution bybit.ExecutionMessage) {
	logger.LogInfo("TestStrategy [%s] получил исполнение: %s, цена: %s, объем: %s", 
		s.userID, execution.ExecID, execution.ExecPrice, execution.ExecQty)
}

// OnWallet обрабатывает обновление кошелька
func (s *TestStrategy) OnWallet(ctx context.Context, wallet bybit.WalletMessage) {
	logger.LogInfo("TestStrategy [%s] получил обновление кошелька", s.userID)
}

// Start запускает стратегию
func (s *TestStrategy) Start(ctx context.Context) {
	logger.LogInfo("TestStrategy [%s] запущена", s.userID)
}

// Stop останавливает стратегию
func (s *TestStrategy) Stop(ctx context.Context) {
	logger.LogInfo("TestStrategy [%s] остановлена", s.userID)
} 