package services

import (
	"SmallBot/env"
	"SmallBot/integration/bybit"
	"SmallBot/logger"
	"SmallBot/types"
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BybitService struct {
	bybitClient     bybit.Client
	wsClient        *bybit.WebSocketClient
	privateWsClient *bybit.WebSocketClient
	wsHandler       types.BybitWebSocketHandlerInterface
	wsMutex         sync.Mutex
	orderActive     bool
	lastOrderID     string
	sellOrderID     string
	buyOrderID      string
}

func NewBybitService(
	bybitClient bybit.Client,
	wsHandler types.BybitWebSocketHandlerInterface,
) *BybitService {
	recvWindow, _ := strconv.Atoi(env.GetBybitRecvWindow())
	wsURL := env.GetBybitWsUrl() + "/v5/public/spot"
	wsClient := bybit.NewWebSocketClient(wsURL, recvWindow, "", "")

	return &BybitService{
		bybitClient: bybitClient,
		wsClient:    wsClient,
		wsHandler:   wsHandler,
	}
}

func (s *BybitService) GetWalletBalance(ctx context.Context, token string) (*bybit.BybitWalletBalance, error) {
	balance, err := s.bybitClient.GetWalletBalance(ctx)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (s *BybitService) GetFeeRate(
	ctx context.Context,
	token string,
	category string,
	symbol string,
	baseCoin string,
) (*bybit.BybitFeeRateResponse, error) {

	// Получаем ставки комиссии через клиент Bybit
	var symbolPtr, baseCoinPtr *string
	if symbol != "" {
		symbolPtr = &symbol
	}
	if baseCoin != "" {
		baseCoinPtr = &baseCoin
	}

	feeRate, err := s.bybitClient.GetFeeRate(ctx, category, symbolPtr, baseCoinPtr)
	if err != nil {
		return nil, err
	}

	return feeRate, nil
}

// преобразует строку в decimal.Decimal
func parseDecimal(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}

// возвращает количество знаков после запятой
func getPrecision(s string) int {
	if s == "" {
		return 0
	}
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return 0
	}
	return len(parts[1])
}

// запускает WebSocket-соединение и подписку на каналы
func (s *BybitService) StartWebSocket(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				symbol := env.GetSymbol()
				publicChannels := []string{fmt.Sprintf("tickers.%s", symbol)}
				logger.LogInfo("Подписываемся на каналы для активных инструментов: %v", publicChannels)

				// Подключаемся к WebSocket
				if err := s.wsClient.Connect(ctx); err != nil {
					logger.LogError("Failed to connect to WebSocket: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Запускаем обработку сообщений
				s.wsClient.StartMessageHandler(context.Background(), s.wsHandler.HandleMessage)

				// Подписываемся на публичные каналы
				if err := s.wsClient.Subscribe(ctx, publicChannels); err != nil {
					logger.LogError("Failed to subscribe to public channels: %v", err)
					s.wsClient.Close()
					time.Sleep(5 * time.Second)
					continue
				}

				// Логируем успешную подписку
				logger.LogInfo("Успешно подписались на %d каналов для активных инструментов", len(publicChannels))

				// Ждем завершения контекста
				<-ctx.Done()
				return
			}
		}
	}()
}

// StartPrivateWebSocket запускает приватные WebSocket-соединения для активных аккаунтов
func (s *BybitService) StartPrivateWebSocket(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.closePrivateWebSockets()
				return
			default:
				s.wsMutex.Lock()

				// Проверяем, существует ли уже подключение
				if s.privateWsClient != nil {
					s.wsMutex.Unlock()
					time.Sleep(5 * time.Second)
					continue
				}

				// Создаем соединения
				privateWsURL := env.GetBybitWsUrl() + "/v5/private"
				recvWindow, _ := strconv.Atoi(env.GetBybitRecvWindow())

				wsClient := bybit.NewWebSocketClient(privateWsURL, recvWindow, env.GetBybitApiToken(), env.GetBybitApiSecret())

				// Подключаемся и подписываемся
				if err := wsClient.Connect(ctx); err != nil {
					logger.LogError("Failed to connect to private WebSocket: %v", err)
					s.wsMutex.Unlock()
					time.Sleep(5 * time.Second)
					continue
				}

				privateChannels := []string{
					"order.spot",
					"execution.spot",
					//"execution.fast.spot",
					"wallet",
				}

				wsClient.StartMessageHandler(ctx, func(ctx context.Context, msg bybit.WebSocketMessage) {
					s.wsHandler.HandlePrivateMessage(ctx, msg)
				})

				if err := wsClient.Subscribe(ctx, privateChannels); err != nil {
					logger.LogError("Failed to subscribe to private channels: %v", err)
					wsClient.Close()
					s.wsMutex.Unlock()
					time.Sleep(5 * time.Second)
					continue
				}

				// Сохраняем клиент
				s.privateWsClient = wsClient
				logger.LogInfo("Успешно подключились к приватному WebSocket")
				s.wsMutex.Unlock()

				// Ждем завершения контекста
				<-ctx.Done()
				return
			}
		}
	}()
}

