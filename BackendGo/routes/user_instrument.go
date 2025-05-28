package routes

import (
	"CryptoLens_Backend/handlers"
	"CryptoLens_Backend/middleware"
	"net/http"
)

type UserInstrumentRoutes struct {
	handler *handlers.UserInstrumentHandler
}

func NewUserInstrumentRoutes(handler *handlers.UserInstrumentHandler) *UserInstrumentRoutes {
	return &UserInstrumentRoutes{
		handler: handler,
	}
}

func (r *UserInstrumentRoutes) Register() {
	http.HandleFunc("/api/v1/user/instruments/add", middleware.AuthMiddleware(r.handler.AddInstrument))
	http.HandleFunc("/api/v1/user/instruments/list", middleware.AuthMiddleware(r.handler.GetUserInstruments))
	http.HandleFunc("/api/v1/user/instruments/status", middleware.AuthMiddleware(r.handler.UpdateInstrumentStatus))
	http.HandleFunc("/api/v1/user/instruments/remove", middleware.AuthMiddleware(r.handler.RemoveInstrument))
}
