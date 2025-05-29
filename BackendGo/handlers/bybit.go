package handlers

import (
	"CryptoLens_Backend/types"
	"encoding/json"
	"net/http"
	"strings"
)

type BybitHandler struct {
	bybitService types.BybitServiceInterface
}

func NewBybitHandler(bybitService types.BybitServiceInterface) *BybitHandler {
	return &BybitHandler{
		bybitService: bybitService,
	}
}

func (h *BybitHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	// Убираем префикс "Bearer " если он есть
	token = strings.TrimPrefix(token, "Bearer ")

	balance, err := h.bybitService.GetWalletBalance(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

func (h *BybitHandler) GetFeeRate(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	// Убираем префикс "Bearer " если он есть
	token = strings.TrimPrefix(token, "Bearer ")

	// Получаем параметры из query string
	category := r.URL.Query().Get("category")
	if category == "" {
		category = "spot" // значение по умолчанию
	}

	symbol := r.URL.Query().Get("symbol")
	baseCoin := r.URL.Query().Get("base_coin")

	feeRate, err := h.bybitService.GetFeeRate(r.Context(), token, category, symbol, baseCoin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feeRate)
}

func (h *BybitHandler) GetInstruments(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		category = "spot"
	}

	instruments, err := h.bybitService.GetInstruments(r.Context(), category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data":   instruments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
