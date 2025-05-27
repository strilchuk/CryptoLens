package services

import (
	"CryptoLens_Backend/models"
	"CryptoLens_Backend/repositories"
	"context"
	"database/sql"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type UserService struct {
	userRepo *repositories.UserRepository
	jwtKey   []byte
	db       *sql.DB
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewUserService(userRepo *repositories.UserRepository, jwtKey []byte, db *sql.DB) *UserService {
	return &UserService{
		userRepo: userRepo,
		jwtKey:   jwtKey,
		db:       db,
	}
}

func (s *UserService) Register(ctx context.Context, req models.RegisterRequest) (*models.RegisterResponse, error) {
	// Проверяем, существует ли пользователь с таким email
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	// Получаем ID типа пользователя "user"
	var userTypeID string
	err = s.db.QueryRowContext(ctx, "SELECT id FROM user_types WHERE name = 'user'").Scan(&userTypeID)
	if err != nil {
		return nil, errors.New("failed to get user type")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаем пользователя
	user := &models.User{
		Nickname:   req.Nickname,
		Email:      req.Email,
		Password:   string(hashedPassword),
		UserTypeID: userTypeID,
		CreatedAt:  &time.Time{},
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Генерируем JWT токен
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &models.RegisterResponse{
		User:  *user,
		Token: token,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		AccessToken: token,
		TokenType:   "bearer",
		ExpiresIn:   3600,
	}, nil
}

func (s *UserService) Logout(ctx context.Context, token string) (*models.LogoutResponse, error) {
	// В данном случае просто возвращаем успешный ответ
	// В реальном приложении здесь можно добавить токен в черный список
	return &models.LogoutResponse{
		Status:  "Success",
		Message: "Successfully logged out",
	}, nil
}

func (s *UserService) GetAccount(ctx context.Context, token string) (*models.User, error) {
	// Валидируем токен и получаем ID пользователя
	userID, err := s.validateToken(token)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) generateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "crypto-lens",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtKey)
}

func (s *UserService) validateToken(tokenString string) (string, error) {
	// Убираем префикс "Bearer " если он есть
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
} 