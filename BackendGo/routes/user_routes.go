package routes

import (
	"CryptoLens_Backend/handlers"
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
	http.HandleFunc("/api/v1/user/register", r.handler.Register)
	http.HandleFunc("/api/v1/user/login", r.handler.Login)
	http.HandleFunc("/api/v1/user/logout", r.handler.Logout)
	http.HandleFunc("/api/v1/user/account", r.handler.GetAccount)
} 