// closePrivateWebSockets закрывает все приватные WebSocket-соединения
func (s *BybitService) closePrivateWebSockets() {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	if s.privateWsClient != nil {
		s.privateWsClient.Close()
		s.privateWsClient = nil
		logger.LogInfo("Закрыто приватное WebSocket-соединение")
	}
}

func (s *BybitService) CreateLimitOrder(ctx context.Context, symbol string, side string, qty string, price string) (*bybit.BybitOrderResponse, error) {
	order, err := s.bybitClient.CreateOrder(ctx, symbol, side, "Limit", qty, &price, "GTC", nil)

	var sSide string
	if side == "Buy" {
		sSide = "покупку"
	} else {
		sSide = "продажу"
	}

	if err != nil {
		logger.LogError("[TradeLogic] Ошибка создания ордера на %s: %v", sSide, err)
		return nil, err
	}

	logger.LogInfo("[TradeLogic] Создан ордер на %s: Symbol=%s, Price=%s, Size=%s, OrderID=%s",
		sSide, symbol, price, qty, order.OrderID)

	return order, err
}

func (s *BybitService) CancelOrder(ctx context.Context, symbol string, orderID string) (*bybit.BybitOrderResponse, error) {
	return s.bybitClient.CancelOrder(ctx, symbol, orderID)
}

func (s *BybitService) CancelAllOrders(ctx context.Context, symbol string) (*bybit.BybitOrderResponse, error) {
	return s.bybitClient.CancelAllOrders(ctx, symbol)
}

func (s *BybitService) IsOrderActive() bool {
	return s.orderActive
}

func (s *BybitService) SetOrderActive(active bool) {
	s.orderActive = active
}

func (s *BybitService) SetLastOrderID(orderID string) {
	s.lastOrderID = orderID
}

func (s *BybitService) GetLastOrderID() string {
	return s.lastOrderID
}

func (s *BybitService) SetWebSocketHandler(handler types.BybitWebSocketHandlerInterface) {
	s.wsHandler = handler
}

// Получает баланс в USDT
func (s *BybitService) GetUSDTBalance(ctx context.Context) (decimal.Decimal, error) {
	balance, err := s.GetWalletBalance(ctx, "")
	if err != nil {
		return decimal.Zero, fmt.Errorf("ошибка получения баланса: %w", err)
	}

	// Ищем USDT в балансе
	for _, coin := range balance.List[0].Coins {
		if coin.Coin == "USDT" {
			return parseDecimal(coin.WalletBalance), nil
		}
	}

	return decimal.Zero, fmt.Errorf("USDT не найден в балансе")
}

// Получает баланс в BTC
func (s *BybitService) GetBTCBalance(ctx context.Context) (decimal.Decimal, error) {
	balance, err := s.GetWalletBalance(ctx, "")
	if err != nil {
		return decimal.Zero, fmt.Errorf("ошибка получения баланса: %w", err)
	}

	// Ищем BTC в балансе
	for _, coin := range balance.List[0].Coins {
		if coin.Coin == "BTC" {
			return parseDecimal(coin.WalletBalance), nil
		}
	}

	return decimal.Zero, fmt.Errorf("BTC не найден в балансе")
}

// Получает волатильность за последний час
func (s *BybitService) GetVolatility(ctx context.Context, symbol string) (decimal.Decimal, error) {
	// Получаем свечи за последний час (60 минут)
	klines, err := s.bybitClient.GetKlines(ctx, "spot", symbol, "15", 4, nil, nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("ошибка получения свечей: %w", err)
	}

	// Проверяем достаточность данных
	if len(klines.List) < 4 {
		return decimal.Zero, fmt.Errorf("недостаточно данных для расчета волатильности: получено %d свечей, нужно минимум 4", len(klines.List))
	}

	// Собираем цены закрытия
	var prices []decimal.Decimal
	for i, kline := range klines.List {
		closePrice, err := decimal.NewFromString(kline[4]) // Используем цену закрытия
		if err != nil {
			return decimal.Zero, fmt.Errorf("ошибка парсинга цены закрытия для свечи %d: %w", i, err)
		}
		prices = append(prices, closePrice)
	}

	// Рассчитываем среднюю цену
	var sum decimal.Decimal
	for _, price := range prices {
		sum = sum.Add(price)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(prices))))

	// Рассчитываем сумму квадратов отклонений
	var squaredDiffs decimal.Decimal
	for _, price := range prices {
		diff := price.Sub(mean)
		squaredDiffs = squaredDiffs.Add(diff.Mul(diff))
	}

	// Рассчитываем выборочную дисперсию (N-1) и стандартное отклонение
	variance := squaredDiffs.Div(decimal.NewFromInt(int64(len(prices) - 1)))
	volatility := decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))

	// Рассчитываем относительную волатильность в процентах
	relativeVolatility := volatility.Div(mean).Mul(decimal.NewFromInt(100))

	return relativeVolatility, nil
}

