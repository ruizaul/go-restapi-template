package users

import (
	"database/sql"
	"net/http"

	"go-api-template/internal/auth/services"
	"go-api-template/internal/users/handlers"
	"go-api-template/internal/users/repositories"
	userservices "go-api-template/internal/users/services"
	"go-api-template/pkg/middleware"
)

// RegisterRoutes registers all user routes (protected with auth)
func RegisterRoutes(mux *http.ServeMux, db *sql.DB, jwtService *services.JWTService) {
	repo := repositories.NewUserRepository(db)
	service := userservices.NewUserService(repo)
	handler := handlers.NewUserHandler(service)

	// All user routes require authentication
	mux.HandleFunc("GET /users", middleware.RequireAuth(jwtService, handler.List))
	mux.HandleFunc("GET /users/{id}", middleware.RequireAuth(jwtService, handler.GetByID))
	mux.HandleFunc("POST /users", middleware.RequireAuth(jwtService, handler.Create))
	mux.HandleFunc("PATCH /users/{id}", middleware.RequireAuth(jwtService, handler.Update))
	mux.HandleFunc("DELETE /users/{id}", middleware.RequireAuth(jwtService, handler.Delete))
}
