package bybit

import (
	"SmallBot/logger"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"strings"
	"sync"
	"time"
)

type WebSocketClient struct {
	url        string
	conn       *websocket.Conn
	recvWindow int
	apiKey     string // Для приватных каналов
	apiSecret  string // Для приватных каналов
	mutex      sync.Mutex
}

// представляет базовое сообщение WebSocket
type WebSocketMessage struct {
	Topic string          `json:"topic"`
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data"`
	Ts    int64           `json:"ts"`
}

// представляет сообщение тикера
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

// представляет сообщение книги ордеров
type OrderBookMessage struct {
	Symbol   string      `json:"s"`
	Bids     [][2]string `json:"b"`
	Asks     [][2]string `json:"a"`
	UpdateID int64       `json:"u"`
	Seq      int64       `json:"seq"`
}

// представляет сообщение о сделке
type TradeMessage struct {
	ID           string `json:"i"`
	Timestamp    int64  `json:"T"`
	Price        string `json:"p"`
	Volume       string `json:"v"`
	Side         string `json:"S"`
	Symbol       string `json:"s"`
	IsBlockTrade bool   `json:"BT"`
	IsRPI        bool   `json:"RPI"`
}

// представляет сообщение об ордере
type OrderMessage struct {
	OrderID      string `json:"orderId"`
	OrderLinkID  string `json:"orderLinkId"`
	Symbol       string `json:"symbol"`
	Side         string `json:"side"`
	OrderType    string `json:"orderType"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	TimeInForce  string `json:"timeInForce"`
	OrderStatus  string `json:"orderStatus"`
	CreatedTime  string `json:"createdTime"`
	UpdatedTime  string `json:"updatedTime"`
	CumExecQty   string `json:"cumExecQty"`
	CumExecValue string `json:"cumExecValue"`
	CumExecFee   string `json:"cumExecFee"`
	Category     string `json:"category"`
}

// представляет сообщение об исполнении ордера
type ExecutionMessage struct {
	ExecID      string `json:"execId"`
	OrderID     string `json:"orderId"`
	OrderLinkID string `json:"orderLinkId"`
	Symbol      string `json:"symbol"`
	Side        string `json:"side"`
	ExecPrice   string `json:"execPrice"`
	ExecQty     string `json:"execQty"`
	ExecFee     string `json:"execFee"`
	FeeRate     string `json:"feeRate"`
	IsMaker     bool   `json:"isMaker"`
	OrderType   string `json:"orderType"`
	ExecTime    string `json:"execTime"`
	Category    string `json:"category"`
}

type WalletMessage struct {
	AccountType string `json:"accountType"`
	Coin        []struct {
		Coin          string `json:"coin"`
		WalletBalance string `json:"walletBalance"`
		Free          string `json:"free"`
		Locked        string `json:"locked"`
	} `json:"coin"`
}

func NewWebSocketClient(url string, recvWindow int, apiKey, apiSecret string) *WebSocketClient {
	return &WebSocketClient{
		url:        url,
		recvWindow: recvWindow,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
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

	logger.LogInfo("Успешно подключились к WebSocket: %s", c.url)

	// Аутентификация для приватных каналов
	if c.apiKey != "" && c.apiSecret != "" {
		if err := c.authenticate(ctx); err != nil {
			c.conn.Close()
			c.conn = nil
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	go c.startPing(ctx)
	return nil
}

// authenticate отправляет запрос на аутентификацию для приватных каналов
func (c *WebSocketClient) authenticate(ctx context.Context) error {
	if c.conn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	expires := time.Now().UnixMilli() + int64(c.recvWindow)
	authMsg := map[string]interface{}{
		"op": "auth",
		"args": []interface{}{
			c.apiKey,
			expires,
			c.generateSignature(expires),
		},
	}

	logger.LogInfo("Sending auth message: %v", authMsg)
	if err := c.conn.WriteJSON(authMsg); err != nil {
		return fmt.Errorf("failed to send auth message: %w", err)
	}

	// Ожидаем подтверждения аутентификации
	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	var response struct {
		Success bool   `json:"success"`
		RetMsg  string `json:"ret_msg"`
	}
	if err := json.Unmarshal(msg, &response); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("authentication failed: %s", response.RetMsg)
	}

	logger.LogInfo("Успешная аутентификация для WebSocket")
	return nil
}

// generateSignature создает подпись для аутентификации
func (c *WebSocketClient) generateSignature(expires int64) string {
	val := fmt.Sprintf("GET/realtime%d", expires)
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(val))
	return hex.EncodeToString(h.Sum(nil))
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
	messageChan := make(chan WebSocketMessage)
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.LogInfo("WebSocket context canceled, attempting reconnect...")
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

				logger.LogDebug("Waiting for WebSocket message...")
				_, msg, err := c.conn.ReadMessage()
				if err != nil {
					logger.LogError("Failed to read WebSocket message: %v", err)
					c.Close()
					time.Sleep(5 * time.Second)
					continue
				}

				logger.LogDebug("Received raw WebSocket message: %s", string(msg))
				var message WebSocketMessage
				if err := json.Unmarshal(msg, &message); err != nil {
					logger.LogError("Failed to parse WebSocket message: %v", err)
					continue
				}

				if message.Topic == "" && strings.Contains(string(msg), "pong") {
					logger.LogDebug("Received pong message")
					continue
				}

				messageChan <- message
			}
		}
	}()

	go func() {
		for msg := range messageChan {
			handler(ctx, msg)
		}
	}()
}

// отправляет пинг каждые 20 секунд
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

func (c *WebSocketClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
