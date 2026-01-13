package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"go-api-template/internal/users/models"
	"go-api-template/internal/users/services"
	"go-api-template/pkg/response"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	service *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// List godoc
// @Summary      List all users
// @Description  Get a paginated list of users
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query     int  false  "Limit (default 20, max 100)"
// @Param        offset  query     int  false  "Offset (default 0)"
// @Success      200     {object}  models.UsersListResponse
// @Failure      401     {object}  response.Response
// @Failure      500     {object}  response.Response
// @Router       /users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))   //nolint:errcheck // default 0 is fine
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset")) //nolint:errcheck // default 0 is fine

	users, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		response.InternalError(w, "Failed to retrieve users")
		return
	}

	if users == nil {
		users = []models.User{}
	}

	response.Success(w, users)
}

// GetByID godoc
// @Summary      Get user by ID
// @Description  Retrieve a user by their unique identifier
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  models.UserResponse
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, map[string]string{"id": "Invalid UUID format"})
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if errors.Is(err, services.ErrUserNotFound) {
		response.NotFound(w, map[string]string{"id": "User not found"})
		return
	}
	if err != nil {
		response.InternalError(w, "Failed to retrieve user")
		return
	}

	response.Success(w, user)
}

// Create godoc
// @Summary      Create a new user
// @Description  Create a new user with email and name
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      models.CreateUserRequest  true  "User data"
// @Success      201      {object}  models.UserResponse
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      409      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, map[string]string{"body": "Invalid JSON"})
		return
	}

	// Basic validation
	if req.Email == "" {
		response.BadRequest(w, map[string]string{"email": "Email is required"})
		return
	}
	if req.Name == "" {
		response.BadRequest(w, map[string]string{"name": "Name is required"})
		return
	}

	user, err := h.service.Create(r.Context(), &req)
	if errors.Is(err, services.ErrEmailAlreadyExists) {
		response.Conflict(w, map[string]string{"email": "Email already exists"})
		return
	}
	if err != nil {
		response.InternalError(w, "Failed to create user")
		return
	}

	response.Created(w, user)
}

// Update godoc
// @Summary      Update a user
// @Description  Update user's email and/or name
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                    true  "User ID (UUID)"
// @Param        request  body      models.UpdateUserRequest  true  "User data to update"
// @Success      200      {object}  models.UserResponse
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      409      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /users/{id} [patch]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, map[string]string{"id": "Invalid UUID format"})
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, map[string]string{"body": "Invalid JSON"})
		return
	}

	user, err := h.service.Update(r.Context(), id, &req)
	if errors.Is(err, services.ErrUserNotFound) {
		response.NotFound(w, map[string]string{"id": "User not found"})
		return
	}
	if errors.Is(err, services.ErrEmailAlreadyExists) {
		response.Conflict(w, map[string]string{"email": "Email already exists"})
		return
	}
	if err != nil {
		response.InternalError(w, "Failed to update user")
		return
	}

	response.Success(w, user)
}

// Delete godoc
// @Summary      Delete a user
// @Description  Soft delete a user by ID
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      204  "No Content"
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, map[string]string{"id": "Invalid UUID format"})
		return
	}

	err = h.service.Delete(r.Context(), id)
	if errors.Is(err, services.ErrUserNotFound) {
		response.NotFound(w, map[string]string{"id": "User not found"})
		return
	}
	if err != nil {
		response.InternalError(w, "Failed to delete user")
		return
	}

	response.NoContent(w)
}
