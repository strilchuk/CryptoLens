package services

import (
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/models"
	"CryptoLens_Backend/repositories"
	"CryptoLens_Backend/types"
	"context"
	"errors"
	"fmt"
)

type UserInstrumentService struct {
	userInstrumentRepo *repositories.UserInstrumentRepository
	bybitInstrumentRepo *repositories.BybitInstrumentRepository
	strategyManager types.StrategyManagerInterface
}

func NewUserInstrumentService(
	userInstrumentRepo *repositories.UserInstrumentRepository,
	bybitInstrumentRepo *repositories.BybitInstrumentRepository,
	strategyManager types.StrategyManagerInterface,
) *UserInstrumentService {
	return &UserInstrumentService{
		userInstrumentRepo: userInstrumentRepo,
		bybitInstrumentRepo: bybitInstrumentRepo,
		strategyManager: strategyManager,
	}
}

// AddInstrument добавляет инструмент для пользователя
func (s *UserInstrumentService) AddInstrument(ctx context.Context, userID string, symbol string) (*models.UserInstrument, error) {
	// Проверяем существование инструмента в Bybit
	exists, err := s.bybitInstrumentRepo.Exists(ctx, symbol)
	if err != nil {
		logger.LogError("Failed to check instrument existence: %v", err)
		return nil, err
	}
	if !exists {
		return nil, errors.New("instrument not found in Bybit")
	}

	// Проверяем, не добавлен ли уже этот инструмент пользователю
	exists, err = s.userInstrumentRepo.Exists(ctx, userID, symbol)
	if err != nil {
		logger.LogError("Failed to check user instrument existence: %v", err)
		return nil, err
	}
	if exists {
		return nil, errors.New("instrument already added for this user")
	}

	// Создаем связь пользователя с инструментом
	instrument, err := s.userInstrumentRepo.Create(ctx, userID, symbol)
	if err != nil {
		return nil, err
	}

	// Обновляем символы в StrategyManager
	if err := s.strategyManager.UpdateUserInstruments(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to update user instruments in strategy manager: %w", err)
	}

	return instrument, nil
}

// GetUserInstruments получает все инструменты пользователя
func (s *UserInstrumentService) GetUserInstruments(ctx context.Context, userID string) ([]models.UserInstrument, error) {
	instruments, err := s.userInstrumentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Получаем данные по каждому инструменту из таблицы bybit_instruments
	for i := range instruments {
		bybitInstrument, err := s.bybitInstrumentRepo.GetBySymbol(ctx, instruments[i].Symbol)
		if err != nil {
			logger.LogError("Failed to get bybit instrument for symbol %s: %v", instruments[i].Symbol, err)
			continue
		}
		instruments[i].BybitInstrument = bybitInstrument
	}

	return instruments, nil
}

// UpdateInstrumentStatus обновляет статус инструмента пользователя
func (s *UserInstrumentService) UpdateInstrumentStatus(ctx context.Context, id string, isActive bool) error {
	if err := s.userInstrumentRepo.Update(ctx, id, isActive); err != nil {
		return err
	}

	// Получаем userID для обновления StrategyManager
	instrument, err := s.userInstrumentRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get instrument for update: %w", err)
	}

	// Обновляем символы в StrategyManager
	if err := s.strategyManager.UpdateUserInstruments(ctx, instrument.UserID); err != nil {
		return fmt.Errorf("failed to update user instruments in strategy manager: %w", err)
	}

	return nil
}

// RemoveInstrument удаляет инструмент у пользователя
func (s *UserInstrumentService) RemoveInstrument(ctx context.Context, id string) error {
	// Получаем userID перед удалением
	instrument, err := s.userInstrumentRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get instrument for removal: %w", err)
	}

	if err := s.userInstrumentRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Обновляем символы в StrategyManager
	if err := s.strategyManager.UpdateUserInstruments(ctx, instrument.UserID); err != nil {
		return fmt.Errorf("failed to update user instruments in strategy manager: %w", err)
	}

	return nil
} 