package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"tacoshare-delivery-api/internal/notifications/models"
	"tacoshare-delivery-api/internal/notifications/services"
	"tacoshare-delivery-api/pkg/httpx"

	"github.com/google/uuid"
)

const (
	errNotificationNotFound     = "notification not found"
	errUnauthorizedNotification = "unauthorized access to notification"
)

// NotificationHandler handles notification HTTP requests
type NotificationHandler struct {
	service *services.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// RegisterToken godoc
//
//	@Summary		Register FCM token
//	@Description	Register a Firebase Cloud Messaging token for push notifications. Tokens are device-specific and allow sending notifications to that device. If the token already exists, it will be updated as active.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.RegisterTokenRequest	true	"FCM token details (token, device_type, optional device_id)"
//	@Success		200		{object}	models.TokenResponse		"Token registered successfully"
//	@Failure		400		{object}	httpx.JSendFail				"Invalid request body or validation failed"
//	@Failure		401		{object}	httpx.JSendError			"Unauthorized - invalid or missing token"
//	@Failure		500		{object}	httpx.JSendError			"Internal server error - failed to register token"
//	@Security		BearerAuth
//	@Router			/notifications/register-token [post]
func (h *NotificationHandler) RegisterToken(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userIDVal := r.Context().Value("user_id")
	userIDStr, ok := userIDVal.(string)
	if !ok {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido en el contexto",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido",
		})
		return
	}

	var req models.RegisterTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de la solicitud inválido",
		})
		return
	}

	// Validate required fields
	if req.Token == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"token": "El token FCM es requerido",
		})
		return
	}

	if req.DeviceType == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"device_type": "El tipo de dispositivo es requerido",
		})
		return
	}

	token, err := h.service.RegisterToken(r.Context(), userID, req.Token, req.DeviceType, req.DeviceID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error al registrar el token")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, token)
}

// UnregisterToken godoc
//
//	@Summary		Unregister FCM token
//	@Description	Deactivate a Firebase Cloud Messaging token. Use this when the user logs out or uninstalls the app to stop receiving notifications on that device.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.UnregisterTokenRequest	true	"FCM token to deactivate"
//	@Success		200		{object}	httpx.JSendSuccess				"Token unregistered successfully"
//	@Failure		400		{object}	httpx.JSendFail					"Invalid request body or missing token"
//	@Failure		401		{object}	httpx.JSendError				"Unauthorized - invalid or missing token"
//	@Failure		404		{object}	httpx.JSendFail					"Token not found"
//	@Failure		500		{object}	httpx.JSendError				"Internal server error - failed to unregister token"
//	@Security		BearerAuth
//	@Router			/notifications/unregister-token [delete]
func (h *NotificationHandler) UnregisterToken(w http.ResponseWriter, r *http.Request) {
	var req models.UnregisterTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de la solicitud inválido",
		})
		return
	}

	if req.Token == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"token": "El token FCM es requerido",
		})
		return
	}

	err := h.service.UnregisterToken(r.Context(), req.Token)
	if err != nil {
		if err.Error() == "fcm token not found" {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"token": "Token FCM no encontrado",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error al desregistrar el token")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]any{
		"message": "Token desregistrado exitosamente",
	})
}

// ListNotifications godoc
//
//	@Summary		List notifications
//	@Description	Get paginated list of notifications for the authenticated user. Supports pagination with page and limit query parameters. Default page size is 20, maximum is 100.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int								false	"Page number (default: 1)"					minimum(1)	default(1)
//	@Param			limit	query		int								false	"Items per page (default: 20, max: 100)"	minimum(1)	maximum(100)	default(20)
//	@Success		200		{object}	models.NotificationListResponse	"Successfully retrieved notifications"
//	@Failure		400		{object}	httpx.JSendFail					"Invalid query parameters"
//	@Failure		401		{object}	httpx.JSendError				"Unauthorized - invalid or missing token"
//	@Failure		500		{object}	httpx.JSendError				"Internal server error - failed to retrieve notifications"
//	@Security		BearerAuth
//	@Router			/notifications [get]
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userIDStr, ok := userIDVal.(string)
	if !ok {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido en el contexto",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido",
		})
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	notifications, pagination, err := h.service.ListNotifications(r.Context(), userID, page, limit)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error al obtener notificaciones")
		return
	}

	response := map[string]any{
		"items":      notifications,
		"pagination": pagination,
	}

	httpx.RespondSuccess(w, http.StatusOK, response)
}

