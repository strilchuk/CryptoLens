package container

import (
	"CryptoLens_Backend/env"
	"CryptoLens_Backend/handlers"
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/repositories"
	"CryptoLens_Backend/routes"
	"CryptoLens_Backend/services"
	"CryptoLens_Backend/types"
	"context"
	"database/sql"
	"strconv"
)

type Container struct {
	DB                   *sql.DB
	UserRepo             *repositories.UserRepository
	UserService          types.UserServiceInterface
	UserHandler          *handlers.UserHandler
	UserRoutes           *routes.UserRoutes
	BybitClient          bybit.Client
	BybitService         types.BybitServiceInterface
	BybitHandler         types.BybitHandlerInterface
	BybitRoutes          *routes.BybitRoutes
	UserInstrumentRepo   *repositories.UserInstrumentRepository
	BybitInstrumentRepo  *repositories.BybitInstrumentRepository
	UserInstrumentService types.UserInstrumentServiceInterface
	UserInstrumentHandler *handlers.UserInstrumentHandler
	UserInstrumentRoutes  *routes.UserInstrumentRoutes
}

func NewContainer(db *sql.DB, jwtKey []byte) *Container {
	// Инициализация репозиториев
	userRepo := repositories.NewUserRepository(db)
	userInstrumentRepo := repositories.NewUserInstrumentRepository(db)
	bybitInstrumentRepo := repositories.NewBybitInstrumentRepository(db)

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
	userInstrumentService := services.NewUserInstrumentService(userInstrumentRepo, bybitInstrumentRepo)
	bybitService := services.NewBybitService(bybitClient, db, userService)

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(userService)
	userInstrumentHandler := handlers.NewUserInstrumentHandler(userInstrumentService)
	bybitHandler := handlers.NewBybitHandler(bybitService)

	// Инициализация маршрутов
	userRoutes := routes.NewUserRoutes(userHandler)
	userInstrumentRoutes := routes.NewUserInstrumentRoutes(userInstrumentHandler)
	bybitRoutes := routes.NewBybitRoutes(bybitHandler)

	return &Container{
		DB:                   db,
		UserRepo:             userRepo,
		UserService:          userService,
		UserHandler:          userHandler,
		UserRoutes:           userRoutes,
		BybitClient:          bybitClient,
		BybitService:         bybitService,
		BybitHandler:         bybitHandler,
		BybitRoutes:          bybitRoutes,
		UserInstrumentRepo:   userInstrumentRepo,
		BybitInstrumentRepo:  bybitInstrumentRepo,
		UserInstrumentService: userInstrumentService,
		UserInstrumentHandler: userInstrumentHandler,
		UserInstrumentRoutes:  userInstrumentRoutes,
	}
}

func (c *Container) RegisterRoutes() {
	c.UserRoutes.Register()
	c.UserInstrumentRoutes.Register()
	c.BybitRoutes.Register()
}

func (c *Container) StartBackgroundTasks(ctx context.Context) {
	// Запускаем обновление инструментов
	go c.BybitService.StartInstrumentsUpdate(ctx)
	// Запускаем WebSocket
	go c.BybitService.StartWebSocket(ctx)
} 