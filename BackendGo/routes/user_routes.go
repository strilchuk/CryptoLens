package routes

import (
	"CryptoLens_Backend/handlers"
	"CryptoLens_Backend/middleware"
	"net/http"
)

type UserRoutes struct {
	handler *handlers.UserHandler
}

func NewUserRoutes(handler *handlers.UserHandler) *UserRoutes {
	return &UserRoutes{
		handler: handler,
	}
}

func (r *UserRoutes) Register() {
	// Публичные маршруты (без аутентификации)
	http.HandleFunc("/api/v1/user/register", r.handler.Register)
	http.HandleFunc("/api/v1/user/login", r.handler.Login)

	// Защищенные маршруты (требуют аутентификации)
	http.HandleFunc("/api/v1/user/logout", middleware.AuthMiddleware(r.handler.Logout))
	http.HandleFunc("/api/v1/user/account", middleware.AuthMiddleware(r.handler.GetAccount))
} 