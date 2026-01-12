package websockets

import (
	"net/http"

	"tacoshare-delivery-api/internal/websockets/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

// RegisterRoutes registers all WebSocket routes
func RegisterRoutes(mux *http.ServeMux, handler *handlers.WSHandler) {
	// General WebSocket connection (authenticated users)
	// Use WebSocketAuth instead of RequireAuth to avoid interfering with upgrade
	mux.Handle("GET /ws", middleware.WebSocketAuth(
		http.HandlerFunc(handler.HandleConnection),
	))

	// Order-specific WebSocket channel
	mux.Handle("GET /ws/orders/{order_id}", middleware.WebSocketAuth(
		http.HandlerFunc(handler.HandleOrderChannel),
	))

	// Driver-specific WebSocket channel
	mux.Handle("GET /ws/drivers/{driver_id}", middleware.WebSocketAuth(
		http.HandlerFunc(handler.HandleDriverChannel),
	))
}
