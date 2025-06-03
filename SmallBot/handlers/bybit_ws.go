package handlers

import (
	"SmallBot/env"
	"SmallBot/integration/bybit"
	"SmallBot/logger"
	"SmallBot/types"
	"context"
	"encoding/json"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
)

const (
	EntryOffsetPercent = 0.05 // 0.05%
	ProfitMultiplier   = 1.5  // 1.5x волатильности
	OrderSizePercent   = 20.0 // 80% от баланса
	BuyOrderTimeout    = 1 * time.Minute
)

type BybitWebSocketHandler struct {
	msgChan        chan *bybit.WebSocketMessage
	service        types.BybitServiceInterface
	buyOrderTimers map[string]time.Time // orderID -> creation time
	mu             sync.Mutex
}

func NewBybitWebSocketHandler(service types.BybitServiceInterface) *BybitWebSocketHandler {
	handler := &BybitWebSocketHandler{
		msgChan:        make(chan *bybit.WebSocketMessage, 100),
		service:        service,
		buyOrderTimers: make(map[string]time.Time),
	}
	go handler.processMessages()
	go handler.monitorChannel()
	go handler.buyOrderTimeoutWatcher()
	return handler
}

func (h *BybitWebSocketHandler) processMessages() {
	for msg := range h.msgChan {
		ctx := context.Background()
		h.processMessage(ctx, msg)
	}
}

func (h *BybitWebSocketHandler) processMessage(ctx context.Context, msg *bybit.WebSocketMessage) {
	logger.LogDebug("Начало обработки сообщения: Topic=%s, Data=%s", msg.Topic, string(msg.Data))
	defer logger.LogDebug("Конец обработки сообщения: Topic=%s", msg.Topic)

	if msg.Topic == "" {
		var subResponse struct {
			Op      string   `json:"op"`
			Success bool     `json:"success"`
			Args    []string `json:"args"`
			RetMsg  string   `json:"ret_msg"`
		}
		if err := json.Unmarshal(msg.Data, &subResponse); err == nil && subResponse.Op == "subscribe" {
			if subResponse.Success {
				logger.LogInfo("Подписка подтверждена для каналов: %v", subResponse.Args)
			} else {
				logger.LogError("Ошибка подписки: %s", subResponse.RetMsg)
			}
		}
		return
	}

	topicParts := strings.Split(msg.Topic, ".")
	if len(topicParts) < 2 {
		logger.LogError("Неверный формат топика: %s", msg.Topic)
		return
	}

	messageType := topicParts[0]
	// Получаем последнюю часть топика как символ
	//symbol := topicParts[len(topicParts)-1]

	switch messageType {
	case "tickers":
		var tickerMsg bybit.TickerMessage
		if err := json.Unmarshal(msg.Data, &tickerMsg); err != nil {
			logger.LogError("Ошибка разбора сообщения тикера: %v", err)
			return
		}
		h.handleTickerMessage(ctx, tickerMsg)
	case "order.spot":
		var orderMsg bybit.OrderMessage
		if err := json.Unmarshal(msg.Data, &orderMsg); err != nil {
			logger.LogError("Ошибка разбора сообщения ордера: %v", err)
			return
		}
		h.handleOrderMessage(ctx, orderMsg)
	default:
		logger.LogInfo("Неизвестный тип сообщения: %s", messageType)
	}
}

func (h *BybitWebSocketHandler) HandleMessage(ctx context.Context, msg bybit.WebSocketMessage) {
	select {
	case h.msgChan <- &msg:
	default:
		h.msgChan <- &msg
	}
}

func (h *BybitWebSocketHandler) monitorChannel() {
	for {
		time.Sleep(time.Second)
		channelLen := len(h.msgChan)
		channelCap := cap(h.msgChan)
		fillPercentage := float64(channelLen) / float64(channelCap) * 100

		if fillPercentage >= 100 {
			logger.LogError("[CRITICAL] Канал сообщений переполнен! Заполнение: %.2f%%", fillPercentage)
		} else if fillPercentage >= 80 {
			logger.LogWarn("[WARNING] Канал сообщений почти заполнен! Заполнение: %.2f%%", fillPercentage)
		}
	}
}

