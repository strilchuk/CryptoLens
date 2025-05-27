package models

import (
	"time"
)

type BybitAccount struct {
	ID          int64      `json:"id"`
	UserID      string     `json:"user_id"`
	APIKey      string     `json:"api_key"`
	APISecret   string     `json:"api_secret"`
	AccountType string     `json:"account_type"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   *time.Time `json:"-"`
	UpdatedAt   *time.Time `json:"-"`
	DeletedAt   *time.Time `json:"-"`
}

type CreateBybitAccountRequest struct {
	APIKey      string `json:"api_key" validate:"required"`
	APISecret   string `json:"api_secret" validate:"required"`
	AccountType string `json:"account_type" validate:"required,oneof=UNIFIED SPOT FUTURES"`
}

type UpdateBybitAccountRequest struct {
	APIKey      string `json:"api_key" validate:"omitempty"`
	APISecret   string `json:"api_secret" validate:"omitempty"`
	AccountType string `json:"account_type" validate:"omitempty,oneof=UNIFIED SPOT FUTURES"`
	IsActive    *bool  `json:"is_active" validate:"omitempty"`
}

type BybitAccountResponse struct {
	ID          int64  `json:"id"`
	UserID      string `json:"user_id"`
	APIKey      string `json:"api_key"`
	AccountType string `json:"account_type"`
	IsActive    bool   `json:"is_active"`
} 