package services

import (
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/models"
	"CryptoLens_Backend/repositories"
	"context"
	"errors"
)

type UserInstrumentService struct {
	userInstrumentRepo *repositories.UserInstrumentRepository
	bybitInstrumentRepo *repositories.BybitInstrumentRepository
}

func NewUserInstrumentService(
	userInstrumentRepo *repositories.UserInstrumentRepository,
	bybitInstrumentRepo *repositories.BybitInstrumentRepository,
) *UserInstrumentService {
	return &UserInstrumentService{
		userInstrumentRepo: userInstrumentRepo,
		bybitInstrumentRepo: bybitInstrumentRepo,
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
	return s.userInstrumentRepo.Create(ctx, userID, symbol)
}

// GetUserInstruments получает все инструменты пользователя
func (s *UserInstrumentService) GetUserInstruments(ctx context.Context, userID string) ([]models.UserInstrument, error) {
	return s.userInstrumentRepo.GetByUserID(ctx, userID)
}

// UpdateInstrumentStatus обновляет статус инструмента пользователя
func (s *UserInstrumentService) UpdateInstrumentStatus(ctx context.Context, id string, isActive bool) error {
	return s.userInstrumentRepo.Update(ctx, id, isActive)
}

// RemoveInstrument удаляет инструмент у пользователя
func (s *UserInstrumentService) RemoveInstrument(ctx context.Context, id string) error {
	return s.userInstrumentRepo.Delete(ctx, id)
} 