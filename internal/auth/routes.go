package auth

import (
	"database/sql"
	"net/http"
	"time"

	"go-api-template/internal/auth/handlers"
	"go-api-template/internal/auth/services"
	"go-api-template/pkg/config"
	"go-api-template/pkg/middleware"
)

// RegisterRoutes registers all auth routes
func RegisterRoutes(mux *http.ServeMux, db *sql.DB, cfg *config.Config) *services.JWTService {
	// Initialize JWT service with config
	jwtService := services.NewJWTService(
		cfg.JWT.SecretKey,
		time.Duration(cfg.JWT.AccessTokenTTL)*time.Minute,
		time.Duration(cfg.JWT.RefreshTokenTTL)*time.Hour,
	)

	// Initialize auth service
	authService := services.NewAuthService(db, jwtService)

	// Initialize handler
	handler := handlers.NewAuthHandler(authService)

	// Public routes (no auth required)
	mux.HandleFunc("POST /auth/register", handler.Register)
	mux.HandleFunc("POST /auth/login", handler.Login)
	mux.HandleFunc("POST /auth/refresh", handler.Refresh)

	// Protected routes (auth required)
	mux.HandleFunc("GET /auth/me", middleware.RequireAuth(jwtService, handler.GetProfile))
	mux.HandleFunc("POST /auth/logout", middleware.RequireAuth(jwtService, handler.Logout))

	return jwtService
}
