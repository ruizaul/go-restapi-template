package handlers

import (
	"encoding/json"
	"net/http"

	"tacoshare-delivery-api/internal/merchants/models"
	"tacoshare-delivery-api/internal/merchants/services"
	"tacoshare-delivery-api/pkg/httpx"
	"tacoshare-delivery-api/pkg/middleware"

	"github.com/google/uuid"
)

// MerchantHandler handles merchant-related HTTP requests
type MerchantHandler struct {
	service *services.MerchantService
}

// NewMerchantHandler creates a new merchant handler
func NewMerchantHandler(service *services.MerchantService) *MerchantHandler {
	return &MerchantHandler{service: service}
}

// CreateMerchant godoc
//
//	@Summary		Create merchant
//	@Description	Create a new merchant profile for the authenticated user
//	@Tags			merchants
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.CreateMerchantRequest	true	"Merchant details"
//	@Success		201		{object}	models.MerchantResponse			"Merchant created successfully"
//	@Failure		400		{object}	httpx.JSendFail					"Validation failed"
//	@Failure		401		{object}	httpx.JSendError				"Unauthorized"
//	@Failure		500		{object}	httpx.JSendError				"Internal server error"
//	@Security		BearerAuth
//	@Router			/merchants [post]
func (h *MerchantHandler) CreateMerchant(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Parse request body
	var req models.CreateMerchantRequest
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

	// Create merchant
	merchant, err := h.service.CreateMerchant(userID, &req)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusCreated, merchant)
}

// GetMyMerchant godoc
//
//	@Summary		Get my merchant
//	@Description	Get merchant profile for the authenticated user
//	@Tags			merchants
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.MerchantResponse	"Merchant retrieved successfully"
//	@Failure		401	{object}	httpx.JSendError		"Unauthorized"
//	@Failure		404	{object}	httpx.JSendFail			"Merchant not found"
//	@Failure		500	{object}	httpx.JSendError		"Internal server error"
//	@Security		BearerAuth
//	@Router			/merchants/me [get]
func (h *MerchantHandler) GetMyMerchant(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Get merchant
	merchant, err := h.service.GetMerchantByUserID(userID)
	if err != nil {
		httpx.RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, merchant)
}

// UpdateMyMerchant godoc
//
//	@Summary		Update my merchant
//	@Description	Update merchant profile for the authenticated user
//	@Tags			merchants
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.UpdateMerchantRequest	true	"Merchant update details"
//	@Success		200		{object}	models.MerchantResponse			"Merchant updated successfully"
//	@Failure		400		{object}	httpx.JSendFail					"Validation failed"
//	@Failure		401		{object}	httpx.JSendError				"Unauthorized"
//	@Failure		404		{object}	httpx.JSendFail					"Merchant not found"
//	@Failure		500		{object}	httpx.JSendError				"Internal server error"
//	@Security		BearerAuth
//	@Router			/merchants/me [patch]
func (h *MerchantHandler) UpdateMyMerchant(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Get merchant ID first
	merchant, err := h.service.GetMerchantByUserID(userID)
	if err != nil {
		httpx.RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Parse request body
	var req models.UpdateMerchantRequest
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

	// Update merchant
	updatedMerchant, err := h.service.UpdateMerchant(merchant.ID, &req)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, updatedMerchant)
}

// GetMerchantByID godoc
//
//	@Summary		Get merchant by ID
//	@Description	Get merchant details by ID (admin only)
//	@Tags			merchants
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Merchant ID (UUID)"
//	@Success		200	{object}	models.MerchantResponse	"Merchant retrieved successfully"
//	@Failure		400	{object}	httpx.JSendFail			"Invalid merchant ID"
//	@Failure		401	{object}	httpx.JSendError		"Unauthorized"
//	@Failure		404	{object}	httpx.JSendFail			"Merchant not found"
//	@Failure		500	{object}	httpx.JSendError		"Internal server error"
//	@Security		BearerAuth
//	@Router			/merchants/{id} [get]
func (h *MerchantHandler) GetMerchantByID(w http.ResponseWriter, r *http.Request) {
	// Parse merchant ID from path
	merchantIDStr := r.PathValue("id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "ID de negocio inválido",
		})
		return
	}

	// Get merchant
	merchant, err := h.service.GetMerchantByID(merchantID)
	if err != nil {
		httpx.RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, merchant)
}

// ListMerchants godoc
//
//	@Summary		List merchants
//	@Description	List all merchants with optional filters
//	@Tags			merchants
//	@Accept			json
//	@Produce		json
//	@Param			city			query		string						false	"Filter by city"
//	@Param			business_type	query		string						false	"Filter by business type"
//	@Param			status			query		string						false	"Filter by status" Enums(active, inactive, suspended)
//	@Success		200				{object}	models.MerchantListResponse	"Merchants retrieved successfully"
//	@Failure		500				{object}	httpx.JSendError			"Internal server error"
//	@Router			/merchants [get]
func (h *MerchantHandler) ListMerchants(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	city := r.URL.Query().Get("city")
	businessType := r.URL.Query().Get("business_type")
	status := r.URL.Query().Get("status")

	// Get merchants
	merchants, err := h.service.GetAllMerchants(city, businessType, status)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, merchants)
}
