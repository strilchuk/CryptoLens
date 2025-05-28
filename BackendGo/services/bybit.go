package services

import (
	"CryptoLens_Backend/integration/bybit"
	"context"
	"database/sql"
	"errors"
)

type BybitService struct {
	bybitClient bybit.Client
	db          *sql.DB
	userService *UserService
}

func NewBybitService(bybitClient bybit.Client, db *sql.DB, userService *UserService) *BybitService {
	return &BybitService{
		bybitClient: bybitClient,
		db:          db,
		userService: userService,
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
