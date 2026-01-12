package drivers

import (
	"net/http"

	"tacoshare-delivery-api/internal/drivers/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

// RegisterRoutes registers all driver routes
func RegisterRoutes(mux *http.ServeMux, locationHandler *handlers.LocationHandler) {
	// Protected routes (drivers only)
	mux.Handle("PATCH /api/v1/drivers/me/location", middleware.RequireAuth(
		middleware.RequireRole("driver")(http.HandlerFunc(locationHandler.UpdateMyLocation)),
	))
	mux.Handle("GET /api/v1/drivers/me/location", middleware.RequireAuth(
		middleware.RequireRole("driver")(http.HandlerFunc(locationHandler.GetMyLocation)),
	))
	mux.Handle("PATCH /api/v1/drivers/me/availability", middleware.RequireAuth(
		middleware.RequireRole("driver")(http.HandlerFunc(locationHandler.UpdateMyAvailability)),
	))
}
