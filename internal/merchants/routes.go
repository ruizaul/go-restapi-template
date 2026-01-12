package merchants

import (
	"net/http"

	"tacoshare-delivery-api/internal/merchants/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

// RegisterRoutes registers all merchant routes
func RegisterRoutes(mux *http.ServeMux, handler *handlers.MerchantHandler) {
	// Public routes
	mux.HandleFunc("GET /api/v1/merchants", handler.ListMerchants)

	// Protected routes (authenticated users)
	mux.Handle("POST /api/v1/merchants", middleware.RequireAuth(
		http.HandlerFunc(handler.CreateMerchant),
	))
	mux.Handle("GET /api/v1/merchants/me", middleware.RequireAuth(
		http.HandlerFunc(handler.GetMyMerchant),
	))
	mux.Handle("PATCH /api/v1/merchants/me", middleware.RequireAuth(
		http.HandlerFunc(handler.UpdateMyMerchant),
	))

	// Admin routes
	mux.Handle("GET /api/v1/merchants/{id}", middleware.RequireAuth(
		middleware.RequireRole("admin")(http.HandlerFunc(handler.GetMerchantByID)),
	))
}
