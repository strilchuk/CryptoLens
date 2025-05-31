package trading

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/storages"
	"CryptoLens_Backend/types"
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"sync"
)

// StrategyManager управляет стратегиями
type StrategyManager struct {
	strategies         map[string][]types.Strategy // userID -> список стратегий
	userInstruments    map[string][]string         // userID -> список символов из user_instruments
	bybitClient        bybit.Client
	userInstrumentRepo types.UserInstrumentRepositoryInterface
	bybitAccountRepo   types.BybitAccountRepositoryInterface
	mutex              sync.Mutex
}

// NewStrategyManager создает новый менеджер стратегий
func NewStrategyManager(client bybit.Client, userInstrumentRepo types.UserInstrumentRepositoryInterface, bybitAccountRepo types.BybitAccountRepositoryInterface) *StrategyManager {
	return &StrategyManager{
		strategies:         make(map[string][]types.Strategy),
		userInstruments:    make(map[string][]string),
		bybitClient:        client,
		userInstrumentRepo: userInstrumentRepo,
		bybitAccountRepo:   bybitAccountRepo,
	}
}

// getBybitAccount получает аккаунт Bybit для пользователя
func (m *StrategyManager) getBybitAccount(ctx context.Context, userID string) (*bybit.BybitAccount, error) {
	return m.bybitAccountRepo.GetActiveAccountByUserID(ctx, userID)
}

// CreateOrder создает ордер
func (m *StrategyManager) CreateOrder(ctx context.Context, userID, symbol, side, orderType, qty string, price *string) (*bybit.BybitOrderResponse, error) {
	// Получаем аккаунт Bybit пользователя
	account, err := m.getBybitAccount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Bybit account: %w", err)
	}

	// Создаем ордер через клиент Bybit
	order, err := m.bybitClient.CreateOrder(ctx, account, symbol, side, orderType, qty, price, "GTC", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

// CancelOrder отменяет ордер
func (m *StrategyManager) CancelOrder(ctx context.Context, userID, symbol, orderID string) error {
	// Получаем аккаунт Bybit пользователя
	account, err := m.getBybitAccount(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get Bybit account: %w", err)
	}

	// Отменяем ордер через клиент Bybit
	if err, _ := m.bybitClient.CancelOrder(ctx, account, symbol, orderID); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

// AddStrategy добавляет стратегию для пользователя
func (m *StrategyManager) AddStrategy(userID string, strategy types.Strategy) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.strategies[userID] = append(m.strategies[userID], strategy)
}

// RemoveStrategy удаляет стратегию для пользователя
func (m *StrategyManager) RemoveStrategy(userID string, strategy types.Strategy) {
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

// UpdateUserInstruments обновляет список активных символов пользователя
func (m *StrategyManager) UpdateUserInstruments(ctx context.Context, userID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	symbols, err := m.userInstrumentRepo.GetActiveInstrumentsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get active instruments for user %s: %w", userID, err)
	}
	m.userInstruments[userID] = symbols
	return nil
}

// isSymbolRelevant проверяет, относится ли символ к активным инструментам пользователя
func (m *StrategyManager) isSymbolRelevant(userID, symbol string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, s := range m.userInstruments[userID] {
		if s == symbol {
			return true
		}
	}
	return false
}

// HandleTicker обрабатывает тикер
func (m *StrategyManager) HandleTicker(ctx context.Context, ticker bybit.TickerMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for userID, strategies := range m.strategies {
		if m.isSymbolRelevant(userID, ticker.Symbol) {
			for _, s := range strategies {
				s.OnTicker(ctx, ticker)
			}
		}
	}
}

// HandleOrderBook обрабатывает книгу ордеров
func (m *StrategyManager) HandleOrderBook(ctx context.Context, orderBook bybit.OrderBookMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for userID, strategies := range m.strategies {
		if m.isSymbolRelevant(userID, orderBook.Symbol) {
			for _, s := range strategies {
				s.OnOrderBook(ctx, orderBook)
			}
		}
	}
}

// HandleTrade обрабатывает сделку
func (m *StrategyManager) HandleTrade(ctx context.Context, trade bybit.TradeMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for userID, strategies := range m.strategies {
		if m.isSymbolRelevant(userID, trade.Symbol) {
			for _, s := range strategies {
				s.OnTrade(ctx, trade)
			}
		}
	}
}

// HandleOrder обрабатывает ордер
func (m *StrategyManager) HandleOrder(ctx context.Context, order bybit.OrderMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for userID, strategies := range m.strategies {
		if m.isSymbolRelevant(userID, order.Symbol) {
			for _, s := range strategies {
				s.OnOrder(ctx, order)
			}
		}
	}
}

// HandleExecution обрабатывает исполнение
func (m *StrategyManager) HandleExecution(ctx context.Context, execution bybit.ExecutionMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for userID, strategies := range m.strategies {
		if m.isSymbolRelevant(userID, execution.Symbol) {
			for _, s := range strategies {
				s.OnExecution(ctx, execution)
			}
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
func (m *StrategyManager) GetStrategies(userID string) []types.Strategy {
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

// Методы для чтения данных из Redis
func (m *StrategyManager) GetTicker(ctx context.Context, symbol string) (*bybit.TickerMessage, error) {
	return storages.GetTicker(ctx, symbol)
}

func (m *StrategyManager) GetTickerHistory(ctx context.Context, symbol string, limit int64) ([]bybit.TickerMessage, error) {
	return storages.GetTickerHistory(ctx, symbol, limit)
}

func (m *StrategyManager) GetOrderBook(ctx context.Context, symbol string) (*bybit.OrderBookMessage, error) {
	return storages.GetOrderBook(ctx, symbol)
}

func (m *StrategyManager) GetOrderBookHistory(ctx context.Context, symbol string, limit int64) ([]bybit.OrderBookMessage, error) {
	return storages.GetOrderBookHistory(ctx, symbol, limit)
}

func (m *StrategyManager) GetOrderBookSpread(ctx context.Context, symbol string) (decimal.Decimal, error) {
	return storages.GetOrderBookSpread(ctx, symbol)
}

func (m *StrategyManager) GetPublicTrades(ctx context.Context, symbol string, limit int64) ([]bybit.TradeMessage, error) {
	return storages.GetPublicTrades(ctx, symbol, limit)
}

func (m *StrategyManager) GetPrivateOrder(ctx context.Context, userID, orderID string) (*bybit.OrderMessage, error) {
	return storages.GetPrivateOrder(ctx, userID, orderID)
}

func (m *StrategyManager) GetPrivateExecution(ctx context.Context, userID, execID string) (*bybit.ExecutionMessage, error) {
	return storages.GetPrivateExecution(ctx, userID, execID)
}

func (m *StrategyManager) GetPrivateWallet(ctx context.Context, userID string) (*bybit.WalletMessage, error) {
	return storages.GetPrivateWallet(ctx, userID)
}

// GetWalletBalance получает баланс кошелька через API
func (m *StrategyManager) GetWalletBalance(ctx context.Context, userID string) (*bybit.BybitWalletBalance, error) {
	// Получаем аккаунт Bybit пользователя
	account, err := m.getBybitAccount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Bybit account: %w", err)
	}

	// Получаем баланс через клиент Bybit
	balance, err := m.bybitClient.GetWalletBalance(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	return balance, nil
}
