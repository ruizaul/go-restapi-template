package auth

import (
	"net/http"

	"tacoshare-delivery-api/internal/auth/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

// RegisterRoutes registers all auth routes
func RegisterRoutes(mux *http.ServeMux, handler *handlers.AuthHandler) {
	// Public auth routes
	mux.HandleFunc("POST /api/v1/auth/register", handler.Register)
	mux.HandleFunc("POST /api/v1/auth/verify-otp", handler.VerifyOTP)
	mux.HandleFunc("POST /api/v1/auth/login", handler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", handler.RefreshToken)

	// Protected auth routes (require authentication)
	mux.Handle("POST /api/v1/auth/logout", middleware.RequireAuth(http.HandlerFunc(handler.Logout)))
	mux.Handle("POST /api/v1/auth/logout-all", middleware.RequireAuth(http.HandlerFunc(handler.LogoutAllDevices)))
	mux.Handle("GET /api/v1/auth/sessions", middleware.RequireAuth(http.HandlerFunc(handler.GetActiveSessions)))
}
