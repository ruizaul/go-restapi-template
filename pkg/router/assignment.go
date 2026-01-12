package router

import (
	"net/http"

	"tacoshare-delivery-api/internal/orders/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

// AssignmentRouter handles assignment-related routes
type AssignmentRouter struct {
	handler *handlers.AssignmentHandler
}

// NewAssignmentRouter creates a new assignment router
func NewAssignmentRouter(handler *handlers.AssignmentHandler) *AssignmentRouter {
	return &AssignmentRouter{handler: handler}
}

// RegisterRoutes registers all assignment routes
func (ar *AssignmentRouter) RegisterRoutes(mux *http.ServeMux) {
	// Get pending assignments (driver only)
	mux.Handle("GET /api/v1/assignments/pending", middleware.RequireAuth(
		http.HandlerFunc(ar.handler.GetPendingAssignments),
	))

	// Accept assignment (driver only)
	mux.Handle("POST /api/v1/assignments/{order_id}/accept", middleware.RequireAuth(
		http.HandlerFunc(ar.handler.AcceptAssignment),
	))

	// Reject assignment (driver only)
	mux.Handle("POST /api/v1/assignments/{order_id}/reject", middleware.RequireAuth(
		http.HandlerFunc(ar.handler.RejectAssignment),
	))
}
