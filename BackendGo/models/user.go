package models

import (
	"time"
)

type UserType struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

type User struct {
	ID              string     `json:"id"`
	UserTypeID      string     `json:"user_type_id"`
	Nickname        string     `json:"nickname"`
	Email           string     `json:"email"`
	Password        string     `json:"-"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       *time.Time `json:"-"`
	UpdatedAt       *time.Time `json:"-"`
	DeletedAt       *time.Time `json:"-"`
}

type RegisterRequest struct {
	Nickname            string `json:"nickname" validate:"required,min=3,max=255"`
	Email              string `json:"email" validate:"required,email"`
	Password           string `json:"password" validate:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type LogoutResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
} 