// Универсальная функция для расчёта параметров ордера
func (h *BybitWebSocketHandler) prepareOrderParams(
	ctx context.Context,
	symbol string,
	currentPrice decimal.Decimal,
) (buyPrice, sellPrice, orderSize, fee, volatility, usdBalance, btcBalance decimal.Decimal, err error) {
	// Получаем волатильность
	volatility, err = h.service.GetVolatility(ctx, symbol)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка получения волатильности: %v", err)
		return
	}

	// Получаем комиссию
	fee, err = h.service.GetTradingFee(ctx, symbol)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка получения комиссии: %v", err)
		return
	}

	// Получаем баланс USDT
	usdBalance, err = h.service.GetUSDTBalance(ctx)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка получения баланса: %v", err)
		return
	}
	logger.LogInfo("[TradeLogic] Текущий баланс USD: %s", usdBalance.String())

	// Проверяем баланс BTC
	btcBalance, err = h.service.GetBTCBalance(ctx)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка получения баланса BTC: %v", err)
		return
	}
	logger.LogInfo("[TradeLogic] Текущий баланс BTC: %s", btcBalance.String())

	// Рассчитываем цены для ордеров
	buyPrice, sellPrice, err = h.service.CalculateOrderPrices(
		ctx,
		symbol,
		currentPrice,
		volatility,
		fee,
		decimal.NewFromFloat(EntryOffsetPercent),
		decimal.NewFromFloat(ProfitMultiplier),
	)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка расчета цен: %v", err)
		return
	}

	// Рассчитываем размер ордера
	orderSize, err = h.service.CalculateOrderSize(
		ctx,
		symbol,
		usdBalance,
		decimal.NewFromFloat(OrderSizePercent),
		currentPrice,
	)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка расчета размера ордера: %v", err)
		return
	}
	orderSize = orderSize.Round(6)

	logger.LogInfo("[TradeLogic] buyPrice=%s, sellPrice=%s, orderSize=%s, fee=%s, volatility=%s",
		buyPrice.String(), sellPrice.String(), orderSize.String(), fee.String(), volatility.String())

	return
}

func (h *BybitWebSocketHandler) handleTickerMessage(ctx context.Context, msg bybit.TickerMessage) {
	jsonStr, _ := json.Marshal(msg)
	logger.LogDebug("handleTickerMessage: %s", string(jsonStr))

	// Проверяем, нет ли активного ордера
	if h.service.IsOrderActive() {
		logger.LogDebug("[TradeLogic] Есть активный ордер, пропускаем")
		return
	}

	// Получаем текущую цену
	currentPrice, err := decimal.NewFromString(msg.LastPrice)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка парсинга цены: %v", err)
		return
	}
	logger.LogInfo("[TradeLogic] Текущая цена: %s для символа %s", currentPrice.String(), msg.Symbol)

	// Получаем параметры для ордера
	buyPrice, sellPrice, orderSize, _, _, _, btcBalance, err := h.prepareOrderParams(ctx, msg.Symbol, currentPrice)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка подготовки параметров для ордера: %v", err)
		return
	}

	// Создаем ордер на покупку
	buyOrder, err := h.service.CreateLimitOrder(
		ctx,
		msg.Symbol,
		"Buy",
		orderSize.StringFixed(6),
		buyPrice.StringFixed(2),
	)

	if err == nil {
		h.service.SetBuyOrderID(buyOrder.OrderID)
		h.addBuyOrderTimer(buyOrder.OrderID)
	}

	// Проверяем, достаточно ли BTC для продажи
	if btcBalance.GreaterThanOrEqual(orderSize) {
		// Создаем ордер на продажу
		sellOrder, err := h.service.CreateLimitOrder(
			ctx,
			msg.Symbol,
			"Sell",
			orderSize.StringFixed(6),
			sellPrice.StringFixed(2),
		)

		if err == nil {
			h.service.SetSellOrderID(sellOrder.OrderID)
		}
	} else {
		logger.LogInfo("[TradeLogic] Недостаточно BTC для создания ордера на продажу: баланс=%s, требуется=%s. Продолжаем только с ордером на покупку",
			btcBalance.String(), orderSize.String())
	}

	h.service.SetOrderActive(true)
}

