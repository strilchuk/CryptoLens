package types

import (
	"CryptoLens_Backend/models"
	"context"
)

type BybitInstrumentRepositoryInterface interface {
	GetBySymbol(ctx context.Context, symbol string) (*models.BybitInstrument, error)
}

type UserStrategyRepositoryInterface interface {
	Create(ctx context.Context, userID string, strategyName string) (*models.UserStrategy, error)
	GetByID(ctx context.Context, id string) (*models.UserStrategy, error)
	GetByUserID(ctx context.Context, userID string) ([]models.UserStrategy, error)
	Update(ctx context.Context, id string, isActive bool) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, userID string, strategyName string) (bool, error)
	GetActiveStrategies(ctx context.Context) ([]models.UserStrategy, error)
	DeactivateAllStrategies(ctx context.Context) error
}
