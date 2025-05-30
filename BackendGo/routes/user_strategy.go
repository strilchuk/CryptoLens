package routes

import (
	"CryptoLens_Backend/handlers"
	"CryptoLens_Backend/middleware"
	"net/http"
)

type UserStrategyRoutes struct {
	handler *handlers.UserStrategyHandler
}

func NewUserStrategyRoutes(handler *handlers.UserStrategyHandler) *UserStrategyRoutes {
	return &UserStrategyRoutes{
		handler: handler,
	}
}

func (r *UserStrategyRoutes) Register() {
	http.HandleFunc("/api/v1/user/strategies", middleware.AuthMiddleware(r.handler.GetUserStrategies))
	http.HandleFunc("/api/v1/user/strategies/add", middleware.AuthMiddleware(r.handler.AddStrategy))
	http.HandleFunc("/api/v1/user/strategies/update", middleware.AuthMiddleware(r.handler.UpdateStrategyStatus))
	http.HandleFunc("/api/v1/user/strategies/remove", middleware.AuthMiddleware(r.handler.RemoveStrategy))
} 