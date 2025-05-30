package services

import (
	"CryptoLens_Backend/env"
	"CryptoLens_Backend/handlers"
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/models"
	"CryptoLens_Backend/repositories"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BybitService struct {
	bybitClient         bybit.Client
	wsClient            *bybit.WebSocketClient
	privateWsClients    map[string]*bybit.WebSocketClient // Карта приватных клиентов по userID
	db                  *sql.DB
	userService         *UserService
	bybitInstrumentRepo *repositories.BybitInstrumentRepository
	userInstrumentRepo  *repositories.UserInstrumentRepository
	wsHandler           *handlers.BybitWebSocketHandler
	wsMutex             sync.Mutex
}

func NewBybitService(bybitClient bybit.Client, db *sql.DB, userService *UserService) *BybitService {
	recvWindow, _ := strconv.Atoi(env.GetBybitRecvWindow())
	apiMode := env.GetBybitApiMode()
	wsURL := env.GetBybitWsUrl() + "/v5/public/spot"
	if apiMode == "test" {
		wsURL = env.GetBybitWsTestUrl() + "/v5/public/spot"
	}
	wsClient := bybit.NewWebSocketClient(wsURL, recvWindow, "", "")

	return &BybitService{
		bybitClient:         bybitClient,
		wsClient:            wsClient,
		privateWsClients:    make(map[string]*bybit.WebSocketClient),
		db:                  db,
		userService:         userService,
		bybitInstrumentRepo: repositories.NewBybitInstrumentRepository(db),
		userInstrumentRepo:  repositories.NewUserInstrumentRepository(db),
		wsHandler:           handlers.NewBybitWebSocketHandler(),
	}
}

func (s *BybitService) GetWalletBalance(ctx context.Context, token string) (*bybit.BybitWalletBalance, error) {
	// Получаем ID пользователя из токена
	userID, err := s.userService.validateToken(token)
	if err != nil {
		return nil, err
	}

	// Получаем аккаунт Bybit пользователя
	account, err := s.getBybitAccount(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Получаем баланс через клиент Bybit
	balance, err := s.bybitClient.GetWalletBalance(ctx, account)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (s *BybitService) GetFeeRate(ctx context.Context, token string, category string, symbol string, baseCoin string) (*bybit.BybitFeeRateResponse, error) {
	// Получаем ID пользователя из токена
	userID, err := s.userService.validateToken(token)
	if err != nil {
		return nil, err
	}

	// Получаем аккаунт Bybit пользователя
	account, err := s.getBybitAccount(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Получаем ставки комиссии через клиент Bybit
	var symbolPtr, baseCoinPtr *string
	if symbol != "" {
		symbolPtr = &symbol
	}
	if baseCoin != "" {
		baseCoinPtr = &baseCoin
	}

	feeRate, err := s.bybitClient.GetFeeRate(ctx, account, category, symbolPtr, baseCoinPtr)
	if err != nil {
		return nil, err
	}

	return feeRate, nil
}

func (s *BybitService) GetInstruments(ctx context.Context, category string) ([]models.BybitInstrument, error) {
	return s.bybitInstrumentRepo.GetInstruments(ctx, category)
}

func (s *BybitService) StartInstrumentsUpdate(ctx context.Context) {
	interval, err := time.ParseDuration(env.GetBybitInstrumentsUpdateInterval())
	if err != nil {
		interval = 5 * time.Minute // значение по умолчанию
		logger.LogError("Failed to parse BYBIT_INSTRUMENTS_UPDATE_INTERVAL, using default: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.updateInstruments(ctx); err != nil {
				logger.LogError("Error updating instruments: %v", err)
			}
		}
	}
}

func (s *BybitService) updateInstruments(ctx context.Context) error {
	response, err := s.bybitClient.GetInstruments(ctx, "spot")
	if err != nil {
		logger.LogError("Failed to get instruments from Bybit: %v", err)
		return err
	}

	var instruments []models.BybitInstrument
	for _, instrument := range response.List {
		if instrument.Symbol == "BTCUSDT" {
			//logger.LogInfo("MinOrderQty: from %v to %v", instrument.LotSizeFilter.MinOrderQty, decimal.RequireFromString(instrument.LotSizeFilter.MinOrderQty))
		}
		instruments = append(instruments, models.BybitInstrument{
			Symbol:           instrument.Symbol,
			Category:         response.Category,
			BaseCoin:         instrument.BaseCoin,
			QuoteCoin:        instrument.QuoteCoin,
			Innovation:       instrument.Innovation,
			Status:           instrument.Status,
			MarginTrading:    instrument.MarginTrading,
			StTag:            instrument.StTag,
			BasePrecision:    decimal.RequireFromString(instrument.LotSizeFilter.BasePrecision),
			QuotePrecision:   decimal.RequireFromString(instrument.LotSizeFilter.QuotePrecision),
			MinOrderQty:      decimal.RequireFromString(instrument.LotSizeFilter.MinOrderQty),
			MaxOrderQty:      decimal.RequireFromString(instrument.LotSizeFilter.MaxOrderQty),
			MinOrderAmt:      decimal.RequireFromString(instrument.LotSizeFilter.MinOrderAmt),
			MaxOrderAmt:      decimal.RequireFromString(instrument.LotSizeFilter.MaxOrderAmt),
			TickSize:         decimal.RequireFromString(instrument.PriceFilter.TickSize),
			PriceLimitRatioX: decimal.RequireFromString(instrument.RiskParameters.PriceLimitRatioX),
			PriceLimitRatioY: decimal.RequireFromString(instrument.RiskParameters.PriceLimitRatioY),
		})
	}

	if err := s.bybitInstrumentRepo.SaveInstruments(ctx, instruments); err != nil {
		logger.LogError("Failed to save instruments: %v", err)
		return err
	}

	logger.LogInfo("Successfully updated %d instruments", len(instruments))
	return nil
}

func (s *BybitService) getBybitAccount(ctx context.Context, userID string) (*bybit.BybitAccount, error) {
	var account bybit.BybitAccount
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, api_key, api_secret, account_type, is_active 
		FROM bybit_accounts 
		WHERE user_id = $1 AND is_active = true AND deleted_at IS NULL`,
		userID,
	).Scan(
		&account.ID,
		&account.UserID,
		&account.APIKey,
		&account.APISecret,
		&account.AccountType,
		&account.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("аккаунт Bybit не найден")
		}
		return nil, err
	}

	return &account, nil
}

// parseDecimal преобразует строку в decimal.Decimal
func parseDecimal(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}

// getPrecision возвращает количество знаков после запятой
func getPrecision(s string) int {
	if s == "" {
		return 0
	}
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return 0
	}
	return len(parts[1])
}

func (s *BybitService) UpdateInstruments(ctx context.Context) error {
	// Получаем список инструментов из API
	response, err := s.bybitClient.GetInstruments(ctx, "spot")
	if err != nil {
		return fmt.Errorf("failed to get instruments: %w", err)
	}

	// Преобразуем инструменты в модель
	var modelInstruments []models.BybitInstrument
	for _, inst := range response.List {
		modelInstruments = append(modelInstruments, models.BybitInstrument{
			Symbol:           inst.Symbol,
			Category:         response.Category,
			BaseCoin:         inst.BaseCoin,
			QuoteCoin:        inst.QuoteCoin,
			Innovation:       inst.Innovation,
			Status:           inst.Status,
			MarginTrading:    inst.MarginTrading,
			StTag:            inst.StTag,
			BasePrecision:    decimal.RequireFromString(inst.LotSizeFilter.BasePrecision),
			QuotePrecision:   decimal.RequireFromString(inst.LotSizeFilter.QuotePrecision),
			MinOrderQty:      decimal.RequireFromString(inst.LotSizeFilter.MinOrderQty),
			MaxOrderQty:      decimal.RequireFromString(inst.LotSizeFilter.MaxOrderQty),
			MinOrderAmt:      decimal.RequireFromString(inst.LotSizeFilter.MinOrderAmt),
			MaxOrderAmt:      decimal.RequireFromString(inst.LotSizeFilter.MaxOrderAmt),
			TickSize:         decimal.RequireFromString(inst.PriceFilter.TickSize),
			PriceLimitRatioX: decimal.RequireFromString(inst.RiskParameters.PriceLimitRatioX),
			PriceLimitRatioY: decimal.RequireFromString(inst.RiskParameters.PriceLimitRatioY),
		})
	}

	// Сохраняем инструменты в базу данных
	if err := s.bybitInstrumentRepo.SaveInstruments(ctx, modelInstruments); err != nil {
		return fmt.Errorf("failed to save instruments: %w", err)
	}

	return nil
}

// StartWebSocket запускает WebSocket-соединение и подписку на каналы
func (s *BybitService) StartWebSocket(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Получаем активные инструменты через репозиторий
				activeSymbols, err := s.userInstrumentRepo.GetActiveInstruments(ctx)
				if err != nil {
					logger.LogError("Failed to get active instruments: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Формируем каналы для подписки
				var publicChannels []string
				for _, symbol := range activeSymbols {
					publicChannels = append(publicChannels,
						fmt.Sprintf("tickers.%s", symbol),
						fmt.Sprintf("orderbook.50.%s", symbol),
						fmt.Sprintf("publicTrade.%s", symbol),
					)
				}

				// Логируем каналы
				if len(publicChannels) == 0 {
					logger.LogInfo("Нет активных инструментов для подписки")
					time.Sleep(5 * time.Second)
					continue
				}
				logger.LogInfo("Подписываемся на каналы для %d активных инструментов: %v", len(activeSymbols), publicChannels)

				// Подключаемся к WebSocket
				if err := s.wsClient.Connect(ctx); err != nil {
					logger.LogError("Failed to connect to WebSocket: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Запускаем обработку сообщений
				s.wsClient.StartMessageHandler(ctx, s.wsHandler.HandleMessage)

				// Подписываемся на публичные каналы
				if err := s.wsClient.Subscribe(ctx, publicChannels); err != nil {
					logger.LogError("Failed to subscribe to public channels: %v", err)
					s.wsClient.Close()
					time.Sleep(5 * time.Second)
					continue
				}

				// Логируем успешную подписку
				logger.LogInfo("Успешно подписались на %d каналов для %d активных инструментов", len(publicChannels), len(activeSymbols))

				// Ждем завершения контекста
				<-ctx.Done()
				return
			}
		}
	}()
}

// StartPrivateWebSocket запускает приватные WebSocket-соединения для активных аккаунтов
func (s *BybitService) StartPrivateWebSocket(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.closePrivateWebSockets()
				return
			default:
				// Получаем активные аккаунты Bybit
				accounts, err := s.getActiveBybitAccounts(ctx)
				if err != nil {
					logger.LogError("Failed to get active Bybit accounts: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				s.wsMutex.Lock()
				// Закрываем соединения для неактивных аккаунтов
				for userID := range s.privateWsClients {
					if !s.isAccountActive(userID, accounts) {
						if client, exists := s.privateWsClients[userID]; exists {
							client.Close()
							delete(s.privateWsClients, userID)
							logger.LogInfo("Закрыто приватное WebSocket-соединение для userID: %s", userID)
						}
					}
				}

				// Создаем соединения для активных аккаунтов
				apiMode := env.GetBybitApiMode()
				privateWsURL := env.GetBybitWsUrl() + "/v5/private"
				if apiMode == "test" {
					privateWsURL = env.GetBybitWsTestUrl() + "/v5/private"
				}
				recvWindow, _ := strconv.Atoi(env.GetBybitRecvWindow())

				for _, account := range accounts {
					if _, exists := s.privateWsClients[account.UserID]; !exists {
						wsClient := bybit.NewWebSocketClient(privateWsURL, recvWindow, account.APIKey, account.APISecret)
						s.privateWsClients[account.UserID] = wsClient

						// Подключаемся и подписываемся
						if err := wsClient.Connect(ctx); err != nil {
							logger.LogError("Failed to connect to private WebSocket for userID %s: %v", account.UserID, err)
							delete(s.privateWsClients, account.UserID)
							continue
						}

						privateChannels := []string{
							"order.spot",
							"execution.spot",
							//"execution.fast.spot",
							"wallet",
						}
						wsClient.StartMessageHandler(ctx, s.wsHandler.HandlePrivateMessage)

						if err := wsClient.Subscribe(ctx, privateChannels); err != nil {
							logger.LogError("Failed to subscribe to private channels for userID %s: %v", account.UserID, err)
							wsClient.Close()
							delete(s.privateWsClients, account.UserID)
							continue
						}

						logger.LogInfo("Успешно подключились к приватному WebSocket для userID: %s", account.UserID)
					}
				}
				s.wsMutex.Unlock()

				time.Sleep(30 * time.Second) // Проверяем аккаунты каждые 30 секунд
			}
		}
	}()
}

// getActiveBybitAccounts получает все активные аккаунты Bybit
func (s *BybitService) getActiveBybitAccounts(ctx context.Context) ([]bybit.BybitAccount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, api_key, api_secret, account_type, is_active 
		FROM bybit_accounts 
		WHERE is_active = true AND deleted_at IS NULL`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query active accounts: %w", err)
	}
	defer rows.Close()

	var accounts []bybit.BybitAccount
	for rows.Next() {
		var account bybit.BybitAccount
		if err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.APIKey,
			&account.APISecret,
			&account.AccountType,
			&account.IsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// isAccountActive проверяет, активен ли аккаунт
func (s *BybitService) isAccountActive(userID string, accounts []bybit.BybitAccount) bool {
	for _, account := range accounts {
		if account.UserID == userID {
			return true
		}
	}
	return false
}

// closePrivateWebSockets закрывает все приватные WebSocket-соединения
func (s *BybitService) closePrivateWebSockets() {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	for userID, client := range s.privateWsClients {
		client.Close()
		delete(s.privateWsClients, userID)
		logger.LogInfo("Закрыто приватное WebSocket-соединение для userID: %s", userID)
	}
}
