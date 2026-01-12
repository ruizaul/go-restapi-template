package notifications

import (
	"net/http"

	"tacoshare-delivery-api/internal/notifications/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

// RegisterRoutes registers all notification routes
func RegisterRoutes(mux *http.ServeMux, handler *handlers.NotificationHandler, adminHandler *handlers.AdminNotificationHandler) {
	// User notification routes (protected)
	mux.Handle("POST /api/v1/notifications/register-token", middleware.RequireAuth(
		http.HandlerFunc(handler.RegisterToken),
	))
	mux.Handle("DELETE /api/v1/notifications/unregister-token", middleware.RequireAuth(
		http.HandlerFunc(handler.UnregisterToken),
	))
	mux.Handle("GET /api/v1/notifications", middleware.RequireAuth(
		http.HandlerFunc(handler.ListNotifications),
	))
	mux.Handle("GET /api/v1/notifications/unread-count", middleware.RequireAuth(
		http.HandlerFunc(handler.GetUnreadCount),
	))
	mux.Handle("GET /api/v1/notifications/{id}", middleware.RequireAuth(
		http.HandlerFunc(handler.GetNotification),
	))
	mux.Handle("PUT /api/v1/notifications/{id}/read", middleware.RequireAuth(
		http.HandlerFunc(handler.MarkAsRead),
	))
	mux.Handle("PUT /api/v1/notifications/read-all", middleware.RequireAuth(
		http.HandlerFunc(handler.MarkAllAsRead),
	))
	mux.Handle("DELETE /api/v1/notifications/{id}", middleware.RequireAuth(
		http.HandlerFunc(handler.DeleteNotification),
	))

	// Admin notification routes (admin only)
	mux.Handle("POST /api/v1/notifications/send", middleware.RequireAuth(
		middleware.RequireRole("admin")(http.HandlerFunc(adminHandler.SendNotification)),
	))
}