func (h *BybitWebSocketHandler) handleOrderMessage(ctx context.Context, msg bybit.OrderMessage) {
	jsonStr, _ := json.Marshal(msg)
	logger.LogDebug("handleOrderMessage: %s", string(jsonStr))

	// Если ордер исполнен
	if msg.OrderStatus == "Filled" {
		logger.LogInfo("[TradeLogic] Ордер исполнен: Symbol=%s, Side=%s, Price=%s, Size=%s, OrderID=%s",
			msg.Symbol, msg.Side, msg.Price, msg.Qty, msg.OrderID)

		// Получаем текущую цену
		currentPrice, err := decimal.NewFromString(msg.Price)
		if err != nil {
			logger.LogError("[TradeLogic] Ошибка парсинга цены: %v", err)
			return
		}

		// Получаем параметры для ордера
		buyPrice, sellPrice, _, _, _, usdBalance, btcBalance, err := h.prepareOrderParams(ctx, msg.Symbol, currentPrice)
		if err != nil {
			logger.LogError("[TradeLogic] Ошибка подготовки параметров для ордера: %v", err)
			return
		}

		// Если исполнился ордер на покупку
		if msg.Side == "Buy" && msg.OrderID == h.service.GetBuyOrderID() {
			// Отменяем ордер на продажу
			_, err = h.service.CancelOrder(ctx, msg.Symbol, h.service.GetSellOrderID())
			if err != nil {
				logger.LogError("[TradeLogic] Ошибка отмены ордера на продажу: %v", err)
			} else {
				logger.LogInfo("[TradeLogic] Ордер на продажу успешно отменен")
			}

			// Округляем размер ордера до 6 знаков после запятой
			qty, err := decimal.NewFromString(msg.Qty)
			if err != nil {
				logger.LogError("[TradeLogic] Ошибка парсинга размера ордера: %v", err)
				return
			}
			qty = qty.Round(6)

			// Проверяем, достаточно ли BTC для продажи
			if btcBalance.LessThan(qty) {
				logger.LogError("[TradeLogic] Недостаточно BTC для продажи: баланс=%s, требуется=%s",
					btcBalance.String(), qty.String())
				return
			}
			// Создаем новый ордер на продажу
			sellOrder, err := h.service.CreateLimitOrder(
				ctx,
				msg.Symbol,
				"Sell",
				qty.StringFixed(6),
				sellPrice.StringFixed(2),
			)
			if sellOrder != nil {
				h.service.SetSellOrderID(sellOrder.OrderID)
			}
		}
		// Если исполнился ордер на продажу
		if msg.Side == "Sell" && msg.OrderID == h.service.GetSellOrderID() {
			// Отменяем ордер на покупку
			_, err = h.service.CancelOrder(ctx, msg.Symbol, h.service.GetLastOrderID())
			if err != nil {
				logger.LogError("[TradeLogic] Ошибка отмены ордера на покупку: %v", err)
			} else {
				logger.LogInfo("[TradeLogic] Ордер на покупку успешно отменен")
			}

			// Округляем размер ордера до 6 знаков после запятой
			qty, err := decimal.NewFromString(msg.Qty)
			if err != nil {
				logger.LogError("[TradeLogic] Ошибка парсинга размера ордера: %v", err)
				return
			}
			qty = qty.Round(6)

			// Рассчитываем необходимую сумму USDT для покупки
			requiredUSDT := qty.Mul(buyPrice)

			// Проверяем, достаточно ли USDT для покупки
			if usdBalance.LessThan(requiredUSDT) {
				logger.LogError("[TradeLogic] Недостаточно USDT для покупки: баланс=%s, требуется=%s",
					usdBalance.String(), requiredUSDT.String())
				return
			}

			// Создаем новый ордер на покупку
			buyOrder, err := h.service.CreateLimitOrder(
				ctx,
				msg.Symbol,
				"Buy",
				qty.StringFixed(6),
				buyPrice.StringFixed(2),
			)
			if buyOrder != nil {
				h.service.SetLastOrderID(buyOrder.OrderID)
			}
		}
		h.removeBuyOrderTimer(msg.OrderID)
	}

	// Если ордер отменён
	if msg.OrderStatus == "Cancelled" {
		logger.LogInfo("[TradeLogic] Ордер отменён: Symbol=%s, Side=%s, OrderID=%s", msg.Symbol, msg.Side, msg.OrderID)

		if msg.Side == "Buy" {
			h.service.SetBuyOrderID("")
		} else {
			h.service.SetSellOrderID("")
		}

		// Проверяем, есть ли ещё активные ордера (buy/sell)
		buyID := h.service.GetBuyOrderID()
		sellID := h.service.GetSellOrderID()
		if (buyID == "") && (sellID == "") {
			h.service.SetOrderActive(false)
			logger.LogInfo("[TradeLogic] Нет активных ордеров, SetOrderActive(false)")
		} else {
			logger.LogInfo("[TradeLogic] Есть активные ордера, пропускаем, buyID=%s, sellID=%s", buyID, sellID)
		}
	}
}