// GetNotification godoc
//
//	@Summary		Get notification details
//	@Description	Get detailed information about a specific notification. User must own the notification to access it.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int							true	"Notification ID"
//	@Success		200	{object}	models.NotificationResponse	"Successfully retrieved notification"
//	@Failure		400	{object}	httpx.JSendFail				"Invalid notification ID"
//	@Failure		401	{object}	httpx.JSendError			"Unauthorized - invalid or missing token"
//	@Failure		403	{object}	httpx.JSendFail				"Forbidden - notification belongs to another user"
//	@Failure		404	{object}	httpx.JSendFail				"Notification not found"
//	@Failure		500	{object}	httpx.JSendError			"Internal server error - failed to retrieve notification"
//	@Security		BearerAuth
//	@Router			/notifications/{id} [get]
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userIDStr, ok := userIDVal.(string)
	if !ok {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido en el contexto",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido",
		})
		return
	}

	// Parse notification ID from path
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "ID de notificación inválido",
		})
		return
	}

	notification, err := h.service.GetNotification(r.Context(), id, userID)
	if err != nil {
		if err.Error() == errNotificationNotFound {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"id": "Notificación no encontrada",
			})
			return
		}
		if err.Error() == errUnauthorizedNotification {
			httpx.RespondFail(w, http.StatusForbidden, map[string]any{
				"error": "No tiene acceso a esta notificación",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error al obtener notificación")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, notification)
}

// MarkAsRead godoc
//
//	@Summary		Mark notification as read
//	@Description	Mark a specific notification as read. This updates the is_read field and sets the read_at timestamp. User must own the notification.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int					true	"Notification ID"
//	@Success		200	{object}	httpx.JSendSuccess	"Notification marked as read successfully"
//	@Failure		400	{object}	httpx.JSendFail		"Invalid notification ID"
//	@Failure		401	{object}	httpx.JSendError	"Unauthorized - invalid or missing token"
//	@Failure		403	{object}	httpx.JSendFail		"Forbidden - notification belongs to another user"
//	@Failure		404	{object}	httpx.JSendFail		"Notification not found"
//	@Failure		500	{object}	httpx.JSendError	"Internal server error - failed to update notification"
//	@Security		BearerAuth
//	@Router			/notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userIDStr, ok := userIDVal.(string)
	if !ok {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido en el contexto",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido",
		})
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "ID de notificación inválido",
		})
		return
	}

	err = h.service.MarkAsRead(r.Context(), id, userID)
	if err != nil {
		if err.Error() == errNotificationNotFound {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"id": "Notificación no encontrada",
			})
			return
		}
		if err.Error() == errUnauthorizedNotification {
			httpx.RespondFail(w, http.StatusForbidden, map[string]any{
				"error": "No tiene acceso a esta notificación",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error al marcar notificación como leída")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]any{
		"message": "Notificación marcada como leída",
	})
}

// MarkAllAsRead godoc
//
//	@Summary		Mark all notifications as read
//	@Description	Mark all unread notifications for the authenticated user as read. This updates all notifications with is_read=false to is_read=true and sets their read_at timestamps.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httpx.JSendSuccess	"All notifications marked as read successfully"
//	@Failure		401	{object}	httpx.JSendError	"Unauthorized - invalid or missing token"
//	@Failure		500	{object}	httpx.JSendError	"Internal server error - failed to update notifications"
//	@Security		BearerAuth
//	@Router			/notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userIDStr, ok := userIDVal.(string)
	if !ok {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido en el contexto",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido",
		})
		return
	}

	err = h.service.MarkAllAsRead(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error al marcar notificaciones como leídas")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]any{
		"message": "Todas las notificaciones marcadas como leídas",
	})
}

// DeleteNotification godoc
//
//	@Summary		Delete notification
//	@Description	Delete a specific notification. User must own the notification to delete it. This permanently removes the notification from the database.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int					true	"Notification ID"
//	@Success		200	{object}	httpx.JSendSuccess	"Notification deleted successfully"
//	@Failure		400	{object}	httpx.JSendFail		"Invalid notification ID"
//	@Failure		401	{object}	httpx.JSendError	"Unauthorized - invalid or missing token"
//	@Failure		403	{object}	httpx.JSendFail		"Forbidden - notification belongs to another user"
//	@Failure		404	{object}	httpx.JSendFail		"Notification not found"
//	@Failure		500	{object}	httpx.JSendError	"Internal server error - failed to delete notification"
//	@Security		BearerAuth
//	@Router			/notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userIDStr, ok := userIDVal.(string)
	if !ok {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido en el contexto",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido",
		})
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "ID de notificación inválido",
		})
		return
	}

	err = h.service.DeleteNotification(r.Context(), id, userID)
	if err != nil {
		if err.Error() == errNotificationNotFound {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"id": "Notificación no encontrada",
			})
			return
		}
		if err.Error() == errUnauthorizedNotification {
			httpx.RespondFail(w, http.StatusForbidden, map[string]any{
				"error": "No tiene acceso a esta notificación",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error al eliminar notificación")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]any{
		"message": "Notificación eliminada exitosamente",
	})
}

// GetUnreadCount godoc
//
//	@Summary		Get unread notification count
//	@Description	Get the count of unread notifications for the authenticated user. Useful for displaying notification badges.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.UnreadCountResponse	"Successfully retrieved unread count"
//	@Failure		401	{object}	httpx.JSendError			"Unauthorized - invalid or missing token"
//	@Failure		500	{object}	httpx.JSendError			"Internal server error - failed to count notifications"
//	@Security		BearerAuth
//	@Router			/notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userIDStr, ok := userIDVal.(string)
	if !ok {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido en el contexto",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "ID de usuario inválido",
		})
		return
	}

	count, err := h.service.GetUnreadCount(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error al obtener contador de notificaciones")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]any{
		"count": count,
	})
}
