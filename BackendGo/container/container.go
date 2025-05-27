package container

import (
	"CryptoLens_Backend/handlers"
	"CryptoLens_Backend/repositories"
	"CryptoLens_Backend/routes"
	"CryptoLens_Backend/services"
	"database/sql"
)

type Container struct {
	DB          *sql.DB
	UserRepo    *repositories.UserRepository
	UserService *services.UserService
	UserHandler *handlers.UserHandler
	UserRoutes  *routes.UserRoutes
}

func NewContainer(db *sql.DB, jwtKey []byte) *Container {
	// Инициализация репозиториев
	userRepo := repositories.NewUserRepository(db)

	// Инициализация сервисов
	userService := services.NewUserService(userRepo, jwtKey, db)

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(userService)

	// Инициализация маршрутов
	userRoutes := routes.NewUserRoutes(userHandler)

	return &Container{
		DB:          db,
		UserRepo:    userRepo,
		UserService: userService,
		UserHandler: userHandler,
		UserRoutes:  userRoutes,
	}
}

func (c *Container) RegisterRoutes() {
	c.UserRoutes.Register()
} 