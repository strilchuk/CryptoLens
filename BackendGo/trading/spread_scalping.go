package trading

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/storages"
	"CryptoLens_Backend/types"
	"context"
	"encoding/json"
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
	msgChan        chan interface{}                         // Канал для сообщений
	stopChan       chan struct{}                            // Канал для остановки
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
		msgChan:        make(chan interface{}, 1000), // Буфер на 1000 сообщений
		stopChan:       make(chan struct{}),
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
	logger.LogInfo("SpreadScalping [%s] lastPrice: %s", s.userID, lastPrice.String())

	// Рассчитываем minSpread (0.02% от цены или минимум 1 USDT)
	calculatedSpread := lastPrice.Mul(decimal.NewFromFloat(0.0002))
	minSpread := decimal.NewFromFloat(1)
	if calculatedSpread.GreaterThan(minSpread) {
		s.minSpread = calculatedSpread
	} else {
		s.minSpread = minSpread
	}
	logger.LogDebug("SpreadScalping [%s] рассчитанный minSpread: %s", s.userID, s.minSpread.String())

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
	logger.LogInfo("SpreadScalping [%s] usdtBalance: %s", s.userID, usdtBalance.String())

	targetValue := usdtBalance.Mul(decimal.NewFromFloat(0.1)) // 10% баланса
	quantity := targetValue.Div(lastPrice)                    // В BTC
	logger.LogDebug("SpreadScalping [%s] начальный объем (quantity): %s", s.userID, quantity.String())

	// Проверяем минимальный размер ордера
	if quantity.LessThan(instrument.MinOrderQty) {
		quantity = instrument.MinOrderQty
		logger.LogDebug("SpreadScalping [%s] объем (quantity) скорректирован до минимального (minOrderQty): %s", s.userID, quantity.String())
	}

	// Проверяем максимальный размер ордера
	if quantity.GreaterThan(instrument.MaxOrderQty) {
		quantity = instrument.MaxOrderQty
		logger.LogDebug("SpreadScalping [%s] объем (quantity) скорректирован до максимального (maxOrderQty): %s", s.userID, quantity.String())
	}

	// Проверяем минимальную стоимость ордера
	minOrderAmt := lastPrice.Mul(quantity)
	if minOrderAmt.LessThan(instrument.MinOrderAmt) {
		quantity = instrument.MinOrderAmt.Div(lastPrice)
		logger.LogDebug("SpreadScalping [%s] объем (quantity) скорректирован по минимальной стоимости (minOrderAmt): %s", s.userID, quantity.String())
	}

	// Проверяем максимальную стоимость ордера
	maxOrderAmt := lastPrice.Mul(quantity)
	if maxOrderAmt.GreaterThan(instrument.MaxOrderAmt) {
		quantity = instrument.MaxOrderAmt.Div(lastPrice)
		logger.LogDebug("SpreadScalping [%s] объем (quantity) скорректирован по максимальной стоимости (maxOrderAmt): %s", s.userID, quantity.String())
	}

	// Округляем до basePrecision
	precisionStr := instrument.BasePrecision.String()
	precisionPlaces := int32(0)
	if parts := strings.Split(precisionStr, "."); len(parts) == 2 {
		precisionPlaces = int32(len(parts[1]))
	}
	quantity = quantity.Round(precisionPlaces)

	logger.LogDebug("SpreadScalping [%s] объем (quantity) округлен до точности базовой монеты (basePrecision): %s", s.userID, quantity.String())

	s.quantity = quantity

	// Рассчитываем minProfit (комиссии + маржа)
	feeRate := decimal.NewFromFloat(0.002) // 0.2%
	tradeValue := lastPrice.Mul(s.quantity)
	fees := tradeValue.Mul(feeRate)
	s.minProfit = fees.Add(decimal.NewFromFloat(0.1)) // Комиссии + 0.1 USDT
	logger.LogDebug("SpreadScalping [%s] рассчитанная минимальная прибыль (minProfit): %s (комиссии (fees): %s)", s.userID, s.minProfit.String(), fees.String())

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

