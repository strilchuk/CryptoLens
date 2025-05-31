package handlers

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/storages"
	"CryptoLens_Backend/types"
	"context"
	"encoding/json"
	"github.com/shopspring/decimal"
	"strings"
)

// BybitWebSocketHandler обрабатывает WebSocket сообщения от Bybit
type BybitWebSocketHandler struct {
	strategyManager types.StrategyManagerInterface
	tradeLogRepo    types.TradeLogRepositoryInterface
}

// NewBybitWebSocketHandler создает новый обработчик WebSocket сообщений
func NewBybitWebSocketHandler(strategyManager types.StrategyManagerInterface, tradeLogRepo types.TradeLogRepositoryInterface) *BybitWebSocketHandler {
	return &BybitWebSocketHandler{
		strategyManager: strategyManager,
		tradeLogRepo:    tradeLogRepo,
	}
}

// HandleMessage обрабатывает входящие WebSocket сообщения
func (h *BybitWebSocketHandler) HandleMessage(ctx context.Context, msg bybit.WebSocketMessage) {
	logger.LogDebug("WebSocket сообщение: Topic=%s, Data=%s", msg.Topic, string(msg.Data))

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
	symbol := topicParts[len(topicParts)-1]

	switch messageType {
	case "tickers":
		var tickerMsg bybit.TickerMessage
		if err := json.Unmarshal(msg.Data, &tickerMsg); err != nil {
			logger.LogError("Ошибка разбора сообщения тикера: %v", err)
			return
		}
		if err := storages.SaveTicker(ctx, symbol, tickerMsg); err != nil {
			logger.LogError("Ошибка сохранения тикера: %v", err)
		}
		if err := storages.SaveTickerHistory(ctx, symbol, tickerMsg); err != nil {
			logger.LogError("Ошибка сохранения истории тикера: %v", err)
		}
		h.handleTickerMessage(ctx, tickerMsg)

	case "orderbook":
		var orderBookMsg bybit.OrderBookMessage
		if err := json.Unmarshal(msg.Data, &orderBookMsg); err != nil {
			logger.LogError("Ошибка разбора сообщения книги ордеров: %v", err)
			return
		}
		if err := storages.SaveOrderBook(ctx, symbol, orderBookMsg); err != nil {
			logger.LogError("Ошибка сохранения книги ордеров: %v", err)
		}
		if err := storages.SaveOrderBookHistory(ctx, symbol, orderBookMsg); err != nil {
			logger.LogError("Ошибка сохранения истории книги ордеров: %v", err)
		}

		// Вычисляем и сохраняем спред
		if len(orderBookMsg.Bids) > 0 && len(orderBookMsg.Asks) > 0 {
			bestBid, _ := decimal.NewFromString(orderBookMsg.Bids[0][0])
			bestAsk, _ := decimal.NewFromString(orderBookMsg.Asks[0][0])
			spread := bestAsk.Sub(bestBid)
			if err := storages.SaveOrderBookSpread(ctx, symbol, spread); err != nil {
				logger.LogError("Ошибка сохранения спреда: %v", err)
			}
		}

		h.handleOrderBookMessage(ctx, orderBookMsg)
		h.strategyManager.HandleOrderBook(ctx, orderBookMsg)

	case "publicTrade":
		var trades []bybit.TradeMessage
		if err := json.Unmarshal(msg.Data, &trades); err != nil {
			logger.LogError("Ошибка разбора сообщения о сделке: %v", err)
			return
		}
		for _, trade := range trades {
			if err := storages.SavePublicTrade(ctx, symbol, trade); err != nil {
				logger.LogError("Ошибка сохранения сделки: %v", err)
			}
			h.handleTradeMessage(ctx, trade)
			h.strategyManager.HandleTrade(ctx, trade)
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

// HandlePrivateMessage обрабатывает приватные WebSocket сообщения
func (h *BybitWebSocketHandler) HandlePrivateMessage(ctx context.Context, msg bybit.WebSocketMessage, userID string) {
	logger.LogDebug("Приватное WebSocket сообщение: Topic=%s, Data=%s", msg.Topic, string(msg.Data))

	switch msg.Topic {
	case "order.spot":
		var orders []bybit.OrderMessage
		if err := json.Unmarshal(msg.Data, &orders); err != nil {
			logger.LogError("Ошибка разбора сообщения ордера: %v", err)
			return
		}
		for _, order := range orders {
			if err := storages.SavePrivateOrder(ctx, userID, order.OrderID, order); err != nil {
				logger.LogError("Ошибка сохранения ордера: %v", err)
			}
			logger.LogInfo("Ордер: UserID=%s, Symbol=%s, OrderID=%s, Status=%s",
				userID, order.Symbol, order.OrderID, order.OrderStatus)
		}

	case "execution.spot":
	//case "execution.fast.spot":
		var executions []bybit.ExecutionMessage
		if err := json.Unmarshal(msg.Data, &executions); err != nil {
			logger.LogError("Ошибка разбора сообщения исполнения: %v", err)
			return
		}
		for _, exec := range executions {
			if err := storages.SavePrivateExecution(ctx, userID, exec.ExecID, exec); err != nil {
				logger.LogError("Ошибка сохранения исполнения: %v", err)
			}
			if err := h.tradeLogRepo.SaveExecution(ctx, userID, exec); err != nil {
				logger.LogError("Ошибка сохранения исполнения в trade_logs: %v", err)
			}
			logger.LogInfo("Исполнение: UserID=%s, Symbol=%s, ExecID=%s, Price=%s, Qty=%s",
				userID, exec.Symbol, exec.ExecID, exec.ExecPrice, exec.ExecQty)
		}

	case "wallet":
		var wallet bybit.WalletMessage
		if err := json.Unmarshal(msg.Data, &wallet); err != nil {
			logger.LogError("Ошибка разбора сообщения кошелька: %v", err)
			return
		}
		if err := storages.SavePrivateWallet(ctx, userID, wallet); err != nil {
			logger.LogError("Ошибка сохранения кошелька: %v", err)
		}
		for _, coin := range wallet.Coin {
			logger.LogInfo("Баланс: UserID=%s, Coin=%s, WalletBalance=%s, Free=%s",
				userID, coin.Coin, coin.WalletBalance, coin.Free)
		}

	default:
		logger.LogInfo("Неизвестный приватный топик: %s", msg.Topic)
	}
}
