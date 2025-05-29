package handlers

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/logger"
	"context"
	"encoding/json"
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
	// Логируем входящее сообщение
	logger.LogInfo("Получено WebSocket сообщение: %s", msg.Topic)

	switch msg.Topic {
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

	case "trades":
		var tradeMsg bybit.TradeMessage
		if err := json.Unmarshal(msg.Data, &tradeMsg); err != nil {
			logger.LogError("Ошибка разбора сообщения о сделке: %v", err)
			return
		}
		h.handleTradeMessage(ctx, tradeMsg)

	default:
		logger.LogInfo("Неизвестный тип сообщения: %s", msg.Topic)
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
		msg.Symbol, msg.Price, msg.Size, msg.Side)
	// Здесь можно добавить дополнительную логику обработки сделок
} 