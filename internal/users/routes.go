package users

import (
	"database/sql"
	"net/http"

	"go-api-template/internal/users/handlers"
	"go-api-template/internal/users/repositories"
	"go-api-template/internal/users/services"
)

// RegisterRoutes registers all user routes
func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repo := repositories.NewUserRepository(db)
	service := services.NewUserService(repo)
	handler := handlers.NewUserHandler(service)

	mux.HandleFunc("GET /users", handler.List)
	mux.HandleFunc("GET /users/{id}", handler.GetByID)
	mux.HandleFunc("POST /users", handler.Create)
	mux.HandleFunc("PATCH /users/{id}", handler.Update)
	mux.HandleFunc("DELETE /users/{id}", handler.Delete)
}
