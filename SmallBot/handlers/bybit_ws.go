package handlers

import (
	"SmallBot/integration/bybit"
	"SmallBot/logger"
	"SmallBot/types"
	"context"
	"encoding/json"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

type BybitWebSocketHandler struct {
	msgChan chan *bybit.WebSocketMessage
	service types.BybitServiceInterface
}

func NewBybitWebSocketHandler(service types.BybitServiceInterface) *BybitWebSocketHandler {
	handler := &BybitWebSocketHandler{
		msgChan: make(chan *bybit.WebSocketMessage, 100),
		service: service,
	}
	go handler.processMessages()
	go handler.monitorChannel()
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
	logger.LogInfo("[TradeLogic] Текущая цена: %s для символа %s", currentPrice.String(), msg.Symbol)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка парсинга цены: %v", err)
		return
	}

	// Получаем баланс в USDT
	balance, err := h.service.GetUSDTBalance(ctx)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка получения баланса: %v", err)
		return
	}
	logger.LogInfo("[TradeLogic] Текущий баланс USDT: %s", balance.String())

	// Получаем волатильность
	volatility, err := h.service.GetVolatility(ctx, msg.Symbol)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка получения волатильности: %v", err)
		return
	}
	logger.LogInfo("[TradeLogic] Текущая волатильность для %s: %s", msg.Symbol, volatility.String())

	// Получаем комиссию
	fee, err := h.service.GetTradingFee(ctx, msg.Symbol)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка получения комиссии: %v", err)
		return
	}
	logger.LogInfo("[TradeLogic] Текущая комиссия для %s: %s", msg.Symbol, fee.String())

	// Параметры стратегии
	entryOffsetPercent := decimal.NewFromFloat(0.1) // 0.1%
	profitMultiplier := decimal.NewFromFloat(1.5)   // 1.5x волатильности
	orderSizePercent := decimal.NewFromFloat(50.0)  // 50% от баланса

	// Рассчитываем цены для ордеров
	buyPrice, sellPrice, err := h.service.CalculateOrderPrices(
		ctx,
		msg.Symbol,
		currentPrice,
		volatility,
		fee,
		entryOffsetPercent,
		profitMultiplier,
	)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка расчета цен: %v", err)
		return
	}
	logger.LogInfo("[TradeLogic] Рассчитанные цены: BuyPrice=%s, SellPrice=%s", buyPrice.String(), sellPrice.String())

	// Рассчитываем размер ордера
	orderSize, err := h.service.CalculateOrderSize(
		ctx,
		msg.Symbol,
		balance,
		orderSizePercent,
		currentPrice,
	)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка расчета размера ордера: %v", err)
		return
	}
	// Округляем размер ордера до 6 знаков после запятой
	orderSize = orderSize.Round(6)
	logger.LogInfo("[TradeLogic] Рассчитанный размер ордера: %s", orderSize.String())

	// Создаем ордер на покупку
	buyOrder, err := h.service.CreateLimitOrder(
		ctx,
		msg.Symbol,
		"Buy",
		orderSize.StringFixed(6),
		buyPrice.StringFixed(2),
	)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка создания ордера на покупку: %v", err)
		return
	}

	logger.LogInfo("[TradeLogic] Создан ордер на покупку: Symbol=%s, Price=%s, Size=%s, OrderID=%s",
		msg.Symbol, buyPrice.StringFixed(2), orderSize.StringFixed(6), buyOrder.OrderID)

	// Создаем ордер на продажу
	sellOrder, err := h.service.CreateLimitOrder(
		ctx,
		msg.Symbol,
		"Sell",
		orderSize.StringFixed(6),
		sellPrice.StringFixed(2),
	)
	if err != nil {
		logger.LogError("[TradeLogic] Ошибка создания ордера на продажу: %v", err)
		// Отменяем ордер на покупку
		_, cancelErr := h.service.CancelOrder(ctx, msg.Symbol, buyOrder.OrderID)
		if cancelErr != nil {
			logger.LogError("[TradeLogic] Ошибка отмены ордера на покупку: %v", cancelErr)
		}
		return
	}

	logger.LogInfo("[TradeLogic] Создан ордер на продажу: Symbol=%s, Price=%s, Size=%s, OrderID=%s",
		msg.Symbol, sellPrice.StringFixed(2), orderSize.StringFixed(6), sellOrder.OrderID)

	// Сохраняем информацию об активных ордерах
	h.service.SetLastOrderID(buyOrder.OrderID)
	h.service.SetSellOrderID(sellOrder.OrderID)
	h.service.SetOrderActive(true)
}

