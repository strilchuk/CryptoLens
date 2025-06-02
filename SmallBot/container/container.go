package container

import (
	"SmallBot/env"
	"SmallBot/handlers"
	"SmallBot/integration/bybit"
	"SmallBot/services"
	"SmallBot/types"
	"context"
	"strconv"
)

type Container struct {
	BybitClient  bybit.Client
	BybitService types.BybitServiceInterface
}

func NewContainer() *Container {
	recvWindow, _ := strconv.Atoi(env.GetBybitRecvWindow())
	apiMode := env.GetBybitApiMode()
	apiUrl := env.GetBybitApiUrl()
	bybitClient := bybit.NewClient(apiUrl, recvWindow, apiMode == "test")
	wsHandler := handlers.NewBybitWebSocketHandler()
	bybitService := services.NewBybitService(bybitClient, wsHandler)
	return &Container{
		BybitClient:  bybitClient,
		BybitService: bybitService,
	}
}

func (c *Container) StartBackgroundTasks(ctx context.Context) {
	go c.BybitService.StartWebSocket(ctx)
	go c.BybitService.StartPrivateWebSocket(ctx)
}

func (c *Container) Close() error {
	// Закрываем все необходимые соединения
	return nil
}
