package storages

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/integration/redis"
	"context"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

// Публичные методы для работы с Redis

// SaveTicker сохраняет данные тикера
func SaveTicker(ctx context.Context, symbol string, ticker bybit.TickerMessage) error {
	key := fmt.Sprintf("tickers:%s", symbol)
	data, err := json.Marshal(ticker)
	if err != nil {
		return fmt.Errorf("failed to marshal ticker: %w", err)
	}

	return redis.Client.Set(ctx, key, data, 1*time.Hour).Err()
}

// SaveOrderBook сохраняет данные книги ордеров
func SaveOrderBook(ctx context.Context, symbol string, orderBook bybit.OrderBookMessage) error {
	key := fmt.Sprintf("orderbook:%s", symbol)
	data, err := json.Marshal(orderBook)
	if err != nil {
		return fmt.Errorf("failed to marshal orderbook: %w", err)
	}

	return redis.Client.Set(ctx, key, data, 1*time.Hour).Err()
}

// SavePublicTrade сохраняет данные публичной сделки
func SavePublicTrade(ctx context.Context, symbol string, trade bybit.TradeMessage) error {
	key := fmt.Sprintf("publicTrade:%s", symbol)
	data, err := json.Marshal(trade)
	if err != nil {
		return fmt.Errorf("failed to marshal trade: %w", err)
	}

	// Добавляем сделку в список
	if err := redis.Client.RPush(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("failed to save trade: %w", err)
	}

	// Ограничиваем список 1000 сделками
	if err := redis.Client.LTrim(ctx, key, 0, 999).Err(); err != nil {
		return fmt.Errorf("failed to trim trade list: %w", err)
	}

	// Устанавливаем TTL на 24 часа
	if err := redis.Client.Expire(ctx, key, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set trade list TTL: %w", err)
	}

	return nil
}

// GetTicker получает данные тикера
func GetTicker(ctx context.Context, symbol string) (*bybit.TickerMessage, error) {
	key := fmt.Sprintf("tickers:%s", symbol)
	data, err := redis.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}

	var ticker bybit.TickerMessage
	if err := json.Unmarshal(data, &ticker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ticker: %w", err)
	}

	return &ticker, nil
}

// GetOrderBook получает данные книги ордеров
func GetOrderBook(ctx context.Context, symbol string) (*bybit.OrderBookMessage, error) {
	key := fmt.Sprintf("orderbook:%s", symbol)
	data, err := redis.Client.Get(ctx, key).Bytes()
		if err != nil {
		return nil, fmt.Errorf("failed to get orderbook: %w", err)
	}

	var orderBook bybit.OrderBookMessage
	if err := json.Unmarshal(data, &orderBook); err != nil {
		return nil, fmt.Errorf("failed to unmarshal orderbook: %w", err)
	}

	return &orderBook, nil
}

