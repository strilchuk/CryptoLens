package services

import (
	"CryptoLens_Backend/models"
	"CryptoLens_Backend/repositories"
	"CryptoLens_Backend/trading"
	"context"
	"errors"
)

type UserStrategyService struct {
	userStrategyRepo *repositories.UserStrategyRepository
	strategyManager  *trading.StrategyManager
}

func NewUserStrategyService(
	userStrategyRepo *repositories.UserStrategyRepository,
	strategyManager *trading.StrategyManager,
) *UserStrategyService {
	return &UserStrategyService{
		userStrategyRepo: userStrategyRepo,
		strategyManager:  strategyManager,
	}
}

func (s *UserStrategyService) AddStrategy(ctx context.Context, userID string, strategyName string) (*models.UserStrategy, error) {
	// Проверяем, существует ли уже такая стратегия у пользователя
	exists, err := s.userStrategyRepo.Exists(ctx, userID, strategyName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("стратегия уже добавлена")
	}

	// Создаем запись в БД
	strategy, err := s.userStrategyRepo.Create(ctx, userID, strategyName)
	if err != nil {
		return nil, err
	}

	// Создаем и запускаем стратегию
	switch strategyName {
	case "test":
		testStrategy := trading.NewTestStrategy(userID)
		s.strategyManager.AddStrategy(userID, testStrategy)
		go testStrategy.Start(ctx)
	default:
		return nil, errors.New("неизвестная стратегия")
	}

	return strategy, nil
}

func (s *UserStrategyService) GetUserStrategies(ctx context.Context, userID string) ([]models.UserStrategy, error) {
	return s.userStrategyRepo.GetByUserID(ctx, userID)
}

func (s *UserStrategyService) UpdateStrategyStatus(ctx context.Context, id string, isActive bool) error {
	return s.userStrategyRepo.Update(ctx, id, isActive)
}

func (s *UserStrategyService) RemoveStrategy(ctx context.Context, id string) error {
	return s.userStrategyRepo.Delete(ctx, id)
}

func (s *UserStrategyService) GetActiveStrategies(ctx context.Context) ([]models.UserStrategy, error) {
	return s.userStrategyRepo.GetActiveStrategies(ctx)
}
