package types

import (
	"CryptoLens_Backend/models"
	"context"
)

type UserStrategyServiceInterface interface {
	AddStrategy(ctx context.Context, userID string, strategyName string) (*models.UserStrategy, error)
	GetUserStrategies(ctx context.Context, userID string) ([]models.UserStrategy, error)
	UpdateStrategyStatus(ctx context.Context, id string, isActive bool) error
	RemoveStrategy(ctx context.Context, id string) error
	GetActiveStrategies(ctx context.Context) ([]models.UserStrategy, error)
	LoadActiveStrategies(ctx context.Context) error
}
