package repositories

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/types"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// BybitAccountRepository реализует интерфейс BybitAccountRepositoryInterface
type BybitAccountRepository struct {
	db *sql.DB
}

// NewBybitAccountRepository создает новый репозиторий для работы с аккаунтами Bybit
func NewBybitAccountRepository(db *sql.DB) types.BybitAccountRepositoryInterface {
	return &BybitAccountRepository{db: db}
}

// GetActiveAccountByUserID получает активный аккаунт Bybit для пользователя
func (r *BybitAccountRepository) GetActiveAccountByUserID(ctx context.Context, userID string) (*bybit.BybitAccount, error) {
	var account bybit.BybitAccount
	err := r.db.QueryRowContext(ctx,
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
			return nil, fmt.Errorf("Bybit account not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to get Bybit account: %w", err)
	}
	return &account, nil
}

// CreateAccount создает новый аккаунт Bybit для пользователя
func (r *BybitAccountRepository) CreateAccount(ctx context.Context, userID string, apiKey, apiSecret, accountType string) (*bybit.BybitAccount, error) {
	var account bybit.BybitAccount
	now := time.Now()
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO bybit_accounts (user_id, api_key, api_secret, account_type, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, $5, $6)
		RETURNING id, user_id, api_key, api_secret, account_type, is_active`,
		userID, apiKey, apiSecret, accountType, now, now,
	).Scan(
		&account.ID,
		&account.UserID,
		&account.APIKey,
		&account.APISecret,
		&account.AccountType,
		&account.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bybit account: %w", err)
	}
	return &account, nil
}

// UpdateAccount обновляет аккаунт Bybit для пользователя
func (r *BybitAccountRepository) UpdateAccount(ctx context.Context, userID string, apiKey, apiSecret, accountType string, isActive bool) error {
	now := time.Now()
	result, err := r.db.ExecContext(ctx,
		`UPDATE bybit_accounts 
		SET api_key = $1, api_secret = $2, account_type = $3, is_active = $4, updated_at = $5
		WHERE user_id = $6 AND deleted_at IS NULL`,
		apiKey, apiSecret, accountType, isActive, now, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update Bybit account: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("Bybit account not found for user %s", userID)
	}
	return nil
}

// DeleteAccount удаляет аккаунт Bybit для пользователя (soft delete)
func (r *BybitAccountRepository) DeleteAccount(ctx context.Context, userID string) error {
	now := time.Now()
	result, err := r.db.ExecContext(ctx,
		`UPDATE bybit_accounts 
		SET deleted_at = $1, updated_at = $1
		WHERE user_id = $2 AND deleted_at IS NULL`,
		now, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete Bybit account: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("Bybit account not found for user %s", userID)
	}
	return nil
}

// GetActiveAccounts получает все активные аккаунты Bybit
func (r *BybitAccountRepository) GetActiveAccounts(ctx context.Context) ([]bybit.BybitAccount, error) {
	rows, err := r.db.QueryContext(ctx,
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