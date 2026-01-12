package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"tacoshare-delivery-api/internal/drivers/models"
	"tacoshare-delivery-api/internal/drivers/services"
	"tacoshare-delivery-api/pkg/httpx"
	"tacoshare-delivery-api/pkg/middleware"

	"github.com/google/uuid"
)

// LocationHandler handles driver location-related HTTP requests
type LocationHandler struct {
	service *services.LocationService
}

// NewLocationHandler creates a new location handler
func NewLocationHandler(service *services.LocationService) *LocationHandler {
	return &LocationHandler{service: service}
}

// UpdateMyLocation godoc
//
//	@Summary		Update my location
//	@Description	Update the current driver's real-time location
//	@Tags			drivers
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.UpdateLocationRequest	true	"Location details"
//	@Success		200		{object}	models.DriverLocationResponse	"Location updated successfully"
//	@Failure		400		{object}	httpx.JSendFail					"Validation failed"
//	@Failure		401		{object}	httpx.JSendError				"Unauthorized"
//	@Failure		500		{object}	httpx.JSendError				"Internal server error"
//	@Security		BearerAuth
//	@Router			/drivers/me/location [patch]
func (h *LocationHandler) UpdateMyLocation(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Verify user is a driver
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != middleware.RoleDriver {
		httpx.RespondError(w, http.StatusForbidden, "Solo los conductores pueden actualizar su ubicación")
		return
	}

	// Parse request body
	var req models.UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"body": "Formato de solicitud inválido",
		})
		return
	}

	// Validate request
	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, err)
		return
	}

	// Update location
	location, err := h.service.UpdateLocation(userID, &req)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, location)
}

// GetMyLocation godoc
//
//	@Summary		Get my location
//	@Description	Get the current driver's location
//	@Tags			drivers
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.DriverLocationResponse	"Location retrieved successfully"
//	@Failure		401	{object}	httpx.JSendError				"Unauthorized"
//	@Failure		404	{object}	httpx.JSendFail					"Location not found"
//	@Failure		500	{object}	httpx.JSendError				"Internal server error"
//	@Security		BearerAuth
//	@Router			/drivers/me/location [get]
func (h *LocationHandler) GetMyLocation(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Verify user is a driver
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != middleware.RoleDriver {
		httpx.RespondError(w, http.StatusForbidden, "Solo los conductores pueden ver su ubicación")
		return
	}

	// Get location
	location, err := h.service.GetMyLocation(userID)
	if err != nil {
		httpx.RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, location)
}

// UpdateMyAvailability godoc
//
//	@Summary		Update my availability
//	@Description	Update the current driver's availability status
//	@Tags			drivers
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.UpdateAvailabilityRequest	true	"Availability status"
//	@Success		200		{object}	httpx.JSendSuccess					"Availability updated successfully"
//	@Failure		400		{object}	httpx.JSendFail						"Validation failed"
//	@Failure		401		{object}	httpx.JSendError					"Unauthorized"
//	@Failure		500		{object}	httpx.JSendError					"Internal server error"
//	@Security		BearerAuth
//	@Router			/drivers/me/availability [patch]
func (h *LocationHandler) UpdateMyAvailability(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Verify user is a driver
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != middleware.RoleDriver {
		httpx.RespondError(w, http.StatusForbidden, "Solo los conductores pueden actualizar su disponibilidad")
		return
	}

	// Parse request body
	var req models.UpdateAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"body": "Formato de solicitud inválido",
		})
		return
	}

	// Update availability
	if err := h.service.UpdateAvailability(userID, req.IsAvailable); err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	status := "disponible"
	if !req.IsAvailable {
		status = "no disponible"
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]any{
		"message":      fmt.Sprintf("Disponibilidad actualizada a %s", status),
		"is_available": req.IsAvailable,
	})
}
