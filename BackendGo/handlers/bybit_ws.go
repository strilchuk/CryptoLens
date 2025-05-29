package handlers

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/logger"
	"context"
	"encoding/json"
	"strings"
)

// BybitWebSocketHandler обрабатывает WebSocket сообщения от Bybit
type BybitWebSocketHandler struct {
	// Здесь можно добавить зависимости, если они понадобятся
}

// NewBybitWebSocketHandler создает новый обработчик WebSocket сообщений
func NewBybitWebSocketHandler() *BybitWebSocketHandler {
	return &BybitWebSocketHandler{}
}

// HandleMessage обрабатывает входящие WebSocket сообщения
func (h *BybitWebSocketHandler) HandleMessage(ctx context.Context, msg bybit.WebSocketMessage) {
	// Логируем сырое сообщение
	//logger.LogDebug("WebSocket сообщение: Topic=%s, Data=%s", msg.Topic, string(msg.Data))

	if msg.Topic == "" {
		logger.LogDebug("Сообщение без топика: %s", string(msg.Data))
		// Проверяем подтверждение подписки
		var subResponse struct {
			Op      string   `json:"op"`
			Success bool     `json:"success"`
			Args    []string `json:"args"`
		}
		if err := json.Unmarshal(msg.Data, &subResponse); err == nil && subResponse.Op == "subscribe" {
			if subResponse.Success {
				logger.LogInfo("Подписка подтверждена для каналов: %v", subResponse.Args)
			} else {
				logger.LogError("Ошибка подписки: %s", string(msg.Data))
			}
		}
		return
	}

	// Определяем тип сообщения по топику
	topicParts := strings.Split(msg.Topic, ".")
	if len(topicParts) < 2 {
		logger.LogError("Неверный формат топика: %s", msg.Topic)
		return
	}

	messageType := topicParts[0]
	//symbol := topicParts[1]

	switch messageType {
	case "tickers":
		var tickerMsg bybit.TickerMessage
		if err := json.Unmarshal(msg.Data, &tickerMsg); err != nil {
			logger.LogError("Ошибка разбора сообщения тикера: %v", err)
			return
		}
		h.handleTickerMessage(ctx, tickerMsg)

	case "orderbook":
		var orderBookMsg bybit.OrderBookMessage
		if err := json.Unmarshal(msg.Data, &orderBookMsg); err != nil {
			logger.LogError("Ошибка разбора сообщения книги ордеров: %v", err)
			return
		}
		h.handleOrderBookMessage(ctx, orderBookMsg)

	case "publicTrade":
		var tradeMsgs []bybit.TradeMessage
		if err := json.Unmarshal(msg.Data, &tradeMsgs); err != nil {
			logger.LogError("Ошибка разбора сообщения о сделке: %v", err)
			return
		}
		for _, tradeMsg := range tradeMsgs {
			h.handleTradeMessage(ctx, tradeMsg)
		}

	default:
		logger.LogInfo("Неизвестный тип сообщения: %s", messageType)
	}
}

// handleTickerMessage обрабатывает сообщения тикера
func (h *BybitWebSocketHandler) handleTickerMessage(ctx context.Context, msg bybit.TickerMessage) {
	logger.LogInfo("Тикер %s: цена=%s, объем=%s",
		msg.Symbol, msg.LastPrice, msg.Volume24h)
	// Здесь можно добавить дополнительную логику обработки тикера
}

// handleOrderBookMessage обрабатывает сообщения книги ордеров
func (h *BybitWebSocketHandler) handleOrderBookMessage(ctx context.Context, msg bybit.OrderBookMessage) {
	logger.LogInfo("Книга ордеров %s: %d бидов, %d асков",
		msg.Symbol, len(msg.Bids), len(msg.Asks))
	// Здесь можно добавить дополнительную логику обработки книги ордеров
}

// handleTradeMessage обрабатывает сообщения о сделках
func (h *BybitWebSocketHandler) handleTradeMessage(ctx context.Context, msg bybit.TradeMessage) {
	logger.LogInfo("Сделка %s: цена=%s, объем=%s, сторона=%s",
		msg.Symbol, msg.Price, msg.Volume, msg.Side)
	// Здесь можно добавить дополнительную логику обработки сделок
}
