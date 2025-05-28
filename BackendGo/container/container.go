package container

import (
	"CryptoLens_Backend/env"
	"CryptoLens_Backend/handlers"
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/repositories"
	"CryptoLens_Backend/routes"
	"CryptoLens_Backend/services"
	"context"
	"database/sql"
	"strconv"
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

	// Инициализация клиента Bybit
	recvWindow, _ := strconv.Atoi(env.GetBybitRecvWindow())
	apiMode := env.GetBybitApiMode()
	var apiUrl string
	if apiMode == "test" {
		apiUrl = env.GetBybitApiTestUrl()
	} else {
		apiUrl = env.GetBybitApiUrl()
	}
	bybitClient := bybit.NewClient(apiUrl, recvWindow, apiMode == "test")

	// Инициализация сервисов
	userService := services.NewUserService(userRepo, jwtKey, db)
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

func (c *Container) StartBackgroundTasks(ctx context.Context) {
	// Запускаем обновление инструментов
	go c.BybitService.StartInstrumentsUpdate(ctx)
} 