// GetPublicTrades получает последние N публичных сделок
func GetPublicTrades(ctx context.Context, symbol string, limit int64) ([]bybit.TradeMessage, error) {
	key := fmt.Sprintf("publicTrade:%s", symbol)
	data, err := redis.Client.LRange(ctx, key, -limit, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
		}

	var trades []bybit.TradeMessage
	for _, item := range data {
		var trade bybit.TradeMessage
		if err := json.Unmarshal([]byte(item), &trade); err != nil {
			return nil, fmt.Errorf("failed to unmarshal trade: %w", err)
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

// Приватные методы для работы с Redis

// SavePrivateOrder сохраняет данные приватного ордера
func SavePrivateOrder(ctx context.Context, userID string, orderID string, order bybit.OrderMessage) error {
	key := fmt.Sprintf("private:%s:order:%s", userID, orderID)
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	return redis.Client.Set(ctx, key, data, 24*time.Hour).Err()
}

// SavePrivateExecution сохраняет данные приватного исполнения
func SavePrivateExecution(ctx context.Context, userID string, execID string, execution bybit.ExecutionMessage) error {
	key := fmt.Sprintf("private:%s:execution:%s", userID, execID)
	data, err := json.Marshal(execution)
	if err != nil {
		return fmt.Errorf("failed to marshal execution: %w", err)
	}

	return redis.Client.Set(ctx, key, data, 24*time.Hour).Err()
}

// SavePrivateWallet сохраняет данные приватного кошелька
func SavePrivateWallet(ctx context.Context, userID string, wallet bybit.WalletMessage) error {
	key := fmt.Sprintf("private:%s:wallet", userID)
	data, err := json.Marshal(wallet)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet: %w", err)
	}

	return redis.Client.Set(ctx, key, data, 1*time.Hour).Err()
}

// GetPrivateOrder получает данные приватного ордера
func GetPrivateOrder(ctx context.Context, userID string, orderID string) (*bybit.OrderMessage, error) {
	key := fmt.Sprintf("private:%s:order:%s", userID, orderID)
	data, err := redis.Client.Get(ctx, key).Bytes()
		if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	var order bybit.OrderMessage
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

// GetPrivateExecution получает данные приватного исполнения
func GetPrivateExecution(ctx context.Context, userID string, execID string) (*bybit.ExecutionMessage, error) {
	key := fmt.Sprintf("private:%s:execution:%s", userID, execID)
	data, err := redis.Client.Get(ctx, key).Bytes()
			if err != nil {
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	var execution bybit.ExecutionMessage
	if err := json.Unmarshal(data, &execution); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execution: %w", err)
	}

	return &execution, nil
			}

// GetPrivateWallet получает данные приватного кошелька
func GetPrivateWallet(ctx context.Context, userID string) (*bybit.WalletMessage, error) {
	key := fmt.Sprintf("private:%s:wallet", userID)
	data, err := redis.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	var wallet bybit.WalletMessage
	if err := json.Unmarshal(data, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet: %w", err)
	}

	return &wallet, nil
}

// SaveOrderBookSpread сохраняет спред книги ордеров
func SaveOrderBookSpread(ctx context.Context, symbol string, spread decimal.Decimal) error {
	key := fmt.Sprintf("orderbook:spread:%s", symbol)
	return redis.Client.Set(ctx, key, spread.String(), time.Hour).Err()
}

// GetOrderBookSpread получает текущий спред для книги ордеров
func GetOrderBookSpread(ctx context.Context, symbol string) (decimal.Decimal, error) {
	key := fmt.Sprintf("orderbook:spread:%s", symbol)
	data, err := redis.Client.Get(ctx, key).Result()
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get spread: %w", err)
	}
	spread, err := decimal.NewFromString(data)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to parse spread: %w", err)
	}
	return spread, nil
}

// SaveTickerHistory сохраняет тикер в список истории
func SaveTickerHistory(ctx context.Context, symbol string, ticker bybit.TickerMessage) error {
	key := fmt.Sprintf("tickers:history:%s", symbol)
	data, err := json.Marshal(ticker)
	if err != nil {
		return fmt.Errorf("failed to marshal ticker: %w", err)
	}
	if err := redis.Client.LPush(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("failed to save ticker history: %w", err)
	}
	if err := redis.Client.LTrim(ctx, key, 0, 999).Err(); err != nil { // Храним 1000 записей
		return fmt.Errorf("failed to trim ticker history: %w", err)
	}
	return redis.Client.Expire(ctx, key, time.Hour).Err()
}

// GetTickerHistory получает последние N тикеров
func GetTickerHistory(ctx context.Context, symbol string, limit int64) ([]bybit.TickerMessage, error) {
	key := fmt.Sprintf("tickers:history:%s", symbol)
	data, err := redis.Client.LRange(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker history: %w", err)
	}
	var tickers []bybit.TickerMessage
	for _, item := range data {
		var ticker bybit.TickerMessage
		if err := json.Unmarshal([]byte(item), &ticker); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ticker: %w", err)
			}
		tickers = append(tickers, ticker)
	}
	return tickers, nil
}

// SaveOrderBookHistory сохраняет книгу ордеров в список истории
func SaveOrderBookHistory(ctx context.Context, symbol string, orderBook bybit.OrderBookMessage) error {
	key := fmt.Sprintf("orderbook:history:%s", symbol)
	data, err := json.Marshal(orderBook)
	if err != nil {
		return fmt.Errorf("failed to marshal orderbook: %w", err)
	}
	if err := redis.Client.LPush(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("failed to save orderbook history: %w", err)
	}
	if err := redis.Client.LTrim(ctx, key, 0, 999).Err(); err != nil { // Храним 10 записей
		return fmt.Errorf("failed to trim orderbook history: %w", err)
	}
	return redis.Client.Expire(ctx, key, time.Hour).Err()
}

// GetOrderBookHistory получает последние N книг ордеров
func GetOrderBookHistory(ctx context.Context, symbol string, limit int64) ([]bybit.OrderBookMessage, error) {
	key := fmt.Sprintf("orderbook:history:%s", symbol)
	data, err := redis.Client.LRange(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get orderbook history: %w", err)
	}
	var orderBooks []bybit.OrderBookMessage
	for _, item := range data {
		var orderBook bybit.OrderBookMessage
		if err := json.Unmarshal([]byte(item), &orderBook); err != nil {
			return nil, fmt.Errorf("failed to unmarshal orderbook: %w", err)
	}
		orderBooks = append(orderBooks, orderBook)
	}
	return orderBooks, nil
}

// Close закрывает соединение с Redis
func Close() error {
	return redis.Client.Close()
}
