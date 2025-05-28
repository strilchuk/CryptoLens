package routes

import (
	"CryptoLens_Backend/handlers"
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
	http.HandleFunc("/api/bybit/wallet/balance", r.bybitHandler.GetWalletBalance)
	http.HandleFunc("/api/bybit/wallet/fee-rate", r.bybitHandler.GetFeeRate)
	http.HandleFunc("/api/bybit/instruments", r.bybitHandler.GetInstruments)
} 