func (h *BybitWebSocketHandler) handleOrderMessage(ctx context.Context, msg bybit.OrderMessage) {
	jsonStr, _ := json.Marshal(msg)
	logger.LogDebug("handleOrderMessage: %s", string(jsonStr))

	logger.LogInfo("[TradeLogic] handleOrderMessage: %s", string(jsonStr))

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

		// Получаем волатильность
		volatility, err := h.service.GetVolatility(ctx, msg.Symbol)
		if err != nil {
			logger.LogError("[TradeLogic] Ошибка получения волатильности: %v", err)
			return
		}

		// Получаем комиссию
		fee, err := h.service.GetTradingFee(ctx, msg.Symbol)
		if err != nil {
			logger.LogError("[TradeLogic] Ошибка получения комиссии: %v", err)
			return
		}

		// Параметры стратегии
		entryOffsetPercent := decimal.NewFromFloat(0.1) // 0.1%
		profitMultiplier := decimal.NewFromFloat(1.5)   // 1.5x волатильности

		// Рассчитываем цены для нового ордера
		buyPrice, sellPrice, err := h.service.CalculateOrderPrices(
			ctx,
			msg.Symbol,
			currentPrice,
			volatility,
			fee,
			entryOffsetPercent,
			profitMultiplier,
		)
		if err != nil {
			logger.LogError("[TradeLogic] Ошибка расчета цен: %v", err)
			return
		}

		// Если исполнился ордер на покупку
		if msg.Side == "Buy" && msg.OrderID == h.service.GetLastOrderID() {
			// Округляем размер ордера до 6 знаков после запятой
			qty, err := decimal.NewFromString(msg.Qty)
			if err != nil {
				logger.LogError("[TradeLogic] Ошибка парсинга размера ордера: %v", err)
				return
			}
			qty = qty.Round(6)

			// Создаем новый ордер на продажу
			sellOrder, err := h.service.CreateLimitOrder(
				ctx,
				msg.Symbol,
				"Sell",
				qty.StringFixed(6),
				sellPrice.StringFixed(2),
			)
			if err != nil {
				logger.LogError("[TradeLogic] Ошибка создания ордера на продажу: %v", err)
				return
			}

			logger.LogInfo("[TradeLogic] Создан новый ордер на продажу: Symbol=%s, Price=%s, Size=%s, OrderID=%s",
				msg.Symbol, sellPrice.StringFixed(2), qty.StringFixed(6), sellOrder.OrderID)

			// Сохраняем ID нового ордера на продажу
			h.service.SetSellOrderID(sellOrder.OrderID)
		}
		// Если исполнился ордер на продажу
		if msg.Side == "Sell" && msg.OrderID == h.service.GetSellOrderID() {
			// Округляем размер ордера до 6 знаков после запятой
			qty, err := decimal.NewFromString(msg.Qty)
			if err != nil {
				logger.LogError("[TradeLogic] Ошибка парсинга размера ордера: %v", err)
				return
			}
			qty = qty.Round(6)

			// Создаем новый ордер на покупку
			buyOrder, err := h.service.CreateLimitOrder(
				ctx,
				msg.Symbol,
				"Buy",
				qty.StringFixed(7),
				buyPrice.StringFixed(2),
			)
			if err != nil {
				logger.LogError("[TradeLogic] Ошибка создания ордера на покупку: %v", err)
				return
			}

			logger.LogInfo("[TradeLogic] Создан новый ордер на покупку: Symbol=%s, Price=%s, Size=%s, OrderID=%s",
				msg.Symbol, buyPrice.StringFixed(2), qty.StringFixed(6), buyOrder.OrderID)

			// Сохраняем ID нового ордера на покупку
			h.service.SetLastOrderID(buyOrder.OrderID)
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
