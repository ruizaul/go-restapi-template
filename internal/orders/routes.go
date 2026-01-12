package orders

import (
	"net/http"

	"tacoshare-delivery-api/internal/orders/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

// RegisterRoutes registers all order routes
func RegisterRoutes(mux *http.ServeMux, handler *handlers.OrderHandler) {
	// Public routes (webhook from external backend)
	mux.HandleFunc("POST /api/v1/orders/external", handler.CreateExternalOrder)

	// Protected routes (authenticated users)
	mux.Handle("GET /api/v1/orders", middleware.RequireAuth(
		http.HandlerFunc(handler.ListOrders),
	))
	mux.Handle("GET /api/v1/orders/{id}", middleware.RequireAuth(
		http.HandlerFunc(handler.GetOrder),
	))

	// Driver routes
	mux.Handle("POST /api/v1/orders/{id}/accept", middleware.RequireAuth(
		middleware.RequireRole("driver")(http.HandlerFunc(handler.AcceptOrder)),
	))
	mux.Handle("POST /api/v1/orders/{id}/reject", middleware.RequireAuth(
		middleware.RequireRole("driver")(http.HandlerFunc(handler.RejectOrder)),
	))
	mux.Handle("PATCH /api/v1/orders/{id}", middleware.RequireAuth(
		middleware.RequireRole("driver")(http.HandlerFunc(handler.UpdateOrderStatus)),
	))
	mux.Handle("POST /api/v1/orders/{id}/verify-delivery-code", middleware.RequireAuth(
		middleware.RequireRole("driver")(http.HandlerFunc(handler.VerifyDeliveryCode)),
	))
	mux.Handle("POST /api/v1/orders/{id}/complete-delivery", middleware.RequireAuth(
		middleware.RequireRole("driver")(http.HandlerFunc(handler.CompleteDelivery)),
	))
	mux.Handle("GET /api/v1/drivers/me/active-order", middleware.RequireAuth(
		middleware.RequireRole("driver")(http.HandlerFunc(handler.GetMyActiveOrder)),
	))
	mux.Handle("GET /api/v1/drivers/me/assignments", middleware.RequireAuth(
		middleware.RequireRole("driver")(http.HandlerFunc(handler.GetMyPendingAssignments)),
	))
}
