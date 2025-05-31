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
	manager       *StrategyManager
	minSpread     decimal.Decimal // Минимальный спред для размещения ордера
	minProfit     decimal.Decimal // Минимальная прибыль для продажи (в USDT)
	quantity      string          // Фиксированный объем ордера
	isBuying      bool            // Состояние: true - покупка, false - продажа
	buyPrice      decimal.Decimal // Цена покупки для расчета прибыли
	buyQty        decimal.Decimal // Объем покупки
	activeOrderID string          // ID активного ордера
	baseCoin      string          // Базовая монета (например, BTC для BTCUSDT)
}

// NewSpreadScalpingStrategy создает новую стратегию спред-скальпинга
func NewSpreadScalpingStrategy(userID, symbol string, manager *StrategyManager, minSpread, minProfit decimal.Decimal, quantity string) *SpreadScalpingStrategy {
	// Извлекаем базовую монету из символа (например, BTC из BTCUSDT)
	baseCoin := symbol[:len(symbol)-4] // Предполагаем, что котировочная монета - USDT
	return &SpreadScalpingStrategy{
		userID:    userID,
		symbol:    symbol,
		manager:   manager,
		minSpread: minSpread,
		minProfit: minProfit,
		quantity:  quantity,
		isBuying:  true,
		baseCoin:  baseCoin,
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
	spread, err := storages.GetOrderBookSpread(ctx, s.symbol)
	if err != nil {
		logger.LogError("SpreadScalping [%s] ошибка получения спреда: %v", s.userID, err)
		return
	}
	if spread.LessThan(s.minSpread) {
		return
	}

	// Проверяем баланс
	wallet, err := s.manager.GetWalletBalance(ctx, s.userID)
	if err != nil {
		logger.LogError("SpreadScalping [%s] ошибка получения кошелька: %v", s.userID, err)
		return
	}

	if s.isBuying {
		// Проверяем баланс USDT для покупки
		var freeBalance decimal.Decimal
		if len(wallet.List) > 0 {
			for _, coin := range wallet.List[0].Coins {
				if coin.Coin == "USDT" {
					freeBalance, _ = decimal.NewFromString(coin.WalletBalance)
					break
				}
			}
		}
		if freeBalance.LessThan(decimal.NewFromFloat(10)) {
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

		// Размещаем лимитный ордер на покупку чуть выше лучшего бида
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
	} else {
		// Проверяем баланс базовой монеты для продажи
		var freeBalance decimal.Decimal
		if len(wallet.List) > 0 {
			for _, coin := range wallet.List[0].Coins {
				if coin.Coin == s.baseCoin {
					freeBalance, _ = decimal.NewFromString(coin.WalletBalance)
					break
				}
			}
		}
		if freeBalance.LessThan(decimal.RequireFromString(s.quantity)) {
			logger.LogInfo("SpreadScalping [%s] недостаточный баланс %s: %s", s.userID, s.baseCoin, freeBalance.String())
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

		// Размещаем лимитный ордер на продажу чуть ниже лучшего аска
		if len(orderBook.Asks) > 0 {
			askPrice, _ := decimal.NewFromString(orderBook.Asks[0][0])
			sellPrice := askPrice.Sub(decimal.NewFromFloat(0.01))
			// Проверяем минимальную прибыль
			if sellPrice.Sub(s.buyPrice).Mul(s.buyQty).LessThan(s.minProfit) {
				logger.LogInfo("SpreadScalping [%s] потенциальная прибыль слишком мала: %s", s.userID, sellPrice.Sub(s.buyPrice).Mul(s.buyQty).String())
				return
			}
			priceStr := sellPrice.String()
			order, err := s.manager.CreateOrder(ctx, s.userID, s.symbol, "Sell", "Limit", s.quantity, &priceStr)
			if err != nil {
				logger.LogError("SpreadScalping [%s] ошибка создания ордера на продажу: %v", s.userID, err)
			} else {
				s.activeOrderID = order.OrderID
				logger.LogInfo("SpreadScalping [%s] создан ордер на продажу: %s по цене %s, ID: %s", s.userID, s.symbol, priceStr, order.OrderID)
			}
		}
	}
}

func (s *SpreadScalpingStrategy) OnTrade(ctx context.Context, trade bybit.TradeMessage) {
	// Игнорируем, так как стратегия ориентирована на книгу ордеров
}

func (s *SpreadScalpingStrategy) OnOrder(ctx context.Context, order bybit.OrderMessage) {
	logger.LogInfo("SpreadScalping [%s] обновление ордера: %s, статус: %s", s.userID, order.OrderID, order.OrderStatus)
	if order.OrderID == s.activeOrderID {
		if order.OrderStatus == "Filled" || order.OrderStatus == "Cancelled" {
			s.activeOrderID = ""
		}
	}
}

func (s *SpreadScalpingStrategy) OnExecution(ctx context.Context, execution bybit.ExecutionMessage) {
	logger.LogInfo("SpreadScalping [%s] исполнение: %s, цена: %s, объем: %s, сторона: %s",
		s.userID, execution.ExecID, execution.ExecPrice, execution.ExecQty, execution.Side)
	if execution.Symbol == s.symbol {
		if execution.Side == "Buy" && s.isBuying {
			// Фиксируем цену и объем покупки
			s.buyPrice, _ = decimal.NewFromString(execution.ExecPrice)
			s.buyQty, _ = decimal.NewFromString(execution.ExecQty)
			s.isBuying = false
			logger.LogInfo("SpreadScalping [%s] покупка исполнена, переходим к продаже: цена=%s, объем=%s",
				s.userID, s.buyPrice.String(), s.buyQty.String())
		} else if execution.Side == "Sell" && !s.isBuying {
			// Сбрасываем состояние после продажи
			s.isBuying = true
			s.buyPrice = decimal.Zero
			s.buyQty = decimal.Zero
			logger.LogInfo("SpreadScalping [%s] продажа исполнена, возвращаемся к покупке", s.userID)
		}
	}
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
