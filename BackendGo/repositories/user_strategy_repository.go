package repositories

import (
	"CryptoLens_Backend/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type UserStrategyRepository struct {
	db *sql.DB
}

func NewUserStrategyRepository(db *sql.DB) *UserStrategyRepository {
	return &UserStrategyRepository{db: db}
}

func (r *UserStrategyRepository) Create(ctx context.Context, userID string, strategyName string) (*models.UserStrategy, error) {
	query := `
		INSERT INTO user_strategies (user_id, strategy_name, is_active)
		VALUES ($1, $2, false)
		RETURNING id, user_id, strategy_name, is_active, created_at, updated_at`

	var strategy models.UserStrategy
	err := r.db.QueryRowContext(ctx, query, userID, strategyName).Scan(
		&strategy.ID,
		&strategy.UserID,
		&strategy.StrategyName,
		&strategy.IsActive,
		&strategy.CreatedAt,
		&strategy.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &strategy, nil
}

func (r *UserStrategyRepository) GetByUserID(ctx context.Context, userID string) ([]models.UserStrategy, error) {
	query := `
		SELECT id, user_id, strategy_name, is_active, created_at, updated_at
		FROM user_strategies
		WHERE user_id = $1 AND deleted_at IS NULL`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strategies []models.UserStrategy
	for rows.Next() {
		var strategy models.UserStrategy
		err := rows.Scan(
			&strategy.ID,
			&strategy.UserID,
			&strategy.StrategyName,
			&strategy.IsActive,
			&strategy.CreatedAt,
			&strategy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		strategies = append(strategies, strategy)
	}

	return strategies, nil
}

func (r *UserStrategyRepository) Update(ctx context.Context, id string, isActive bool) error {
	query := `
		UPDATE user_strategies
		SET is_active = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, isActive, time.Now(), id)
	return err
}

func (r *UserStrategyRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE user_strategies
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *UserStrategyRepository) Exists(ctx context.Context, userID string, strategyName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_strategies
			WHERE user_id = $1 AND strategy_name = $2 AND deleted_at IS NULL
		)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, strategyName).Scan(&exists)
	return exists, err
}

func (r *UserStrategyRepository) GetActiveStrategies(ctx context.Context) ([]models.UserStrategy, error) {
	query := `
		SELECT id, user_id, strategy_name, is_active, created_at, updated_at
		FROM user_strategies
		WHERE is_active = true AND deleted_at IS NULL`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strategies []models.UserStrategy
	for rows.Next() {
		var strategy models.UserStrategy
		err := rows.Scan(
			&strategy.ID,
			&strategy.UserID,
			&strategy.StrategyName,
			&strategy.IsActive,
			&strategy.CreatedAt,
			&strategy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		strategies = append(strategies, strategy)
	}

	return strategies, nil
}

// GetByID получает стратегию по ID
func (r *UserStrategyRepository) GetByID(ctx context.Context, id string) (*models.UserStrategy, error) {
	var strategy models.UserStrategy
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, strategy_name, is_active, created_at, updated_at 
		FROM user_strategies 
		WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(
		&strategy.ID,
		&strategy.UserID,
		&strategy.StrategyName,
		&strategy.IsActive,
		&strategy.CreatedAt,
		&strategy.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("стратегия не найдена")
		}
		return nil, err
	}

	return &strategy, nil
}

func (r *UserStrategyRepository) DeactivateAllStrategies(ctx context.Context) error {
	query := `UPDATE user_strategies SET is_active = false`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("ошибка при деактивации всех стратегий: %w", err)
	}
	return nil
}
