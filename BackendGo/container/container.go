package container

import (
	"CryptoLens_Backend/handlers"
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/repositories"
	"CryptoLens_Backend/routes"
	"CryptoLens_Backend/services"
	"database/sql"
)

type Container struct {
	DB           *sql.DB
	UserRepo     *repositories.UserRepository
	UserService  *services.UserService
	UserHandler  *handlers.UserHandler
	UserRoutes   *routes.UserRoutes
	BybitClient  bybit.Client
	BybitService *services.BybitService
	BybitHandler *handlers.BybitHandler
	BybitRoutes  *routes.BybitRoutes
}

func NewContainer(db *sql.DB, jwtKey []byte) *Container {
	// Инициализация репозиториев
	userRepo := repositories.NewUserRepository(db)

	// Инициализация сервисов
	userService := services.NewUserService(userRepo, jwtKey, db)

	// Инициализация клиента Bybit
	bybitClient := bybit.NewClient("https://api.bybit.com", 5000)

	// Инициализация сервиса Bybit
	bybitService := services.NewBybitService(bybitClient, db, userService)

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(userService)
	bybitHandler := handlers.NewBybitHandler(bybitService)

	// Инициализация маршрутов
	userRoutes := routes.NewUserRoutes(userHandler)
	bybitRoutes := routes.NewBybitRoutes(bybitHandler)

	return &Container{
		DB:           db,
		UserRepo:     userRepo,
		UserService:  userService,
		UserHandler:  userHandler,
		UserRoutes:   userRoutes,
		BybitClient:  bybitClient,
		BybitService: bybitService,
		BybitHandler: bybitHandler,
		BybitRoutes:  bybitRoutes,
	}
}

func (c *Container) RegisterRoutes() {
	c.UserRoutes.Register()
	c.BybitRoutes.Register()
} 