func (h *BybitWebSocketHandler) HandlePrivateMessage(ctx context.Context, msg bybit.WebSocketMessage) {
	logger.LogDebug("Приватное WebSocket сообщение: Topic=%s, Data=%s", msg.Topic, string(msg.Data))

	switch msg.Topic {
	case "order.spot":
		var orders []bybit.OrderMessage
		if err := json.Unmarshal(msg.Data, &orders); err != nil {
			logger.LogError("Ошибка разбора сообщения ордера: %v", err)
			return
		}
		for _, order := range orders {
			logger.LogInfo("Ордер: Symbol=%s, OrderID=%s, Status=%s",
				order.Symbol, order.OrderID, order.OrderStatus)

			// Если ордер исполнен или отменён, обрабатываем его
			if order.OrderStatus == "Filled" || order.OrderStatus == "Cancelled" {
				h.handleOrderMessage(ctx, order)
			}
		}

	case "execution.spot":
		//case "execution.fast.spot":
		var executions []bybit.ExecutionMessage
		if err := json.Unmarshal(msg.Data, &executions); err != nil {
			logger.LogError("Ошибка разбора сообщения исполнения: %v", err)
			return
		}
		for _, exec := range executions {
			logger.LogInfo("Исполнение: Symbol=%s, ExecID=%s, Price=%s, Qty=%s",
				exec.Symbol, exec.ExecID, exec.ExecPrice, exec.ExecQty)
		}

	case "wallet":
		var wallets []bybit.WalletMessage
		if err := json.Unmarshal(msg.Data, &wallets); err != nil {
			logger.LogError("Ошибка разбора сообщения кошелька: %v", err)
			return
		}
		if len(wallets) == 0 {
			logger.LogError("Получен пустой массив сообщений о кошельке")
			return
		}
		wallet := wallets[0] // Берем первое сообщение
		for _, coin := range wallet.Coin {
			logger.LogInfo("Баланс: Coin=%s, WalletBalance=%s, Free=%s",
				coin.Coin, coin.WalletBalance, coin.Free)
		}

	default:
		logger.LogInfo("Неизвестный приватный топик: %s", msg.Topic)
	}
}

// Добавлять buy-ордер в таймер
func (h *BybitWebSocketHandler) addBuyOrderTimer(orderID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.buyOrderTimers[orderID] = time.Now()
}

// Удалять buy-ордер из таймера
func (h *BybitWebSocketHandler) removeBuyOrderTimer(orderID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.buyOrderTimers, orderID)
}

// Проверка buy-ордеров на таймаут
func (h *BybitWebSocketHandler) buyOrderTimeoutWatcher() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		time.Sleep(1 * time.Second)
		h.mu.Lock()
		for orderID, created := range h.buyOrderTimers {
			if time.Since(created) > BuyOrderTimeout && h.service.GetSellOrderID() == "" {
				// Отменяем buy-ордер
				ctx := context.Background()
				_, err := h.service.CancelOrder(ctx, env.GetSymbol(), orderID)
				if err != nil {
					logger.LogError("[TradeLogic] Не удалось отменить buy-ордер по таймауту: %v", err)
					continue
				}
				logger.LogInfo("[TradeLogic] Buy-ордер %s отменён по таймауту", orderID)

				// Удаляем старый orderID из таймера
				h.removeBuyOrderTimer(orderID)
				h.service.SetOrderActive(false)
			}
		}
		h.mu.Unlock()
		time.Sleep(10 * time.Second)
	}
}
