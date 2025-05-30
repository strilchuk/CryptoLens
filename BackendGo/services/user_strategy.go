package services

import (
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/models"
	"CryptoLens_Backend/repositories"
	"CryptoLens_Backend/trading"
	"context"
	"errors"
	"fmt"
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

	// Удаляем все существующие стратегии пользователя из менеджера
	strategies := s.strategyManager.GetStrategies(userID)
	for _, st := range strategies {
		s.strategyManager.RemoveStrategy(userID, st)
	}

	// Получаем все активные стратегии пользователя
	activeStrategies, err := s.userStrategyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении активных стратегий: %w", err)
	}

	// Создаем и добавляем все активные стратегии
	for _, st := range activeStrategies {
		if st.IsActive {
			switch st.StrategyName {
			case "test":
				testStrategy := trading.NewTestStrategy(st.UserID)
				s.strategyManager.AddStrategy(st.UserID, testStrategy)
				go testStrategy.Start(ctx)
			default:
				logger.LogError("Неизвестная стратегия при добавлении: %s", st.StrategyName)
			}
		}
	}

	return strategy, nil
}

func (s *UserStrategyService) GetUserStrategies(ctx context.Context, userID string) ([]models.UserStrategy, error) {
	return s.userStrategyRepo.GetByUserID(ctx, userID)
}

func (s *UserStrategyService) UpdateStrategyStatus(ctx context.Context, id string, isActive bool) error {
	// Получаем информацию о стратегии перед обновлением
	strategy, err := s.userStrategyRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("ошибка при получении стратегии: %w", err)
	}

	// Обновляем статус в БД
	if err := s.userStrategyRepo.Update(ctx, id, isActive); err != nil {
		return err
	}

	// Обновляем состояние в StrategyManager
	if isActive {
		// Если стратегия активируется, создаем и добавляем её
		switch strategy.StrategyName {
		case "test":
			testStrategy := trading.NewTestStrategy(strategy.UserID)
			s.strategyManager.AddStrategy(strategy.UserID, testStrategy)
			go testStrategy.Start(ctx)
		default:
			logger.LogError("Неизвестная стратегия при активации: %s", strategy.StrategyName)
		}
	} else {
		// Если стратегия деактивируется, удаляем все стратегии пользователя
		// и создаем заново только активные
		strategies := s.strategyManager.GetStrategies(strategy.UserID)
		for _, st := range strategies {
			s.strategyManager.RemoveStrategy(strategy.UserID, st)
		}
		
		// Получаем все активные стратегии пользователя
		activeStrategies, err := s.userStrategyRepo.GetByUserID(ctx, strategy.UserID)
		if err != nil {
			return fmt.Errorf("ошибка при получении активных стратегий: %w", err)
		}

		// Создаем и добавляем активные стратегии
		for _, st := range activeStrategies {
			if st.IsActive {
				switch st.StrategyName {
				case "test":
					testStrategy := trading.NewTestStrategy(st.UserID)
					s.strategyManager.AddStrategy(st.UserID, testStrategy)
					go testStrategy.Start(ctx)
				}
			}
		}
	}

	logger.LogInfo("Конечное состояние стратегий в менеджере: %+v", s.strategyManager.GetStrategiesInfo())
	return nil
}

func (s *UserStrategyService) RemoveStrategy(ctx context.Context, id string) error {
	// Получаем информацию о стратегии перед удалением
	strategy, err := s.userStrategyRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("ошибка при получении стратегии: %w", err)
	}

	// Удаляем стратегию из БД
	if err := s.userStrategyRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Удаляем все стратегии пользователя и создаем заново только активные
	strategies := s.strategyManager.GetStrategies(strategy.UserID)
	for _, st := range strategies {
		s.strategyManager.RemoveStrategy(strategy.UserID, st)
	}
	
	// Получаем все активные стратегии пользователя
	activeStrategies, err := s.userStrategyRepo.GetByUserID(ctx, strategy.UserID)
	if err != nil {
		return fmt.Errorf("ошибка при получении активных стратегий: %w", err)
	}

	// Создаем и добавляем активные стратегии
	for _, st := range activeStrategies {
		if st.IsActive {
			switch st.StrategyName {
			case "test":
				testStrategy := trading.NewTestStrategy(st.UserID)
				s.strategyManager.AddStrategy(st.UserID, testStrategy)
				go testStrategy.Start(ctx)
			}
		}
	}

	logger.LogInfo("Конечное состояние стратегий в менеджере: %+v", s.strategyManager.GetStrategiesInfo())
	return nil
}

func (s *UserStrategyService) GetActiveStrategies(ctx context.Context) ([]models.UserStrategy, error) {
	return s.userStrategyRepo.GetActiveStrategies(ctx)
}

func (s *UserStrategyService) LoadActiveStrategies(ctx context.Context) error {
	// Получаем все активные стратегии из БД
	strategies, err := s.userStrategyRepo.GetActiveStrategies(ctx)
	if err != nil {
		return fmt.Errorf("ошибка при получении активных стратегий: %w", err)
	}

	// Для каждой стратегии создаем соответствующий экземпляр и добавляем в менеджер
	for _, strategy := range strategies {
		switch strategy.StrategyName {
		case "test":
			testStrategy := trading.NewTestStrategy(strategy.UserID)
			s.strategyManager.AddStrategy(strategy.UserID, testStrategy)
			go testStrategy.Start(ctx)
		default:
			logger.LogError("Неизвестная стратегия при загрузке: %s", strategy.StrategyName)
		}
	}

	return nil
}
