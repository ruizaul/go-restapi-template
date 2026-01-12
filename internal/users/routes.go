package users

import (
	"net/http"

	"tacoshare-delivery-api/internal/users/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

// RegisterRoutes registers all user routes
func RegisterRoutes(mux *http.ServeMux, handler *handlers.UserHandler) {
	// Protected user routes
	mux.Handle("GET /api/v1/users/me", middleware.RequireAuth(http.HandlerFunc(handler.GetMe)))
	mux.Handle("GET /api/v1/users/{id}", middleware.RequireAuth(http.HandlerFunc(handler.GetUserByID)))
}
