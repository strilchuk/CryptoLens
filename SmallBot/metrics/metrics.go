package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"sync"
	"time"
)

// Metrics хранит все метрики бота
type Metrics struct {
	mu sync.RWMutex

	// Счетчики ордеров
	OrdersCreated   int64 `json:"orders_created"`
	OrdersFilled    int64 `json:"orders_filled"`
	OrdersCancelled int64 `json:"orders_cancelled"`
	OrdersTimeout   int64 `json:"orders_timeout"`

	// Счетчики ошибок
	Errors          int64            `json:"errors_total"`
	ErrorsByType    map[string]int64 `json:"errors_by_type"`
	WebSocketErrors int64            `json:"websocket_errors"`
	APIErrors       int64            `json:"api_errors"`

	// Финансовые метрики
	TotalVolume   decimal.Decimal `json:"total_volume"`
	TotalFees     decimal.Decimal `json:"total_fees"`
	RealizedPnL   decimal.Decimal `json:"realized_pnl"`
	UnrealizedPnL decimal.Decimal `json:"unrealized_pnl"`

	// Производительность
	LastOrderTime    time.Time       `json:"last_order_time"`
	AverageExecTime  time.Duration   `json:"average_exec_time"`
	OrderExecTimes   []time.Duration `json:"-"`
	WebSocketLatency time.Duration   `json:"websocket_latency"`

	// Состояние системы
	StartTime       time.Time       `json:"start_time"`
	Uptime          string          `json:"uptime"`
	ActiveOrders    int             `json:"active_orders"`
	LastTickerPrice string          `json:"last_ticker_price"`
	CurrentBalance  decimal.Decimal `json:"current_balance"`
}

var instance *Metrics
var once sync.Once

// GetInstance возвращает синглтон метрик
func GetInstance() *Metrics {
	once.Do(func() {
		instance = &Metrics{
			StartTime:    time.Now(),
			ErrorsByType: make(map[string]int64),
		}
	})
	return instance
}

// IncrementOrdersCreated увеличивает счетчик созданных ордеров
func (m *Metrics) IncrementOrdersCreated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OrdersCreated++
	m.LastOrderTime = time.Now()
}

// IncrementOrdersFilled увеличивает счетчик исполненных ордеров
func (m *Metrics) IncrementOrdersFilled() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OrdersFilled++
}

// IncrementOrdersCancelled увеличивает счетчик отмененных ордеров
func (m *Metrics) IncrementOrdersCancelled() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OrdersCancelled++
}

// IncrementOrdersTimeout увеличивает счетчик ордеров с таймаутом
func (m *Metrics) IncrementOrdersTimeout() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OrdersTimeout++
}

// IncrementError увеличивает счетчик ошибок
func (m *Metrics) IncrementError(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors++
	m.ErrorsByType[errorType]++
}

// IncrementWebSocketError увеличивает счетчик ошибок WebSocket
func (m *Metrics) IncrementWebSocketError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WebSocketErrors++
}

// IncrementAPIError увеличивает счетчик ошибок API
func (m *Metrics) IncrementAPIError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.APIErrors++
}

// AddVolume добавляет объем торговли
func (m *Metrics) AddVolume(volume decimal.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalVolume = m.TotalVolume.Add(volume)
}

// AddFees добавляет комиссии
func (m *Metrics) AddFees(fees decimal.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalFees = m.TotalFees.Add(fees)
}

// UpdatePnL обновляет прибыль/убыток
func (m *Metrics) UpdatePnL(realized, unrealized decimal.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RealizedPnL = realized
	m.UnrealizedPnL = unrealized
}

// RecordOrderExecution записывает время исполнения ордера
func (m *Metrics) RecordOrderExecution(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.OrderExecTimes = append(m.OrderExecTimes, duration)

	// Ограничиваем размер массива
	if len(m.OrderExecTimes) > 1000 {
		m.OrderExecTimes = m.OrderExecTimes[len(m.OrderExecTimes)-1000:]
	}

	// Пересчитываем среднее
	var total time.Duration
	for _, d := range m.OrderExecTimes {
		total += d
	}
	if len(m.OrderExecTimes) > 0 {
		m.AverageExecTime = total / time.Duration(len(m.OrderExecTimes))
	}
}

// UpdateWebSocketLatency обновляет задержку WebSocket
func (m *Metrics) UpdateWebSocketLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WebSocketLatency = latency
}

// UpdateSystemState обновляет состояние системы
func (m *Metrics) UpdateSystemState(activeOrders int, lastPrice string, balance decimal.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveOrders = activeOrders
	m.LastTickerPrice = lastPrice
	m.CurrentBalance = balance
	m.Uptime = time.Since(m.StartTime).String()
}

// GetSnapshot возвращает снимок всех метрик
func (m *Metrics) GetSnapshot() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := *m
	snapshot.Uptime = time.Since(m.StartTime).String()
	return &snapshot
}

// ToJSON возвращает метрики в формате JSON
func (m *Metrics) ToJSON() (string, error) {
	snapshot := m.GetSnapshot()
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetSummary возвращает краткую сводку метрик
func (m *Metrics) GetSummary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.StartTime)
	successRate := float64(0)
	if m.OrdersCreated > 0 {
		successRate = float64(m.OrdersFilled) / float64(m.OrdersCreated) * 100
	}

	return fmt.Sprintf(`
📊 МЕТРИКИ БОТА
═══════════════════════════════
⏱  Uptime: %s
📈 Ордеров создано: %d
✅ Ордеров исполнено: %d (%.1f%%)
❌ Ордеров отменено: %d
⏰ Таймауты: %d
💰 Объем торговли: %s USDT
💸 Комиссии: %s USDT
📊 Realized P&L: %s USDT
🔧 Ошибок: %d (API: %d, WS: %d)
⚡ Средн. время исполнения: %s
🌐 WebSocket задержка: %s
💼 Активных ордеров: %d
💵 Баланс: %s USDT
═══════════════════════════════`,
		uptime,
		m.OrdersCreated,
		m.OrdersFilled, successRate,
		m.OrdersCancelled,
		m.OrdersTimeout,
		m.TotalVolume.StringFixed(2),
		m.TotalFees.StringFixed(4),
		m.RealizedPnL.StringFixed(2),
		m.Errors, m.APIErrors, m.WebSocketErrors,
		m.AverageExecTime,
		m.WebSocketLatency,
		m.ActiveOrders,
		m.CurrentBalance.StringFixed(2),
	)
}
