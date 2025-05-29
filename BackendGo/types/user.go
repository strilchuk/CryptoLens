package types

import (
	"CryptoLens_Backend/models"
	"context"
)

type UserServiceInterface interface {
	Register(ctx context.Context, req models.RegisterRequest) (*models.RegisterResponse, error)
	Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error)
	Logout(ctx context.Context, token string) (*models.LogoutResponse, error)
	GetAccount(ctx context.Context, token string) (*models.User, error)
}

type UserInstrumentServiceInterface interface {
	AddInstrument(ctx context.Context, userID string, symbol string) (*models.UserInstrument, error)
	GetUserInstruments(ctx context.Context, userID string) ([]models.UserInstrument, error)
	UpdateInstrumentStatus(ctx context.Context, id string, isActive bool) error
	RemoveInstrument(ctx context.Context, id string) error
} 