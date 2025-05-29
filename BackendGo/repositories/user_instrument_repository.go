package repositories

import (
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/models"
	"context"
	"database/sql"
	"errors"
)

type UserInstrumentRepository struct {
	db *sql.DB
}

func NewUserInstrumentRepository(db *sql.DB) *UserInstrumentRepository {
	return &UserInstrumentRepository{db: db}
}

// Create создает новую связь пользователя с инструментом
func (r *UserInstrumentRepository) Create(ctx context.Context, userID string, symbol string) (*models.UserInstrument, error) {
	query := `
		INSERT INTO user_instruments (user_id, symbol)
		VALUES ($1, $2)
		RETURNING id, user_id, symbol, is_active, created_at, updated_at`

	instrument := &models.UserInstrument{}
	err := r.db.QueryRowContext(ctx, query, userID, symbol).Scan(
		&instrument.ID,
		&instrument.UserID,
		&instrument.Symbol,
		&instrument.IsActive,
		&instrument.CreatedAt,
		&instrument.UpdatedAt,
	)

	if err != nil {
		logger.LogError("Failed to create user instrument: %v", err)
		return nil, err
	}

	return instrument, nil
}

// GetByUserID получает все инструменты пользователя
func (r *UserInstrumentRepository) GetByUserID(ctx context.Context, userID string) ([]models.UserInstrument, error) {
	query := `
		SELECT id, user_id, symbol, is_active, created_at, updated_at
		FROM user_instruments
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.LogError("Failed to get user instruments: %v", err)
		return nil, err
	}
	defer rows.Close()

	var instruments []models.UserInstrument
	for rows.Next() {
		var instrument models.UserInstrument
		err := rows.Scan(
			&instrument.ID,
			&instrument.UserID,
			&instrument.Symbol,
			&instrument.IsActive,
			&instrument.CreatedAt,
			&instrument.UpdatedAt,
		)
		if err != nil {
			logger.LogError("Failed to scan user instrument: %v", err)
			return nil, err
		}
		instruments = append(instruments, instrument)
	}

	return instruments, nil
}

// Update обновляет статус инструмента пользователя
func (r *UserInstrumentRepository) Update(ctx context.Context, id string, isActive bool) error {
	query := `
		UPDATE user_instruments
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, isActive, id)
	if err != nil {
		logger.LogError("Failed to update user instrument: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("user instrument not found")
	}

	return nil
}

// Delete мягко удаляет связь пользователя с инструментом
func (r *UserInstrumentRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE user_instruments
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.LogError("Failed to delete user instrument: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("user instrument not found")
	}

	return nil
}

// Exists проверяет существование связи пользователя с инструментом
func (r *UserInstrumentRepository) Exists(ctx context.Context, userID string, symbol string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_instruments
			WHERE user_id = $1 AND symbol = $2 AND deleted_at IS NULL
		)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, symbol).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check user instrument existence: %v", err)
		return false, err
	}

	return exists, nil
}

// GetActiveInstruments получает все активные инструменты пользователей
func (r *UserInstrumentRepository) GetActiveInstruments(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT symbol 
		FROM user_instruments 
		WHERE is_active = true 
		AND deleted_at IS NULL`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.LogError("Failed to get active instruments: %v", err)
		return nil, err
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			logger.LogError("Failed to scan symbol: %v", err)
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	if err = rows.Err(); err != nil {
		logger.LogError("Error iterating active instruments: %v", err)
		return nil, err
	}

	return symbols, nil
} 