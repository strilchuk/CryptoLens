package services

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/models"
	"CryptoLens_Backend/repositories"
	"context"
	"database/sql"
	"errors"
	"github.com/shopspring/decimal"
	"log"
	"strings"
	"time"
)

type BybitService struct {
	bybitClient         bybit.Client
	db                  *sql.DB
	userService         *UserService
	bybitInstrumentRepo *repositories.BybitInstrumentRepository
}

func NewBybitService(bybitClient bybit.Client, db *sql.DB, userService *UserService) *BybitService {
	return &BybitService{
		bybitClient:         bybitClient,
		db:                  db,
		userService:         userService,
		bybitInstrumentRepo: repositories.NewBybitInstrumentRepository(db),
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
	ticker := time.NewTicker(1 * time.Minute) // Для тестирования обновляем каждую минуту
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.updateInstruments(ctx); err != nil {
				log.Printf("Error updating instruments: %v", err)
			}
		}
	}
}

func (s *BybitService) updateInstruments(ctx context.Context) error {
	// Получаем инструменты через API Bybit
	response, err := s.bybitClient.GetInstruments(ctx, "spot")
	if err != nil {
		return err
	}

	// Конвертируем в нашу модель
	var dbInstruments []models.BybitInstrument
	for _, inst := range response.List {
		dbInstruments = append(dbInstruments, models.BybitInstrument{
			Symbol:        inst.Symbol,
			Category:      response.Category,
			BaseCoin:      inst.BaseCoin,
			QuoteCoin:     inst.QuoteCoin,
			MinOrderQty:   parseDecimal(inst.LotSizeFilter.MinOrderQty),
			MaxOrderQty:   parseDecimal(inst.LotSizeFilter.MaxOrderQty),
			MinPrice:      parseDecimal(inst.PriceFilter.TickSize),
			MaxPrice:      decimal.Zero, // TODO: Добавить в API
			PriceScale:    getPrecision(inst.LotSizeFilter.QuotePrecision),
			QuantityScale: getPrecision(inst.LotSizeFilter.BasePrecision),
			Status:        inst.Status,
		})
	}

	return s.bybitInstrumentRepo.UpsertInstruments(ctx, dbInstruments)
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
