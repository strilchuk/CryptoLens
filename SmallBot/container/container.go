package container

import (
	"SmallBot/env"
	"SmallBot/handlers"
	"SmallBot/integration/bybit"
	"SmallBot/logger"
	"SmallBot/services"
	"SmallBot/types"
	"context"
	"fmt"
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

	symbol := env.GetSymbol()
	ctx := context.Background()
	if env.GetCancelOrdersOnStart() {
		// Отменяем все ордера при старте
		logger.LogWarn("⚠️  Отмена всех ордеров при старте включена!")

		// Отменяем ордера
		cancelled, err := bybitService.CancelAllOrders(ctx, symbol)
		if err != nil {
			logger.LogError("Ошибка при отмене ордеров %s: %v", symbol, err)
		} else {
			logger.LogInfo("Все ордера %s успешно отменены. Отменено: %v", symbol, cancelled)
		}
	} else {
		logger.LogInfo("✓ Отмена ордеров при старте отключена. Существующие ордера сохранены.")
		orders, err := bybitService.GetOpenOrders(ctx, symbol, nil, 50)

		if err != nil {
			logger.LogError("Не удалось получить список активных ордеров: %v", err)
		} else if len(orders.List) > 0 {
			logger.LogWarn("⚠️  Обнаружено %d активных ордеров для %s", len(orders.List), symbol)
			for _, order := range orders.List {
				logger.LogInfo("  - OrderID: %s, Side: %s, Price: %s, Qty: %s, Status: %s",
					order.OrderID, order.Side, order.Price, order.Qty, order.OrderStatus)
			}
		}
	}

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
	if env.GetCancelOrdersOnShutdown() {
		ctx := context.Background()

		logger.LogInfo("🛑 Отмена всех ордеров при завершении...")

		symbol := env.GetSymbol()
		cancelled, err := c.BybitService.CancelAllOrders(ctx, symbol)
		if err != nil {
			logger.LogError("Ошибка при отмене ордеров %s: %v", symbol, err)
			return fmt.Errorf("failed to cancel orders: %w", err)
		}

		logger.LogInfo("✓ Все ордера %s успешно отменены при завершении. Отменено: %v", symbol, cancelled)

	} else {
		logger.LogWarn("⚠️  Отмена ордеров при завершении отключена. Ордера остаются активными!")
	}

	return nil
}
