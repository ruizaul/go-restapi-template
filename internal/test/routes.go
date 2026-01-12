package test

import (
	"net/http"

	"go-api-template/internal/test/handlers"
	"go-api-template/internal/test/services"
)

// RegisterRoutes registers all routes for the test module
func RegisterRoutes(mux *http.ServeMux) {
	service := services.NewHelloService()
	handler := handlers.NewHelloHandler(service)

	mux.HandleFunc("GET /test/hello", handler.GetHelloWorld)
}
