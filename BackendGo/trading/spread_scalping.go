package trading

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/storages"
	"context"
	"github.com/shopspring/decimal"
)

// SpreadScalpingStrategy реализует стратегию спред-скальпинга
type SpreadScalpingStrategy struct {
	userID        string
	symbol        string
	bybitClient   bybit.Client
	manager       *StrategyManager
	minSpread     decimal.Decimal // Минимальный спред для размещения ордера
	quantity      string          // Фиксированный объем ордера
	activeOrderID string
}

// NewSpreadScalpingStrategy создает новую стратегию спред-скальпинга
func NewSpreadScalpingStrategy(userID, symbol string, client bybit.Client, manager *StrategyManager, minSpread decimal.Decimal, quantity string) *SpreadScalpingStrategy {
	return &SpreadScalpingStrategy{
		userID:      userID,
		symbol:      symbol,
		bybitClient: client,
		manager:     manager,
		minSpread:   minSpread,
		quantity:    quantity,
	}
}

func (s *SpreadScalpingStrategy) OnTicker(ctx context.Context, ticker bybit.TickerMessage) {
	if ticker.Symbol != s.symbol {
		return
	}
	logger.LogInfo("SpreadScalping [%s] получил тикер: %s, цена: %s", s.userID, ticker.Symbol, ticker.LastPrice)
}

func (s *SpreadScalpingStrategy) OnOrderBook(ctx context.Context, orderBook bybit.OrderBookMessage) {
	if orderBook.Symbol != s.symbol {
		return
	}
	// Проверяем спред из Redis
	spread, err := storages.GetOrderBookSpread(ctx, s.symbol)
	if err != nil {
		logger.LogError("SpreadScalping [%s] ошибка получения спреда: %v", s.userID, err)
		return
	}
	if spread.LessThan(s.minSpread) {
		return
	}

	// Проверяем баланс
	wallet, err := storages.GetPrivateWallet(ctx, s.userID)
	if err != nil {
		logger.LogError("SpreadScalping [%s] ошибка получения кошелька: %v", s.userID, err)
		return
	}
	var freeBalance decimal.Decimal
	for _, coin := range wallet.Coin {
		if coin.Coin == "USDT" {
			freeBalance, _ = decimal.NewFromString(coin.Free)
			break
		}
	}
	if freeBalance.LessThan(decimal.NewFromFloat(10)) { // Минимальный баланс 10 USDT
		logger.LogInfo("SpreadScalping [%s] недостаточный баланс: %s USDT", s.userID, freeBalance.String())
		return
	}

	// Отменяем существующий ордер, если есть
	if s.activeOrderID != "" {
		if err := s.manager.CancelOrder(ctx, s.userID, s.symbol, s.activeOrderID); err != nil {
			logger.LogError("SpreadScalping [%s] ошибка отмены ордера %s: %v", s.userID, s.activeOrderID, err)
		} else {
			logger.LogInfo("SpreadScalping [%s] отменен ордер: %s", s.userID, s.activeOrderID)
			s.activeOrderID = ""
		}
	}

	// Размещаем новый лимитный ордер на покупку чуть выше лучшего бида
	if len(orderBook.Bids) > 0 {
		bidPrice, _ := decimal.NewFromString(orderBook.Bids[0][0])
		buyPrice := bidPrice.Add(decimal.NewFromFloat(0.01))
		priceStr := buyPrice.String()
		order, err := s.manager.CreateOrder(ctx, s.userID, s.symbol, "Buy", "Limit", s.quantity, &priceStr)
		if err != nil {
			logger.LogError("SpreadScalping [%s] ошибка создания ордера на покупку: %v", s.userID, err)
		} else {
			s.activeOrderID = order.OrderID
			logger.LogInfo("SpreadScalping [%s] создан ордер на покупку: %s по цене %s, ID: %s", s.userID, s.symbol, priceStr, order.OrderID)
		}
	}
}

func (s *SpreadScalpingStrategy) OnTrade(ctx context.Context, trade bybit.TradeMessage) {
	// Игнорируем, так как стратегия ориентирована на книгу ордеров
}

func (s *SpreadScalpingStrategy) OnOrder(ctx context.Context, order bybit.OrderMessage) {
	logger.LogInfo("SpreadScalping [%s] обновление ордера: %s, статус: %s", s.userID, order.OrderID, order.OrderStatus)
	if order.OrderID == s.activeOrderID && (order.OrderStatus == "Filled" || order.OrderStatus == "Cancelled") {
		s.activeOrderID = ""
	}
}

func (s *SpreadScalpingStrategy) OnExecution(ctx context.Context, execution bybit.ExecutionMessage) {
	logger.LogInfo("SpreadScalping [%s] исполнение: %s, цена: %s, объем: %s", s.userID, execution.ExecID, execution.ExecPrice, execution.ExecQty)
}

func (s *SpreadScalpingStrategy) OnWallet(ctx context.Context, wallet bybit.WalletMessage) {
	logger.LogInfo("SpreadScalping [%s] обновление кошелька", s.userID)
}

func (s *SpreadScalpingStrategy) Start(ctx context.Context) {
	logger.LogInfo("SpreadScalping [%s] запущена для %s", s.userID, s.symbol)
}

func (s *SpreadScalpingStrategy) Stop(ctx context.Context) {
	if s.activeOrderID != "" {
		if err := s.manager.CancelOrder(ctx, s.userID, s.symbol, s.activeOrderID); err != nil {
			logger.LogError("SpreadScalping [%s] ошибка отмены ордера %s при остановке: %v", s.userID, s.activeOrderID, err)
		}
		s.activeOrderID = ""
	}
	logger.LogInfo("SpreadScalping [%s] остановлена", s.userID)
}
