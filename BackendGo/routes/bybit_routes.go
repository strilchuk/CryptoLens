package routes

import (
	"CryptoLens_Backend/handlers"
	"net/http"
)

type BybitRoutes struct {
	handler *handlers.BybitHandler
}

func NewBybitRoutes(handler *handlers.BybitHandler) *BybitRoutes {
	return &BybitRoutes{
		handler: handler,
	}
}

func (r *BybitRoutes) Register() {
	http.HandleFunc("/api/v1/bybit/wallet-balance", r.handler.GetWalletBalance)
	http.HandleFunc("/api/v1/bybit/fee-rate", r.handler.GetFeeRate)
} 