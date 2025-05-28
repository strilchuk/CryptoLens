package services

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"strings"
)

// ValidateToken проверяет JWT токен и возвращает ID пользователя
func ValidateToken(tokenString string) (string, error) {
	// Убираем префикс "Bearer " если он есть
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}

// jwtKey используется для подписи и проверки токенов
var jwtKey []byte

// SetJWTKey устанавливает ключ для JWT
func SetJWTKey(key []byte) {
	jwtKey = key
} 