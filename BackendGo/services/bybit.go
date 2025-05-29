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
	db                  *sql.DB
	userService         *UserService
	bybitInstrumentRepo *repositories.BybitInstrumentRepository
	wsHandler           *handlers.BybitWebSocketHandler
	wsMutex             sync.Mutex
}

func NewBybitService(bybitClient bybit.Client, db *sql.DB, userService *UserService) *BybitService {
	// Инициализация WebSocket-клиента
	recvWindow, _ := strconv.Atoi(env.GetBybitRecvWindow())
	apiMode := env.GetBybitApiMode()
	wsURL := env.GetBybitWsUrl() + "/v5/public/spot"
	if apiMode == "test" {
		wsURL = env.GetBybitWsTestUrl() + "/v5/public/spot"
	}
	wsClient := bybit.NewWebSocketClient(wsURL, recvWindow)

	return &BybitService{
		bybitClient:         bybitClient,
		wsClient:            wsClient,
		db:                  db,
		userService:         userService,
		bybitInstrumentRepo: repositories.NewBybitInstrumentRepository(db),
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
		// Получаем активные инструменты пользователей
		instruments, err := s.bybitInstrumentRepo.GetInstruments(ctx, "spot")
		if err != nil {
			logger.LogError("Failed to get instruments for WebSocket: %v", err)
			return
		}

		// Формируем список каналов для подписки
		var publicChannels []string
		for _, inst := range instruments {
			publicChannels = append(publicChannels,
				fmt.Sprintf("ticker.%s", inst.Symbol),
				fmt.Sprintf("orderbook.25.%s", inst.Symbol),
				fmt.Sprintf("trade.%s", inst.Symbol),
			)
		}

		// Подключаемся к WebSocket
		if err := s.wsClient.Connect(ctx); err != nil {
			logger.LogError("Failed to connect to WebSocket: %v", err)
			return
		}

		// Подписываемся на публичные каналы
		if err := s.wsClient.Subscribe(ctx, publicChannels); err != nil {
			logger.LogError("Failed to subscribe to public channels: %v", err)
			return
		}

		// Запускаем обработку сообщений
		s.wsClient.StartMessageHandler(ctx, s.wsHandler.HandleMessage)
	}()
}

// StartBackgroundTasks дополняем для запуска WebSocket
func (s *BybitService) StartBackgroundTasks(ctx context.Context) {
	s.StartInstrumentsUpdate(ctx)
	s.StartWebSocket(ctx)
}
