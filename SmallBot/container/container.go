package container

import (
	"SmallBot/env"
	"SmallBot/handlers"
	"SmallBot/integration/bybit"
	"SmallBot/logger"
	"SmallBot/services"
	"SmallBot/types"
	"context"
	"strconv"
)

type Container struct {
	BybitClient  bybit.Client
	BybitService types.BybitServiceInterface
	wsHandler    *handlers.BybitWebSocketHandler
}

func NewContainer() *Container {
	recvWindow, _ := strconv.Atoi(env.GetBybitRecvWindow())
	apiMode := env.GetBybitApiMode()
	apiUrl := env.GetBybitApiUrl()
	bybitClient := bybit.NewClient(apiUrl, recvWindow, apiMode == "test")

	// Сначала создаём bybitService с временным nil wsHandler
	var bybitService *services.BybitService
	bybitService = services.NewBybitService(bybitClient, nil)

	// Теперь создаём wsHandler, передавая bybitService
	wsHandler := handlers.NewBybitWebSocketHandler(bybitService)
	bybitService.SetWebSocketHandler(wsHandler)

	// Отменяем все ордера при старте
	ctx := context.Background()
	logger.LogInfo("Отмена всех ордеров при старте...")

	// Отменяем ордера для BTCUSDT
	_, err := bybitService.CancelAllOrders(ctx, "BTCUSDT")
	if err != nil {
		logger.LogError("Ошибка при отмене ордеров BTCUSDT: %v", err)
	} else {
		logger.LogInfo("Все ордера BTCUSDT успешно отменены")
	}

	// Отменяем ордера для ETHUSDT
	//_, err = bybitService.CancelAllOrders(ctx, "ETHUSDT")
	//if err != nil {
	//	logger.LogError("Ошибка при отмене ордеров ETHUSDT: %v", err)
	//} else {
	//	logger.LogInfo("Все ордера ETHUSDT успешно отменены")
	//}

	return &Container{
		BybitClient:  bybitClient,
		BybitService: bybitService,
		wsHandler:    wsHandler,
	}
}

func (c *Container) StartBackgroundTasks(ctx context.Context) {
	go c.BybitService.StartWebSocket(ctx)
	go c.BybitService.StartPrivateWebSocket(ctx)
}

func (c *Container) Close() error {
	ctx := context.Background()
	logger.LogInfo("Отмена всех ордеров при завершении...")

	// Отменяем ордера для BTCUSDT
	_, err := c.BybitService.CancelAllOrders(ctx, "BTCUSDT")
	if err != nil {
		logger.LogError("Ошибка при отмене ордеров BTCUSDT: %v", err)
	} else {
		logger.LogInfo("Все ордера BTCUSDT успешно отменены")
	}

	//// Отменяем ордера для ETHUSDT
	//_, err = c.BybitService.CancelAllOrders(ctx, "ETHUSDT")
	//if err != nil {
	//	logger.LogError("Ошибка при отмене ордеров ETHUSDT: %v", err)
	//} else {
	//	logger.LogInfo("Все ордера ETHUSDT успешно отменены")
	//}

	return nil
}
