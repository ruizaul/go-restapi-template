package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"go-api-template/internal/auth/models"
	"go-api-template/internal/auth/services"
	"go-api-template/pkg/response"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	service *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.RegisterRequest  true  "Registration data"
// @Success      201      {object}  models.AuthResponse
// @Failure      400      {object}  response.Response
// @Failure      409      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, map[string]string{"body": "Invalid JSON"})
		return
	}

	user, tokens, err := h.service.Register(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailAlreadyExists):
			response.Conflict(w, map[string]string{"email": "Email already exists"})
		case errors.Is(err, services.ErrInvalidEmail):
			response.BadRequest(w, map[string]string{"email": "Invalid email format"})
		case errors.Is(err, services.ErrWeakPassword):
			response.BadRequest(w, map[string]string{"password": "Password must be at least 8 characters"})
		case errors.Is(err, services.ErrNameRequired):
			response.BadRequest(w, map[string]string{"name": "Name is required"})
		default:
			response.InternalError(w, "Failed to create user")
		}
		return
	}

	response.Created(w, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.LoginRequest  true  "Login credentials"
// @Success      200      {object}  models.AuthResponse
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, map[string]string{"body": "Invalid JSON"})
		return
	}

	user, tokens, err := h.service.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			response.Unauthorized(w, map[string]string{"credentials": "Invalid email or password"})
			return
		}
		response.InternalError(w, "Failed to authenticate user")
		return
	}

	response.Success(w, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

// Refresh godoc
// @Summary      Refresh tokens
// @Description  Get new access and refresh tokens using a valid refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.RefreshRequest  true  "Refresh token"
// @Success      200      {object}  models.AuthResponse
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, map[string]string{"body": "Invalid JSON"})
		return
	}

	if req.RefreshToken == "" {
		response.BadRequest(w, map[string]string{"refresh_token": "Refresh token is required"})
		return
	}

	user, tokens, err := h.service.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidToken):
			response.Unauthorized(w, map[string]string{"refresh_token": "Invalid refresh token"})
		case errors.Is(err, services.ErrExpiredToken):
			response.Unauthorized(w, map[string]string{"refresh_token": "Refresh token has expired"})
		case errors.Is(err, services.ErrInvalidTokenType):
			response.Unauthorized(w, map[string]string{"refresh_token": "Invalid token type"})
		case errors.Is(err, services.ErrUserNotFound):
			response.Unauthorized(w, map[string]string{"refresh_token": "User not found"})
		default:
			response.InternalError(w, "Failed to refresh tokens")
		}
		return
	}

	response.Success(w, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Get the profile of the currently authenticated user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.ProfileResponse
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /auth/me [get]
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok {
		response.Unauthorized(w, map[string]string{"auth": "User not authenticated"})
		return
	}

	user, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			response.NotFound(w, map[string]string{"user": "User not found"})
			return
		}
		response.InternalError(w, "Failed to retrieve profile")
		return
	}

	response.Success(w, user)
}

// Logout godoc
// @Summary      Logout user
// @Description  Logout the current user (client should discard tokens)
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.MessageResponse
// @Failure      401  {object}  response.Response
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, _ *http.Request) {
	// In a stateless JWT implementation, logout is handled client-side
	// The client should discard the tokens
	// For added security, you could implement a token blacklist
	response.Success(w, map[string]string{"message": "Successfully logged out"})
}

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for user email
	UserEmailKey ContextKey = "user_email"
)
