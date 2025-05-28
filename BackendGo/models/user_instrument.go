package models

import "time"

// UserInstrument представляет связь пользователя с инструментом
type UserInstrument struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Symbol    string     `json:"symbol" db:"symbol"`
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

// CreateUserInstrumentRequest представляет запрос на создание связи
type CreateUserInstrumentRequest struct {
	Symbol string `json:"symbol" validate:"required"`
}

// UpdateUserInstrumentRequest представляет запрос на обновление связи
type UpdateUserInstrumentRequest struct {
	ID       string `json:"id" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
}

// UserInstrumentResponse представляет ответ с информацией о связи
type UserInstrumentResponse struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Symbol    string     `json:"symbol"`
	IsActive  bool       `json:"is_active"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
} 