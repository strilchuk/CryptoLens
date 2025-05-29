package bybit

import (
	"CryptoLens_Backend/logger"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"strings"
	"sync"
	"time"
)

// WebSocketClient представляет WebSocket-клиент для Bybit
type WebSocketClient struct {
	url        string
	conn       *websocket.Conn
	recvWindow int
	mutex      sync.Mutex
}

// WebSocketMessage представляет базовое сообщение WebSocket
type WebSocketMessage struct {
	Topic string          `json:"topic"`
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data"`
	Ts    int64           `json:"ts"`
}

// TickerMessage представляет сообщение тикера
type TickerMessage struct {
	Symbol        string `json:"symbol"`
	LastPrice     string `json:"lastPrice"`
	HighPrice24h  string `json:"highPrice24h"`
	LowPrice24h   string `json:"lowPrice24h"`
	PrevPrice24h  string `json:"prevPrice24h"`
	Volume24h     string `json:"volume24h"`
	Turnover24h   string `json:"turnover24h"`
	Price24hPcnt  string `json:"price24hPcnt"`
	UsdIndexPrice string `json:"usdIndexPrice"`
}

// OrderBookMessage представляет сообщение книги ордеров
type OrderBookMessage struct {
	Symbol string      `json:"s"`
	Bids   [][2]string `json:"b"`
	Asks   [][2]string `json:"a"`
	UpdateID int64     `json:"u"`
	Seq     int64      `json:"seq"`
}

// TradeMessage представляет сообщение о сделке
type TradeMessage struct {
	ID        string `json:"i"`
	Timestamp int64  `json:"T"`
	Price     string `json:"p"`
	Volume    string `json:"v"`
	Side      string `json:"S"`
	Symbol    string `json:"s"`
	IsBlockTrade bool `json:"BT"`
	IsRPI      bool   `json:"RPI"`
}

// NewWebSocketClient создает новый WebSocket-клиент
func NewWebSocketClient(url string, recvWindow int) *WebSocketClient {
	return &WebSocketClient{
		url:        url,
		recvWindow: recvWindow,
	}
}

// Connect устанавливает соединение с WebSocket
func (c *WebSocketClient) Connect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		return nil // Уже подключены
	}

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	c.conn = conn

	// Логируем успешное подключение
	logger.LogInfo("Успешно подключились к WebSocket: %s", c.url)

	// Запускаем пинг каждые 20 секунд
	go c.startPing(ctx)

	return nil
}

// Subscribe подписывается на указанные каналы
func (c *WebSocketClient) Subscribe(ctx context.Context, channels []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	subscribeMsg := map[string]interface{}{
		"op":   "subscribe",
		"args": channels,
	}
	return c.conn.WriteJSON(subscribeMsg)
}

// StartMessageHandler запускает обработку входящих сообщений
func (c *WebSocketClient) StartMessageHandler(ctx context.Context, handler func(context.Context, WebSocketMessage)) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.Close()
				return
			default:
				if c.conn == nil {
					logger.LogError("WebSocket connection is nil, attempting to reconnect")
					if err := c.Connect(ctx); err != nil {
						logger.LogError("Reconnect failed: %v", err)
						time.Sleep(5 * time.Second)
						continue
					}
				}

				_, msg, err := c.conn.ReadMessage()
				if err != nil {
					logger.LogError("Failed to read WebSocket message: %v", err)
					c.Close()
					time.Sleep(5 * time.Second)
					continue
				}

				var message WebSocketMessage
				if err := json.Unmarshal(msg, &message); err != nil {
					logger.LogError("Failed to parse WebSocket message: %v", err)
					continue
				}

				// Пропускаем сообщения пинга
				if message.Topic == "" && strings.Contains(string(msg), "pong") {
					continue
				}

				handler(ctx, message)
			}
		}
	}()
}

// startPing отправляет пинг каждые 20 секунд
func (c *WebSocketClient) startPing(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mutex.Lock()
			if c.conn != nil {
				err := c.conn.WriteJSON(map[string]string{"op": "ping"})
				if err != nil {
					logger.LogError("Failed to send ping: %v", err)
					c.Close() // Закрываем соединение при ошибке пинга
				}
			}
			c.mutex.Unlock()
		}
	}
}

// Close закрывает соединение
func (c *WebSocketClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
