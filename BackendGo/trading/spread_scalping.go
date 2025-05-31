package trading

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/storages"
	"CryptoLens_Backend/types"
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

// SpreadScalpingStrategy реализует стратегию спред-скальпинга
type SpreadScalpingStrategy struct {
	userID         string
	symbol         string
	manager        *StrategyManager
	minSpread      decimal.Decimal                          // Минимальный спред
	minProfit      decimal.Decimal                          // Минимальная прибыль
	quantity       decimal.Decimal                          // Объем ордера
	isBuying       bool                                     // Состояние: true - покупка, false - продажа
	buyPrice       decimal.Decimal                          // Цена покупки
	buyQty         decimal.Decimal                          // Объем покупки
	activeOrderID  string                                   // ID активного ордера
	baseCoin       string                                   // Базовая монета (например, BTC)
	instrumentRepo types.BybitInstrumentRepositoryInterface // Репозиторий
}

// NewSpreadScalpingStrategy создает новую стратегию
func NewSpreadScalpingStrategy(
	userID, symbol string,
	manager *StrategyManager,
	instrumentRepo types.BybitInstrumentRepositoryInterface,
) *SpreadScalpingStrategy {
	baseCoin := symbol[:len(symbol)-4] // Например, BTC из BTCUSDT
	return &SpreadScalpingStrategy{
		userID:         userID,
		symbol:         symbol,
		manager:        manager,
		minSpread:      decimal.NewFromFloat(1),     // Начальное значение, обновится
		minProfit:      decimal.NewFromFloat(0.1),   // Начальное значение
		quantity:       decimal.NewFromFloat(0.001), // Начальное значение
		isBuying:       true,
		baseCoin:       baseCoin,
		instrumentRepo: instrumentRepo,
	}
}

// updateParameters обновляет параметры стратегии
func (s *SpreadScalpingStrategy) updateParameters(ctx context.Context) error {
	// Получаем данные инструмента
	instrument, err := s.instrumentRepo.GetBySymbol(ctx, s.symbol)
	if err != nil {
		return fmt.Errorf("failed to get instrument: %w", err)
	}

	// Получаем тикер для средней цены
	ticker, err := storages.GetTicker(ctx, s.symbol)
	if err != nil {
		return fmt.Errorf("failed to get ticker: %w", err)
	}
	lastPrice, _ := decimal.NewFromString(ticker.LastPrice)
	logger.LogDebug("SpreadScalping [%s] lastPrice: %s", s.userID, lastPrice.String())

	// Рассчитываем minSpread (0.02% от цены или минимум 1 USDT)
	calculatedSpread := lastPrice.Mul(decimal.NewFromFloat(0.0002))
	minSpread := decimal.NewFromFloat(1)
	if calculatedSpread.GreaterThan(minSpread) {
		s.minSpread = calculatedSpread
	} else {
		s.minSpread = minSpread
	}
	logger.LogDebug("SpreadScalping [%s] calculated minSpread: %s", s.userID, s.minSpread.String())

	// Рассчитываем quantity (10% баланса USDT, минимум minOrderQty)
	wallet, err := s.manager.GetWalletBalance(ctx, s.userID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	var usdtBalance decimal.Decimal
	for _, coin := range wallet.List[0].Coins {
		if coin.Coin == "USDT" {
			usdtBalance, _ = decimal.NewFromString(coin.USDValue)
			break
		}
	}
	logger.LogDebug("SpreadScalping [%s] usdtBalance: %s", s.userID, usdtBalance.String())

	targetValue := usdtBalance.Mul(decimal.NewFromFloat(0.1)) // 10% баланса
	quantity := targetValue.Div(lastPrice)                    // В BTC
	logger.LogDebug("SpreadScalping [%s] initial quantity: %s", s.userID, quantity.String())

	// Проверяем минимальный размер ордера
	if quantity.LessThan(instrument.MinOrderQty) {
		quantity = instrument.MinOrderQty
		logger.LogDebug("SpreadScalping [%s] adjusted quantity to minOrderQty: %s", s.userID, quantity.String())
	}

	// Проверяем максимальный размер ордера
	if quantity.GreaterThan(instrument.MaxOrderQty) {
		quantity = instrument.MaxOrderQty
		logger.LogDebug("SpreadScalping [%s] adjusted quantity to maxOrderQty: %s", s.userID, quantity.String())
	}

	// Проверяем минимальную стоимость ордера
	minOrderAmt := lastPrice.Mul(quantity)
	if minOrderAmt.LessThan(instrument.MinOrderAmt) {
		quantity = instrument.MinOrderAmt.Div(lastPrice)
		logger.LogDebug("SpreadScalping [%s] adjusted quantity for minOrderAmt: %s", s.userID, quantity.String())
	}

	// Проверяем максимальную стоимость ордера
	maxOrderAmt := lastPrice.Mul(quantity)
	if maxOrderAmt.GreaterThan(instrument.MaxOrderAmt) {
		quantity = instrument.MaxOrderAmt.Div(lastPrice)
		logger.LogDebug("SpreadScalping [%s] adjusted quantity for maxOrderAmt: %s", s.userID, quantity.String())
	}

	// Округляем до basePrecision
	precisionStr := instrument.BasePrecision.String()
	precisionPlaces := int32(0)
	if parts := strings.Split(precisionStr, "."); len(parts) == 2 {
		precisionPlaces = int32(len(parts[1]))
	}
	quantity = quantity.Round(precisionPlaces)

	logger.LogDebug("SpreadScalping [%s] rounded quantity to basePrecision: %s", s.userID, quantity.String())

	s.quantity = quantity

	// Рассчитываем minProfit (комиссии + маржа)
	feeRate := decimal.NewFromFloat(0.002) // 0.2%
	tradeValue := lastPrice.Mul(s.quantity)
	fees := tradeValue.Mul(feeRate)
	s.minProfit = fees.Add(decimal.NewFromFloat(0.1)) // Комиссии + 0.1 USDT
	logger.LogDebug("SpreadScalping [%s] calculated minProfit: %s (fees: %s)", s.userID, s.minProfit.String(), fees.String())

	logger.LogInfo("SpreadScalping [%s] обновлены параметры: minSpread=%s (%.4f%%), minProfit=%s, quantity=%s, lastPrice=%s, orderValue=%s USDT",
		s.userID,
		s.minSpread.String(),
		s.minSpread.Div(lastPrice).Mul(decimal.NewFromInt(100)).InexactFloat64(),
		s.minProfit.String(),
		s.quantity.String(),
		lastPrice.String(),
		lastPrice.Mul(quantity).String())
	return nil
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
			quantityStr := s.quantity.String()
			order, err := s.manager.CreateOrder(ctx, s.userID, s.symbol, "Buy", "Limit", quantityStr, &priceStr)
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
		if freeBalance.LessThan(s.quantity) {
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
			quantityStr := s.quantity.String()
			order, err := s.manager.CreateOrder(ctx, s.userID, s.symbol, "Sell", "Limit", quantityStr, &priceStr)
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

	// Создаем новый контекст для инициализации параметров
	initCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Обновляем параметры при старте
	if err := s.updateParameters(initCtx); err != nil {
		logger.LogError("SpreadScalping [%s] ошибка инициализации параметров: %v", s.userID, err)
	}

	// Периодическое обновление параметров
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Создаем новый контекст для каждого обновления
				updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				if err := s.updateParameters(updateCtx); err != nil {
					logger.LogError("SpreadScalping [%s] ошибка обновления параметров: %v", s.userID, err)
				}
				cancel()
			}
		}
	}()
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
