package handlers

import (
	"SmallBot/metrics"
	"encoding/json"
	"net/http"
)

// MetricsHandler обрабатывает запросы метрик
type MetricsHandler struct {
	metrics *metrics.Metrics
}

// NewMetricsHandler создает новый обработчик метрик
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		metrics: metrics.GetInstance(),
	}
}

// ServeHTTP обрабатывает HTTP запросы
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/metrics":
		h.handleMetrics(w, r)
	case "/metrics/summary":
		h.handleSummary(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *MetricsHandler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	snapshot := h.metrics.GetSnapshot()
	if err := json.NewEncoder(w).Encode(snapshot); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *MetricsHandler) handleSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(h.metrics.GetSummary()))
}
