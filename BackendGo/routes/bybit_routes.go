package routes

import (
	"CryptoLens_Backend/handlers"
	"CryptoLens_Backend/middleware"
	"net/http"
)

type BybitRoutes struct {
	bybitHandler *handlers.BybitHandler
}

func NewBybitRoutes(bybitHandler *handlers.BybitHandler) *BybitRoutes {
	return &BybitRoutes{
		bybitHandler: bybitHandler,
	}
}

func (r *BybitRoutes) Register() {
	// Все маршруты Bybit требуют аутентификации
	http.HandleFunc("/api/v1/bybit/wallet/balance", middleware.AuthMiddleware(r.bybitHandler.GetWalletBalance))
	http.HandleFunc("/api/v1/bybit/wallet/fee-rate", middleware.AuthMiddleware(r.bybitHandler.GetFeeRate))
	http.HandleFunc("/api/v1/bybit/instruments", middleware.AuthMiddleware(r.bybitHandler.GetInstruments))
}
