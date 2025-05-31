package container

import (
	"CryptoLens_Backend/env"
	"CryptoLens_Backend/handlers"
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/repositories"
	"CryptoLens_Backend/routes"
	"CryptoLens_Backend/services"
	"CryptoLens_Backend/storages"
	"CryptoLens_Backend/trading"
	"CryptoLens_Backend/types"
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

type Container struct {
	DB                    *sql.DB
	UserRepo              *repositories.UserRepository
	UserService           types.UserServiceInterface
	UserHandler           *handlers.UserHandler
	UserRoutes            *routes.UserRoutes
	BybitClient           bybit.Client
	BybitService          types.BybitServiceInterface
	BybitHandler          types.BybitHandlerInterface
	BybitRoutes           *routes.BybitRoutes
	UserInstrumentRepo    *repositories.UserInstrumentRepository
	BybitInstrumentRepo   *repositories.BybitInstrumentRepository
	BybitAccountRepo      types.BybitAccountRepositoryInterface
	UserInstrumentService types.UserInstrumentServiceInterface
	UserInstrumentHandler *handlers.UserInstrumentHandler
	UserInstrumentRoutes  *routes.UserInstrumentRoutes
	UserStrategyRepo      *repositories.UserStrategyRepository
	UserStrategyService   types.UserStrategyServiceInterface
	UserStrategyHandler   *handlers.UserStrategyHandler
	UserStrategyRoutes    *routes.UserStrategyRoutes
	TradeLogRepo          types.TradeLogRepositoryInterface
	WebSocketHandler      types.BybitWebSocketHandlerInterface
}

func NewContainer(db *sql.DB, jwtKey []byte) *Container {
	// Инициализация репозиториев
	userRepo := repositories.NewUserRepository(db)
	userInstrumentRepo := repositories.NewUserInstrumentRepository(db)
	bybitInstrumentRepo := repositories.NewBybitInstrumentRepository(db)
	userStrategyRepo := repositories.NewUserStrategyRepository(db)
	bybitAccountRepo := repositories.NewBybitAccountRepository(db)
	tradeLogRepo := repositories.NewTradeLogRepository(db)

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

	// Создаем менеджер стратегий
	strategyManager := trading.NewStrategyManager(bybitClient, userInstrumentRepo, bybitAccountRepo)

	// Создаем обработчик WebSocket
	wsHandler := handlers.NewBybitWebSocketHandler(strategyManager, tradeLogRepo)

	// Создаем сервисы, зависящие от менеджера стратегий
	userInstrumentService := services.NewUserInstrumentService(userInstrumentRepo, bybitInstrumentRepo, strategyManager)
	userStrategyService := services.NewUserStrategyService(
		userStrategyRepo,
		strategyManager,
		repositories.NewBybitInstrumentRepository(db),
	)

	// Создаем сервис Bybit
	bybitService := services.NewBybitService(bybitClient, db, userService, wsHandler)

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(userService)
	userInstrumentHandler := handlers.NewUserInstrumentHandler(userInstrumentService)
	bybitHandler := handlers.NewBybitHandler(bybitService)
	userStrategyHandler := handlers.NewUserStrategyHandler(userStrategyService)

	// Инициализация маршрутов
	userRoutes := routes.NewUserRoutes(userHandler)
	userInstrumentRoutes := routes.NewUserInstrumentRoutes(userInstrumentHandler)
	bybitRoutes := routes.NewBybitRoutes(bybitHandler)
	userStrategyRoutes := routes.NewUserStrategyRoutes(userStrategyHandler)

	return &Container{
		DB:                    db,
		UserRepo:              userRepo,
		UserService:           userService,
		UserHandler:           userHandler,
		UserRoutes:            userRoutes,
		BybitClient:           bybitClient,
		BybitService:          bybitService,
		BybitHandler:          bybitHandler,
		BybitRoutes:           bybitRoutes,
		UserInstrumentRepo:    userInstrumentRepo,
		BybitInstrumentRepo:   bybitInstrumentRepo,
		BybitAccountRepo:      bybitAccountRepo,
		UserInstrumentService: userInstrumentService,
		UserInstrumentHandler: userInstrumentHandler,
		UserInstrumentRoutes:  userInstrumentRoutes,
		UserStrategyRepo:      userStrategyRepo,
		UserStrategyService:   userStrategyService,
		UserStrategyHandler:   userStrategyHandler,
		UserStrategyRoutes:    userStrategyRoutes,
		TradeLogRepo:          tradeLogRepo,
		WebSocketHandler:      wsHandler,
	}
}

func (c *Container) RegisterRoutes() {
	c.UserRoutes.Register()
	c.UserInstrumentRoutes.Register()
	c.UserStrategyRoutes.Register()
	c.BybitRoutes.Register()
}

func (c *Container) StartBackgroundTasks(ctx context.Context) {
	// Загружаем активные стратегии
	if err := c.UserStrategyService.LoadActiveStrategies(ctx); err != nil {
		logger.LogError("Ошибка при загрузке активных стратегий: %v", err)
	}

	// Запускаем обновление инструментов
	go c.BybitService.StartInstrumentsUpdate(ctx)
	// Запускаем WebSocket
	go c.BybitService.StartWebSocket(ctx)
	// Запускаем Приватный WebSocket
	go c.BybitService.StartPrivateWebSocket(ctx)
}

func (c *Container) Close() error {
	// Закрываем соединение с Redis
	if err := storages.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}
	return nil
}
