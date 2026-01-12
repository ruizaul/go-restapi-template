package handlers

import (
	"encoding/json"
	"net/http"

	"tacoshare-delivery-api/internal/orders/services"
	"tacoshare-delivery-api/pkg/httpx"
	"tacoshare-delivery-api/pkg/middleware"

	"github.com/google/uuid"
)

// AssignmentHandler handles assignment-related HTTP requests
type AssignmentHandler struct {
	assignmentService *services.AssignmentService
}

// NewAssignmentHandler creates a new assignment handler
func NewAssignmentHandler(assignmentService *services.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{
		assignmentService: assignmentService,
	}
}

// GetPendingAssignments godoc
//
//	@Summary		Get pending assignments for driver
//	@Description	Get all pending assignments for the authenticated driver
//	@Tags			assignments
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httpx.JSendSuccess	"Pending assignments retrieved"
//	@Failure		401	{object}	httpx.JSendError	"Unauthorized"
//	@Failure		500	{object}	httpx.JSendError	"Internal server error"
//	@Security		BearerAuth
//	@Router			/assignments/pending [get]
func (h *AssignmentHandler) GetPendingAssignments(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Get pending assignments
	assignments, err := h.assignmentService.GetPendingAssignmentsByDriver(userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, assignments)
}

// AcceptAssignment godoc
//
//	@Summary		Accept assignment
//	@Description	Accept a pending order assignment
//	@Tags			assignments
//	@Accept			json
//	@Produce		json
//	@Param			order_id	path		string				true	"Order ID (UUID)"
//	@Success		200			{object}	httpx.JSendSuccess	"Assignment accepted"
//	@Failure		400			{object}	httpx.JSendFail		"Invalid order ID"
//	@Failure		401			{object}	httpx.JSendError	"Unauthorized"
//	@Failure		404			{object}	httpx.JSendFail		"Assignment not found"
//	@Failure		500			{object}	httpx.JSendError	"Internal server error"
//	@Security		BearerAuth
//	@Router			/assignments/{order_id}/accept [post]
func (h *AssignmentHandler) AcceptAssignment(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	driverID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Parse order ID from path
	orderIDStr := r.PathValue("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"order_id": "ID de orden inválido",
		})
		return
	}

	// Accept assignment
	if err := h.assignmentService.AcceptOrder(orderID, driverID); err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]any{
		"message": "Orden aceptada exitosamente",
	})
}

// RejectAssignment godoc
//
//	@Summary		Reject assignment
//	@Description	Reject a pending order assignment
//	@Tags			assignments
//	@Accept			json
//	@Produce		json
//	@Param			order_id	path		string				true	"Order ID (UUID)"
//	@Param			request		body		RejectRequest		true	"Reject reason"
//	@Success		200			{object}	httpx.JSendSuccess	"Assignment rejected"
//	@Failure		400			{object}	httpx.JSendFail		"Invalid request"
//	@Failure		401			{object}	httpx.JSendError	"Unauthorized"
//	@Failure		500			{object}	httpx.JSendError	"Internal server error"
//	@Security		BearerAuth
//	@Router			/assignments/{order_id}/reject [post]
func (h *AssignmentHandler) RejectAssignment(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	driverID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Parse order ID from path
	orderIDStr := r.PathValue("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"order_id": "ID de orden inválido",
		})
		return
	}

	// Parse request body
	var req RejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"body": "Formato de solicitud inválido",
		})
		return
	}

	// Default reason if empty
	reason := req.Reason
	if reason == "" {
		reason = "No especificado"
	}

	// Reject assignment
	if err := h.assignmentService.RejectOrder(orderID, driverID, reason); err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]any{
		"message": "Orden rechazada",
	})
}

// RejectRequest represents a rejection request
type RejectRequest struct {
	Reason string `json:"reason" example:"Muy lejos"`
}
