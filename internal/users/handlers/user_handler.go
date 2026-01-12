package handlers

import (
	"net/http"

	"tacoshare-delivery-api/internal/users/services"
	"tacoshare-delivery-api/pkg/httpx"
	"tacoshare-delivery-api/pkg/middleware"
	"tacoshare-delivery-api/pkg/validator"

	"github.com/google/uuid"
)

// UserHandler handles user HTTP requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetMe godoc
//
//	@Summary		Get current user profile
//	@Description	Get the authenticated user's profile information including ID, name, email, phone, role, and timestamps. Requires valid JWT token in Authorization header.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.UserProfileResponse	"User profile retrieved successfully with all user details"
//	@Failure		401	{object}	httpx.JSendError			"Unauthorized - missing or invalid token"
//	@Failure		404	{object}	httpx.JSendFail				"User not found - user ID from token doesn't exist in database"
//	@Failure		500	{object}	httpx.JSendError			"Internal server error - failed to retrieve user profile"
//	@Security		BearerAuth
//	@Router			/users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.userService.GetUserProfile(userID)
	if err != nil {
		if err.Error() == "user not found" {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"error": "User not found",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Failed to retrieve user profile")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, user)
}

// GetUserByID godoc
//
//	@Summary		Get user by ID
//	@Description	Get a user's profile information by their UUID. Users can only access their own profile unless they have admin role. Returns full user details including ID, name, email, phone, role, and timestamps.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"User ID in UUID format (e.g., 550e8400-e29b-41d4-a716-446655440000)"
//	@Success		200	{object}	models.UserProfileResponse	"User profile retrieved successfully with all user details"
//	@Failure		400	{object}	httpx.JSendFail				"Invalid user ID format - must be valid UUID"
//	@Failure		401	{object}	httpx.JSendError			"Unauthorized - missing or invalid token"
//	@Failure		403	{object}	httpx.JSendError			"Forbidden - insufficient permissions to access this user's profile"
//	@Failure		404	{object}	httpx.JSendFail				"User not found - no user exists with the provided ID"
//	@Failure		500	{object}	httpx.JSendError			"Internal server error - failed to retrieve user profile"
//	@Security		BearerAuth
//	@Router			/users/{id} [get]
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	// Get user ID from path
	idParam := r.PathValue("id")
	if !validator.IsValidUUID(idParam) {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "Invalid user ID format",
		})
		return
	}

	requestedUserID, err := uuid.Parse(idParam)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "Error parsing user ID",
		})
		return
	}

	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	currentUserRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "User role not found in context")
		return
	}

	// Check if user is requesting their own profile or is an admin
	if requestedUserID != currentUserID && currentUserRole != "admin" {
		httpx.RespondError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	user, err := h.userService.GetUserProfile(requestedUserID)
	if err != nil {
		if err.Error() == "user not found" {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"error": "User not found",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Failed to retrieve user profile")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, user)
}