// Получает комиссию для торговой пары
func (s *BybitService) GetTradingFee(ctx context.Context, symbol string) (decimal.Decimal, error) {
	feeRate, err := s.GetFeeRate(ctx, "", "spot", symbol, "")
	if err != nil {
		return decimal.Zero, fmt.Errorf("ошибка получения комиссии: %w", err)
	}

	// Берем максимальную комиссию из maker и taker
	makerFee := parseDecimal(feeRate.List[0].MakerFeeRate)
	takerFee := parseDecimal(feeRate.List[0].TakerFeeRate)
	if makerFee.GreaterThan(takerFee) {
		return makerFee, nil
	}
	return takerFee, nil
}

// Рассчитывает цены для ордеров на основе волатильности и комиссии
func (s *BybitService) CalculateOrderPrices(
	ctx context.Context,
	symbol string,
	currentPrice decimal.Decimal,
	volatility decimal.Decimal,
	fee decimal.Decimal,
	entryOffsetPercent decimal.Decimal,
	profitMultiplier decimal.Decimal,
) (buyPrice, sellPrice decimal.Decimal, err error) {
	// Рассчитываем смещение для входа
	entryOffset := currentPrice.Mul(entryOffsetPercent).Div(decimal.NewFromInt(100))
	buyPrice = currentPrice.Sub(entryOffset)

	// Рассчитываем целевую прибыль
	// Используем абсолютную волатильность (в USDT)
	volatilityUSDT := currentPrice.Mul(volatility).Div(decimal.NewFromInt(100))
	profitTarget := volatilityUSDT.Mul(profitMultiplier)

	// Добавляем комиссию к целевой прибыли
	totalProfit := profitTarget.Add(fee.Mul(decimal.NewFromInt(2))) // Умножаем на 2, так как комиссия берется дважды
	sellPrice = buyPrice.Add(totalProfit)

	// Округляем цены до 2 знаков после запятой
	buyPrice = buyPrice.Round(1)
	sellPrice = sellPrice.Round(1)

	return buyPrice, sellPrice, nil
}

// Рассчитывает размер ордера на основе баланса
func (s *BybitService) CalculateOrderSize(
	ctx context.Context,
	symbol string,
	balance decimal.Decimal,
	percent decimal.Decimal,
	currentPrice decimal.Decimal,
) (decimal.Decimal, error) {
	// Рассчитываем размер ордера в USDT
	orderSizeUSDT := balance.Mul(percent).Div(decimal.NewFromInt(100))

	// Рассчитываем размер ордера в BTC
	orderSize := orderSizeUSDT.Div(currentPrice)

	// Проверяем минимальный размер ордера
	minOrderSize := decimal.NewFromFloat(0.0001) // Минимальный размер ордера для BTC
	if orderSize.LessThan(minOrderSize) {
		orderSize = minOrderSize
	}

	return orderSize, nil
}

func (s *BybitService) SetSellOrderID(orderID string) {
	s.sellOrderID = orderID
}

func (s *BybitService) GetSellOrderID() string {
	return s.sellOrderID
}

func (s *BybitService) SetBuyOrderID(orderID string) {
	s.buyOrderID = orderID
}

func (s *BybitService) GetBuyOrderID() string {
	return s.buyOrderID
}

func (s *BybitService) IsOrderExists(ctx context.Context, orderID string) (bool, error) {
	if orderID == "" {
		return false, nil
	}

	orderInfo, err := s.bybitClient.GetOrderInfo(ctx, orderID)
	if err != nil {
		return false, fmt.Errorf("ошибка получения информации об ордере: %w", err)
	}

	// Проверяем, что есть хотя бы один ордер в списке
	if len(orderInfo.List) == 0 {
		return false, nil
	}

	// Проверяем статус ордера
	order := orderInfo.List[0]
	return order.OrderStatus == "New", nil
}
