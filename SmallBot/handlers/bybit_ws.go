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
	isActive := h.service.IsOrderActive()
	logger.LogDebug("[TradeLogic] Статус активного ордера: %v", isActive)
	if !isActive {
		price, err := decimal.NewFromString(msg.LastPrice)
		if err != nil {
			logger.LogError("[TradeLogic] Ошибка парсинга цены: %v", err)
			return
		}
		// Вычисляем цену на 10% ниже текущей
		orderPrice := price.Mul(decimal.NewFromFloat(0.9)).StringFixed(2)
		qty := "0.001" // Минимальный объем для BTCUSDT
		orderResp, err := h.service.CreateLimitOrder(ctx, msg.Symbol, "Buy", qty, orderPrice)
		if err != nil {
			logger.LogError("[TradeLogic] Ошибка создания ордера: %v", err)
			return
		}

		logger.LogInfo("[TradeLogic] Создан ордер: Symbol=%s, Price=%s, OrderID=%s",
			msg.Symbol, orderPrice, orderResp.OrderID)

		h.service.SetLastOrderID(orderResp.OrderID)
		h.service.SetOrderActive(true)

		// Запускаем горутину для отмены ордера через 2 минуты
		go func(orderID, symbol string) {
			time.Sleep(30 * time.Second)

			_, err := h.service.CancelOrder(ctx, symbol, orderID)
			if err != nil {
				logger.LogError("[TradeLogic] Ошибка отмены ордера: %v", err)
			} else {
				logger.LogInfo("[TradeLogic] Ордер отменён: Symbol=%s, OrderID=%s", symbol, orderID)
			}

			h.service.SetOrderActive(false)
		}(orderResp.OrderID, msg.Symbol)
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
