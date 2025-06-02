package handlers

import (
	"SmallBot/integration/bybit"
	"SmallBot/logger"
	"context"
	"encoding/json"
	"strings"
)

type BybitWebSocketHandler struct {
	msgChan chan *bybit.WebSocketMessage
}

func NewBybitWebSocketHandler() *BybitWebSocketHandler {
	handler := &BybitWebSocketHandler{
		msgChan: make(chan *bybit.WebSocketMessage, 1000), // Буфер на 1000 сообщений
	}
	go handler.processMessages()
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
	logger.LogDebug("Получено сообщение: Topic=%s", msg.Topic)
	select {
	case h.msgChan <- &msg: // Отправка в канал без блокировки
		logger.LogDebug("Сообщение отправлено в канал: Topic=%s", msg.Topic)
	default:
		logger.LogWarn("Канал переполнен, сообщение отброшено: Topic=%s", msg.Topic)
	}
}

func (h *BybitWebSocketHandler) handleTickerMessage(ctx context.Context, msg bybit.TickerMessage) {
	//logger.LogInfo("Тикер %s: цена=%s, объем=%s",
	//	msg.Symbol, msg.LastPrice, msg.Volume24h)
	// Здесь можно добавить дополнительную логику обработки тикера
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
