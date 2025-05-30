package handlers

import (
	"CryptoLens_Backend/models"
	"CryptoLens_Backend/services"
	"encoding/json"
	"net/http"
)

type UserStrategyHandler struct {
	userStrategyService *services.UserStrategyService
}

func NewUserStrategyHandler(userStrategyService *services.UserStrategyService) *UserStrategyHandler {
	return &UserStrategyHandler{
		userStrategyService: userStrategyService,
	}
}

func (h *UserStrategyHandler) AddStrategy(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем userID из контекста (предполагается, что middleware уже добавил его)
	userID := r.Context().Value("userID").(string)

	strategy, err := h.userStrategyService.AddStrategy(r.Context(), userID, req.StrategyName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.UserStrategyResponse{
		ID:           strategy.ID,
		UserID:       strategy.UserID,
		StrategyName: strategy.StrategyName,
		IsActive:     strategy.IsActive,
		CreatedAt:    strategy.CreatedAt,
		UpdatedAt:    strategy.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserStrategyHandler) GetUserStrategies(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	strategies, err := h.userStrategyService.GetUserStrategies(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]models.UserStrategyResponse, len(strategies))
	for i, strategy := range strategies {
		response[i] = models.UserStrategyResponse{
			ID:           strategy.ID,
			UserID:       strategy.UserID,
			StrategyName: strategy.StrategyName,
			IsActive:     strategy.IsActive,
			CreatedAt:    strategy.CreatedAt,
			UpdatedAt:    strategy.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserStrategyHandler) UpdateStrategyStatus(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateUserStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.userStrategyService.UpdateStrategyStatus(r.Context(), req.ID, req.IsActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserStrategyHandler) RemoveStrategy(w http.ResponseWriter, r *http.Request) {
	strategyID := r.URL.Query().Get("id")
	if strategyID == "" {
		http.Error(w, "Strategy ID is required", http.StatusBadRequest)
		return
	}

	err := h.userStrategyService.RemoveStrategy(r.Context(), strategyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
} 