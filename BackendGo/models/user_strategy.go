package models

import "time"

type UserStrategy struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	StrategyName string     `json:"strategy_name" db:"strategy_name"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at" db:"deleted_at"`
}

type CreateUserStrategyRequest struct {
	StrategyName string `json:"strategy_name" validate:"required"`
}

type UpdateUserStrategyRequest struct {
	ID       string `json:"id" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
}

type UserStrategyResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	StrategyName string     `json:"strategy_name"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
} 