// Start запускает стратегию
func (s *SpreadScalpingStrategy) Start(ctx context.Context) {
	logger.LogInfo("SpreadScalping [%s] запущена для %s", s.userID, s.symbol)

	// Создаем новый контекст для инициализации параметров
	strategyCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обновляем параметры при старте
	if err := s.updateParameters(strategyCtx); err != nil {
		logger.LogError("SpreadScalping [%s] ошибка инициализации параметров: %v", s.userID, err)
	}

	// Запускаем обработчик сообщений
	go s.processMessages()

	// Периодическое обновление параметров
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopChan:
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

// Stop останавливает стратегию
func (s *SpreadScalpingStrategy) Stop(ctx context.Context) {
	close(s.stopChan)
	if s.activeOrderID != "" {
		if err := s.manager.CancelOrder(ctx, s.userID, s.symbol, s.activeOrderID); err != nil {
			logger.LogError("SpreadScalping [%s] ошибка отмены ордера %s при остановке: %v", s.userID, s.activeOrderID, err)
		}
		s.activeOrderID = ""
	}
	logger.LogInfo("SpreadScalping [%s] остановлена", s.userID)
}

// processMessages обрабатывает сообщения из канала
func (s *SpreadScalpingStrategy) processMessages() {
	for {
		select {
		case <-s.stopChan:
			return
		case msg := <-s.msgChan:
			ctx := context.Background()
			switch m := msg.(type) {
			case bybit.TickerMessage:
				logger.LogInfo("SpreadScalping [%s] получен тикер: %s, цена: %s", s.userID, m.Symbol, m.LastPrice)
			case bybit.OrderBookMessage:
				spread, err := storages.GetOrderBookSpread(ctx, s.symbol)
				if err != nil {
					logger.LogError("SpreadScalping [%s] ошибка получения спреда: %v", s.userID, err)
					continue
				}
				if spread.LessThan(s.minSpread) {
					continue
				}

				// Проверяем баланс
				wallet, err := s.manager.GetWalletBalance(ctx, s.userID)
				if err != nil {
					logger.LogError("SpreadScalping [%s] ошибка получения кошелька: %v", s.userID, err)
					continue
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
						logger.LogInfo("SpreadScalping [%s] недостаточно средств: %s USDT", s.userID, freeBalance.String())
						continue
					}

					// Отменяем существующий ордер, если есть
					if s.activeOrderID != "" {
						if err := s.manager.CancelOrder(ctx, s.userID, s.symbol, s.activeOrderID); err != nil {
							logger.LogError("SpreadScalping [%s] ошибка отмены ордера %s: %v", s.userID, s.activeOrderID, err)
						} else {
							logger.LogInfo("SpreadScalping [%s] ордер отменен: %s", s.userID, s.activeOrderID)
							s.activeOrderID = ""
						}
					}

					// Размещаем лимитный ордер на покупку чуть выше лучшего бида
					if len(m.Bids) > 0 {
						bidPrice, _ := decimal.NewFromString(m.Bids[0][0])
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
					} else {
						orderBookJSON, _ := json.MarshalIndent(m, "", "  ")
						logger.LogError("SpreadScalping [%s] книга ордеров пуста: %s", s.userID, string(orderBookJSON))
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
						logger.LogInfo("SpreadScalping [%s] недостаточно средств %s: %s", s.userID, s.baseCoin, freeBalance.String())
						continue
					}

					// Отменяем существующий ордер, если есть
					if s.activeOrderID != "" {
						if err := s.manager.CancelOrder(ctx, s.userID, s.symbol, s.activeOrderID); err != nil {
							logger.LogError("SpreadScalping [%s] ошибка отмены ордера %s: %v", s.userID, s.activeOrderID, err)
						} else {
							logger.LogInfo("SpreadScalping [%s] ордер отменен: %s", s.userID, s.activeOrderID)
							s.activeOrderID = ""
						}
					}

					// Размещаем лимитный ордер на продажу чуть ниже лучшего аска
					if len(m.Asks) > 0 {
						askPrice, _ := decimal.NewFromString(m.Asks[0][0])
						sellPrice := askPrice.Sub(decimal.NewFromFloat(0.01))
						// Проверяем минимальную прибыль
						if sellPrice.Sub(s.buyPrice).Mul(s.buyQty).LessThan(s.minProfit) {
							logger.LogInfo("SpreadScalping [%s] потенциальная прибыль слишком мала: %s", s.userID, sellPrice.Sub(s.buyPrice).Mul(s.buyQty).String())
							continue
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
			case bybit.TradeMessage:
				// Игнорируем, так как стратегия ориентирована на книгу ордеров
			case bybit.OrderMessage:
				logger.LogInfo("SpreadScalping [%s] обновление ордера: %s, статус: %s", s.userID, m.OrderID, m.OrderStatus)
				if m.OrderID == s.activeOrderID {
					if m.OrderStatus == "Filled" || m.OrderStatus == "Cancelled" {
						s.activeOrderID = ""
					}
				}
			case bybit.ExecutionMessage:
				logger.LogInfo("SpreadScalping [%s] исполнение: %s, цена: %s, объем: %s, сторона: %s",
					s.userID, m.ExecID, m.ExecPrice, m.ExecQty, m.Side)
				if m.Symbol == s.symbol {
					if m.Side == "Buy" && s.isBuying {
						// Фиксируем цену и объем покупки
						s.buyPrice, _ = decimal.NewFromString(m.ExecPrice)
						s.buyQty, _ = decimal.NewFromString(m.ExecQty)
						s.isBuying = false
						logger.LogInfo("SpreadScalping [%s] покупка исполнена, переходим к продаже: цена=%s, объем=%s",
							s.userID, s.buyPrice.String(), s.buyQty.String())
					} else if m.Side == "Sell" && !s.isBuying {
						// Сбрасываем состояние после продажи
						s.isBuying = true
						s.buyPrice = decimal.Zero
						s.buyQty = decimal.Zero
						logger.LogInfo("SpreadScalping [%s] продажа исполнена, возвращаемся к покупке", s.userID)
					}
				}
			case bybit.WalletMessage:
				logger.LogInfo("SpreadScalping [%s] обновление кошелька", s.userID)
			}
		}
	}
}

// OnTicker обрабатывает тикер
func (s *SpreadScalpingStrategy) OnTicker(ctx context.Context, ticker bybit.TickerMessage) {
	if ticker.Symbol != s.symbol {
		return
	}
	select {
	case s.msgChan <- ticker:
		logger.LogDebug("SpreadScalping [%s] тикер отправлен в канал: %s", s.userID, ticker.Symbol)
	default:
		logger.LogWarn("SpreadScalping [%s] канал переполнен, тикер отброшен: %s", s.userID, ticker.Symbol)
	}
}

// OnOrderBook обрабатывает книгу ордеров
func (s *SpreadScalpingStrategy) OnOrderBook(ctx context.Context, orderBook bybit.OrderBookMessage) {
	if orderBook.Symbol != s.symbol {
		return
	}
	select {
	case s.msgChan <- orderBook:
		logger.LogDebug("SpreadScalping [%s] книга ордеров отправлена в канал: %s", s.userID, orderBook.Symbol)
	default:
		logger.LogWarn("SpreadScalping [%s] канал переполнен, книга ордеров отброшена: %s", s.userID, orderBook.Symbol)
	}
}

// OnTrade обрабатывает сделку
func (s *SpreadScalpingStrategy) OnTrade(ctx context.Context, trade bybit.TradeMessage) {
	if trade.Symbol != s.symbol {
		return
	}
	select {
	case s.msgChan <- trade:
		logger.LogDebug("SpreadScalping [%s] сделка отправлена в канал: %s", s.userID, trade.Symbol)
	default:
		logger.LogWarn("SpreadScalping [%s] канал переполнен, сделка отброшена: %s", s.userID, trade.Symbol)
	}
}

// OnOrder обрабатывает ордер
func (s *SpreadScalpingStrategy) OnOrder(ctx context.Context, order bybit.OrderMessage) {
	if order.Symbol != s.symbol {
		return
	}
	select {
	case s.msgChan <- order:
		logger.LogDebug("SpreadScalping [%s] ордер отправлен в канал: %s", s.userID, order.Symbol)
	default:
		logger.LogWarn("SpreadScalping [%s] канал переполнен, ордер отброшен: %s", s.userID, order.Symbol)
	}
}

// OnExecution обрабатывает исполнение
func (s *SpreadScalpingStrategy) OnExecution(ctx context.Context, execution bybit.ExecutionMessage) {
	if execution.Symbol != s.symbol {
		return
	}
	select {
	case s.msgChan <- execution:
		logger.LogDebug("SpreadScalping [%s] исполнение отправлено в канал: %s", s.userID, execution.Symbol)
	default:
		logger.LogWarn("SpreadScalping [%s] канал переполнен, исполнение отброшено: %s", s.userID, execution.Symbol)
	}
}

// OnWallet обрабатывает обновление кошелька
func (s *SpreadScalpingStrategy) OnWallet(ctx context.Context, wallet bybit.WalletMessage) {
	select {
	case s.msgChan <- wallet:
		logger.LogDebug("SpreadScalping [%s] кошелек отправлен в канал", s.userID)
	default:
		logger.LogWarn("SpreadScalping [%s] канал переполнен, кошелек отброшен", s.userID)
	}
}
