package services

import (
	"SmallBot/env"
	"SmallBot/integration/bybit"
	"SmallBot/logger"
	"SmallBot/types"
	"context"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BybitService struct {
	bybitClient bybit.Client
	wsClient    *bybit.WebSocketClient
	wsHandler   types.BybitWebSocketHandlerInterface
	wsMutex     sync.Mutex
}

func NewBybitService(
	bybitClient bybit.Client,
	wsHandler types.BybitWebSocketHandlerInterface,
) *BybitService {
	recvWindow, _ := strconv.Atoi(env.GetBybitRecvWindow())
	wsURL := env.GetBybitWsUrl() + "/v5/public/spot"
	wsClient := bybit.NewWebSocketClient(wsURL, recvWindow, "", "")

	return &BybitService{
		bybitClient: bybitClient,
		wsClient:    wsClient,
		wsHandler:   wsHandler,
	}
}

func (s *BybitService) GetWalletBalance(ctx context.Context, token string) (*bybit.BybitWalletBalance, error) {
	balance, err := s.bybitClient.GetWalletBalance(ctx)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (s *BybitService) GetFeeRate(
	ctx context.Context,
	token string,
	category string,
	symbol string,
	baseCoin string,
) (*bybit.BybitFeeRateResponse, error) {

	// Получаем ставки комиссии через клиент Bybit
	var symbolPtr, baseCoinPtr *string
	if symbol != "" {
		symbolPtr = &symbol
	}
	if baseCoin != "" {
		baseCoinPtr = &baseCoin
	}

	feeRate, err := s.bybitClient.GetFeeRate(ctx, category, symbolPtr, baseCoinPtr)
	if err != nil {
		return nil, err
	}

	return feeRate, nil
}

// преобразует строку в decimal.Decimal
func parseDecimal(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}

// возвращает количество знаков после запятой
func getPrecision(s string) int {
	if s == "" {
		return 0
	}
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return 0
	}
	return len(parts[1])
}

// запускает WebSocket-соединение и подписку на каналы
func (s *BybitService) StartWebSocket(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				publicChannels := []string{"tickers.BTCUSDT"}
				logger.LogInfo("Подписываемся на каналы для активных инструментов: %v", publicChannels)

				// Подключаемся к WebSocket
				if err := s.wsClient.Connect(ctx); err != nil {
					logger.LogError("Failed to connect to WebSocket: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Запускаем обработку сообщений
				s.wsClient.StartMessageHandler(context.Background(), s.wsHandler.HandleMessage)

				// Подписываемся на публичные каналы
				if err := s.wsClient.Subscribe(ctx, publicChannels); err != nil {
					logger.LogError("Failed to subscribe to public channels: %v", err)
					s.wsClient.Close()
					time.Sleep(5 * time.Second)
					continue
				}

				// Логируем успешную подписку
				logger.LogInfo("Успешно подписались на %d каналов для активных инструментов", len(publicChannels))

				// Ждем завершения контекста
				<-ctx.Done()
				return
			}
		}
	}()
}

// StartPrivateWebSocket запускает приватные WebSocket-соединения для активных аккаунтов
func (s *BybitService) StartPrivateWebSocket(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.closePrivateWebSockets()
				return
			default:

				s.wsMutex.Lock()

				// Создаем соединения
				privateWsURL := env.GetBybitWsUrl() + "/v5/private"
				recvWindow, _ := strconv.Atoi(env.GetBybitRecvWindow())

				wsClient := bybit.NewWebSocketClient(privateWsURL, recvWindow, env.GetBybitApiToken(), env.GetBybitApiSecret())

				// Подключаемся и подписываемся
				if err := wsClient.Connect(ctx); err != nil {
					logger.LogError("Failed to connect to private WebSocket: %v", err)
					continue
				}

				privateChannels := []string{
					"order.spot",
					"execution.spot",
					//"execution.fast.spot",
					"wallet",
				}

				wsClient.StartMessageHandler(ctx, func(ctx context.Context, msg bybit.WebSocketMessage) {
					s.wsHandler.HandlePrivateMessage(ctx, msg)
				})

				if err := wsClient.Subscribe(ctx, privateChannels); err != nil {
					logger.LogError("Failed to subscribe to private channels: %v", err)
					wsClient.Close()
					continue
				}

				logger.LogInfo("Успешно подключились к приватному WebSocket")

				s.wsMutex.Unlock()

			}
		}
	}()
}

// closePrivateWebSockets закрывает все приватные WebSocket-соединения
func (s *BybitService) closePrivateWebSockets() {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()
	//
	//for userID, client := range s.privateWsClients {
	//	client.Close()
	//	delete(s.privateWsClients, userID)
	//	logger.LogInfo("Закрыто приватное WebSocket-соединение для userID: %s", userID)
	//}
}
