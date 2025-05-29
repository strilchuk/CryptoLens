package handlers

import (
	"CryptoLens_Backend/models"
	"CryptoLens_Backend/types"
	"encoding/json"
	"net/http"
)

type UserInstrumentHandler struct {
	userInstrumentService types.UserInstrumentServiceInterface
}

func NewUserInstrumentHandler(userInstrumentService types.UserInstrumentServiceInterface) *UserInstrumentHandler {
	return &UserInstrumentHandler{
		userInstrumentService: userInstrumentService,
	}
}

// AddInstrument добавляет инструмент для пользователя
func (h *UserInstrumentHandler) AddInstrument(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserInstrumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем userID из контекста (предполагается, что middleware уже добавил его)
	userID := r.Context().Value("userID").(string)

	instrument, err := h.userInstrumentService.AddInstrument(r.Context(), userID, req.Symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.UserInstrumentResponse{
		ID:        instrument.ID,
		UserID:    instrument.UserID,
		Symbol:    instrument.Symbol,
		IsActive:  instrument.IsActive,
		CreatedAt: instrument.CreatedAt,
		UpdatedAt: instrument.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserInstruments получает все инструменты пользователя
func (h *UserInstrumentHandler) GetUserInstruments(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	instruments, err := h.userInstrumentService.GetUserInstruments(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]models.UserInstrumentResponse, len(instruments))
	for i, instrument := range instruments {
		response[i] = models.UserInstrumentResponse{
			ID:        instrument.ID,
			UserID:    instrument.UserID,
			Symbol:    instrument.Symbol,
			IsActive:  instrument.IsActive,
			CreatedAt: instrument.CreatedAt,
			UpdatedAt: instrument.UpdatedAt,
			BybitInstrument: instrument.BybitInstrument,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateInstrumentStatus обновляет статус инструмента пользователя
func (h *UserInstrumentHandler) UpdateInstrumentStatus(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateUserInstrumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.userInstrumentService.UpdateInstrumentStatus(r.Context(), req.ID, req.IsActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RemoveInstrument удаляет инструмент у пользователя
func (h *UserInstrumentHandler) RemoveInstrument(w http.ResponseWriter, r *http.Request) {
	instrumentID := r.URL.Query().Get("id")
	if instrumentID == "" {
		http.Error(w, "Instrument ID is required", http.StatusBadRequest)
		return
	}

	err := h.userInstrumentService.RemoveInstrument(r.Context(), instrumentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
} 