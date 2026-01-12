package handlers

import (
	"encoding/json"
	"net/http"

	"tacoshare-delivery-api/internal/notifications/models"
	"tacoshare-delivery-api/internal/notifications/services"
	"tacoshare-delivery-api/pkg/httpx"

	"github.com/google/uuid"
)

// AdminNotificationHandler handles admin notification operations
type AdminNotificationHandler struct {
	service *services.NotificationService
}

// NewAdminNotificationHandler creates a new admin notification handler
func NewAdminNotificationHandler(service *services.NotificationService) *AdminNotificationHandler {
	return &AdminNotificationHandler{service: service}
}

// SendNotification godoc
//
//	@Summary		Send notification (Admin)
//	@Description	Send a push notification to a specific user. This endpoint creates the notification in the database and sends it via FCM to all active devices. Useful for testing and manual notifications.
//	@Tags			notifications-admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.CreateNotificationRequest	true	"Notification details"
//	@Success		201		{object}	models.NotificationResponse			"Notification sent successfully"
//	@Failure		400		{object}	httpx.JSendFail						"Validation failed"
//	@Failure		401		{object}	httpx.JSendError					"Unauthorized"
//	@Failure		403		{object}	httpx.JSendError					"Forbidden - admin only"
//	@Failure		500		{object}	httpx.JSendError					"Internal server error"
//	@Security		BearerAuth
//	@Router			/notifications/send [post]
func (h *AdminNotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	var req models.CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de la solicitud inválido",
		})
		return
	}

	// Validate required fields
	if req.UserID == uuid.Nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"user_id": "El ID de usuario es requerido",
		})
		return
	}

	if req.Title == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"title": "El título es requerido",
		})
		return
	}

	if req.Body == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"body": "El cuerpo del mensaje es requerido",
		})
		return
	}

	if req.NotificationType == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"notification_type": "El tipo de notificación es requerido",
		})
		return
	}

	// Send notification
	notification, err := h.service.CreateAndSend(r.Context(), &req)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error al enviar notificación")
		return
	}

	httpx.RespondSuccess(w, http.StatusCreated, notification)
}
