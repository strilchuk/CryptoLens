package trading

import (
	"CryptoLens_Backend/integration/bybit"
	"context"
	"fmt"
	"sync"
)

// Strategy определяет интерфейс для торговых стратегий
type Strategy interface {
	// OnTicker вызывается при получении тикера
	OnTicker(ctx context.Context, ticker bybit.TickerMessage)
	// OnOrderBook вызывается при обновлении книги ордеров
	OnOrderBook(ctx context.Context, orderBook bybit.OrderBookMessage)
	// OnTrade вызывается при новой сделке
	OnTrade(ctx context.Context, trade bybit.TradeMessage)
	// OnOrder вызывается при обновлении ордера
	OnOrder(ctx context.Context, order bybit.OrderMessage)
	// OnExecution вызывается при исполнении ордера
	OnExecution(ctx context.Context, execution bybit.ExecutionMessage)
	// OnWallet вызывается при обновлении баланса
	OnWallet(ctx context.Context, wallet bybit.WalletMessage)
	// Start запускает стратегию
	Start(ctx context.Context)
	// Stop останавливает стратегию
	Stop(ctx context.Context)
}

// StrategyManager управляет стратегиями
type StrategyManager struct {
	strategies  map[string][]Strategy // userID -> список стратегий
	bybitClient bybit.Client
	mutex       sync.Mutex
}

// NewStrategyManager создает новый менеджер стратегий
func NewStrategyManager(client bybit.Client) *StrategyManager {
	return &StrategyManager{
		strategies:  make(map[string][]Strategy),
		bybitClient: client,
	}
}

// AddStrategy добавляет стратегию для пользователя
func (m *StrategyManager) AddStrategy(userID string, strategy Strategy) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.strategies[userID] = append(m.strategies[userID], strategy)
}

// RemoveStrategy удаляет стратегию для пользователя
func (m *StrategyManager) RemoveStrategy(userID string, strategy Strategy) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	strategies := m.strategies[userID]
	for i, s := range strategies {
		if s == strategy {
			m.strategies[userID] = append(strategies[:i], strategies[i+1:]...)
			break
		}
	}
}

// HandleTicker обрабатывает тикер
func (m *StrategyManager) HandleTicker(ctx context.Context, ticker bybit.TickerMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for _, strategies := range m.strategies {
		for _, s := range strategies {
			s.OnTicker(ctx, ticker)
		}
	}
}

// HandleOrderBook обрабатывает книгу ордеров
func (m *StrategyManager) HandleOrderBook(ctx context.Context, orderBook bybit.OrderBookMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for _, strategies := range m.strategies {
		for _, s := range strategies {
			s.OnOrderBook(ctx, orderBook)
		}
	}
}

// HandleTrade обрабатывает сделку
func (m *StrategyManager) HandleTrade(ctx context.Context, trade bybit.TradeMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for _, strategies := range m.strategies {
		for _, s := range strategies {
			s.OnTrade(ctx, trade)
		}
	}
}

// HandleOrder обрабатывает ордер
func (m *StrategyManager) HandleOrder(ctx context.Context, order bybit.OrderMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for _, strategies := range m.strategies {
		for _, s := range strategies {
			s.OnOrder(ctx, order)
		}
	}
}

// HandleExecution обрабатывает исполнение
func (m *StrategyManager) HandleExecution(ctx context.Context, execution bybit.ExecutionMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for _, strategies := range m.strategies {
		for _, s := range strategies {
			s.OnExecution(ctx, execution)
		}
	}
}

// HandleWallet обрабатывает обновление кошелька
func (m *StrategyManager) HandleWallet(ctx context.Context, wallet bybit.WalletMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for _, strategies := range m.strategies {
		for _, s := range strategies {
			s.OnWallet(ctx, wallet)
		}
	}
}

// Start запускает все стратегии
func (m *StrategyManager) Start(ctx context.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for _, strategies := range m.strategies {
		for _, s := range strategies {
			s.Start(ctx)
		}
	}
}

// Stop останавливает все стратегии
func (m *StrategyManager) Stop(ctx context.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for _, strategies := range m.strategies {
		for _, s := range strategies {
			s.Stop(ctx)
		}
	}
}

// GetStrategies возвращает все стратегии пользователя
func (m *StrategyManager) GetStrategies(userID string) []Strategy {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.strategies[userID]
}

// GetStrategiesInfo возвращает информацию о стратегиях в менеджере
func (m *StrategyManager) GetStrategiesInfo() map[string][]string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	info := make(map[string][]string)
	for userID, strategies := range m.strategies {
		var strategyNames []string
		for _, s := range strategies {
			strategyNames = append(strategyNames, fmt.Sprintf("%T", s))
		}
		info[userID] = strategyNames
	}
	return info
} 