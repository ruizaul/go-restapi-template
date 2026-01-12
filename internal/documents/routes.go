package documents

import (
	"net/http"

	"tacoshare-delivery-api/internal/documents/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

// RegisterRoutes registers all document routes
func RegisterRoutes(mux *http.ServeMux, handler *handlers.DocumentHandler, uploadHandler *handlers.UploadHandler) {
	// User document routes (protected)
	mux.Handle("GET /api/v1/documents/me", middleware.RequireAuth(
		http.HandlerFunc(handler.GetMyDocuments),
	))
	mux.Handle("PATCH /api/v1/documents/me", middleware.RequireAuth(
		http.HandlerFunc(handler.UpdateDocument),
	))
	mux.Handle("DELETE /api/v1/documents/me", middleware.RequireAuth(
		http.HandlerFunc(handler.DeleteDocument),
	))

	// Upload routes (protected)
	mux.Handle("POST /api/v1/documents/upload", middleware.RequireAuth(
		http.HandlerFunc(uploadHandler.UploadDocument),
	))

	// Admin routes (admin only)
	mux.Handle("GET /api/v1/documents", middleware.RequireAuth(
		middleware.RequireRole("admin")(http.HandlerFunc(handler.GetAllDocuments)),
	))
	mux.Handle("GET /api/v1/documents/{user_id}", middleware.RequireAuth(
		middleware.RequireRole("admin")(http.HandlerFunc(handler.GetDocumentByUserID)),
	))
	mux.Handle("PATCH /api/v1/documents/{document_id}", middleware.RequireAuth(
		middleware.RequireRole("admin")(http.HandlerFunc(handler.UpdateDocumentByID)),
	))
	mux.Handle("PATCH /api/v1/documents/{user_id}/review", middleware.RequireAuth(
		middleware.RequireRole("admin")(http.HandlerFunc(handler.MarkAsReviewed)),
	))
}
