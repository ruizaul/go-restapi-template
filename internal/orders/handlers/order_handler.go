package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"tacoshare-delivery-api/internal/orders/models"
	"tacoshare-delivery-api/internal/orders/services"
	"tacoshare-delivery-api/pkg/httpx"
	"tacoshare-delivery-api/pkg/middleware"

	"github.com/google/uuid"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	orderService      *services.OrderService
	assignmentService *services.AssignmentService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderService *services.OrderService, assignmentService *services.AssignmentService) *OrderHandler {
	return &OrderHandler{
		orderService:      orderService,
		assignmentService: assignmentService,
	}
}

// CreateExternalOrder godoc
//
//	@Summary		Create external order (webhook)
//	@Description	Receive and create a new order from an external backend (webhook endpoint)
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.CreateExternalOrderRequest	true	"External order details"
//	@Success		201		{object}	models.OrderResponse				"Order created and assignment started"
//	@Failure		400		{object}	httpx.JSendFail						"Validation failed"
//	@Failure		500		{object}	httpx.JSendError					"Internal server error"
//	@Router			/orders/external [post]
func (h *OrderHandler) CreateExternalOrder(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.CreateExternalOrderRequest
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

	// Create order
	order, err := h.orderService.CreateExternalOrder(&req)
	if err != nil {
		// Check if it's a distance validation error or service unavailable
		errMsg := err.Error()
		if strings.Contains(errMsg, "excede el límite máximo") ||
			strings.Contains(errMsg, "no se pudo obtener la distancia") ||
			strings.Contains(errMsg, "servicio de validación de distancia no disponible") {
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"distance": errMsg,
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Start assignment process asynchronously
	go func() {
		_ = h.assignmentService.AssignOrderToDriver(order.ID)
	}()

	httpx.RespondSuccess(w, http.StatusCreated, order)
}

// GetOrder godoc
//
//	@Summary		Get order by ID
//	@Description	Get detailed information about a specific order
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Order ID (UUID)"
//	@Success		200	{object}	models.OrderResponse	"Order retrieved successfully"
//	@Failure		400	{object}	httpx.JSendFail			"Invalid order ID"
//	@Failure		401	{object}	httpx.JSendError		"Unauthorized"
//	@Failure		404	{object}	httpx.JSendFail			"Order not found"
//	@Failure		500	{object}	httpx.JSendError		"Internal server error"
//	@Security		BearerAuth
//	@Router			/orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Parse order ID from path
	orderIDStr := r.PathValue("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "ID de orden inválido",
		})
		return
	}

	// Get order
	order, err := h.orderService.GetOrderByID(orderID)
	if err != nil {
		httpx.RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, order)
}

// AcceptOrder godoc
//
//	@Summary		Accept order
//	@Description	Driver accepts an assigned order
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Order ID (UUID)"
//	@Success		200	{object}	models.OrderResponse	"Order accepted successfully"
//	@Failure		400	{object}	httpx.JSendFail			"Invalid order ID"
//	@Failure		401	{object}	httpx.JSendError		"Unauthorized"
//	@Failure		404	{object}	httpx.JSendFail			"Order not found"
//	@Failure		500	{object}	httpx.JSendError		"Internal server error"
//	@Security		BearerAuth
//	@Router			/orders/{id}/accept [post]
func (h *OrderHandler) AcceptOrder(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Verify user is a driver
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != middleware.RoleDriver {
		httpx.RespondError(w, http.StatusForbidden, "Solo los conductores pueden aceptar órdenes")
		return
	}

	// Parse order ID from path
	orderIDStr := r.PathValue("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "ID de orden inválido",
		})
		return
	}

	// Accept order
	if err := h.assignmentService.AcceptOrder(orderID, userID); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated order with full details
	order, err := h.orderService.GetOrderByID(orderID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, order)
}

// RejectOrder godoc
//
//	@Summary		Reject order
//	@Description	Driver rejects an assigned order
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Order ID (UUID)"
//	@Param			request	body		models.CancelOrderRequest	false	"Rejection reason"
//	@Success		200		{object}	models.OrderResponse		"Order rejected successfully"
//	@Failure		400		{object}	httpx.JSendFail				"Invalid order ID"
//	@Failure		401		{object}	httpx.JSendError			"Unauthorized"
//	@Failure		404		{object}	httpx.JSendFail				"Order not found"
//	@Failure		500		{object}	httpx.JSendError			"Internal server error"
//	@Security		BearerAuth
//	@Router			/orders/{id}/reject [post]
func (h *OrderHandler) RejectOrder(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Verify user is a driver
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != middleware.RoleDriver {
		httpx.RespondError(w, http.StatusForbidden, "Solo los conductores pueden rechazar órdenes")
		return
	}

	// Parse order ID from path
	orderIDStr := r.PathValue("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "ID de orden inválido",
		})
		return
	}

	// Parse rejection reason (optional)
	var req models.CancelOrderRequest
	reason := "No especificado"
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Reason != "" {
		reason = req.Reason
	}

	// Reject order
	if err := h.assignmentService.RejectOrder(orderID, userID, reason); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated order with full details
	order, err := h.orderService.GetOrderByID(orderID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, order)
}

// VerifyDeliveryCode godoc
//
//	@Summary		Verify delivery code
//	@Description	Verify the delivery code before completing an order
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Order ID (UUID)"
//	@Param			request	body		models.VerifyDeliveryCodeRequest	true	"Delivery code"
//	@Success		200		{object}	httpx.JSendSuccess{data=map[string]bool}	"Code verification result"
//	@Failure		400		{object}	httpx.JSendFail						"Validation failed"
//	@Failure		401		{object}	httpx.JSendError					"Unauthorized"
//	@Failure		403		{object}	httpx.JSendError					"Forbidden"
//	@Failure		404		{object}	httpx.JSendFail						"Order not found"
//	@Failure		500		{object}	httpx.JSendError					"Internal server error"
//	@Security		BearerAuth
//	@Router			/orders/{id}/verify-delivery-code [post]
func (h *OrderHandler) VerifyDeliveryCode(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Verify user is a driver
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != middleware.RoleDriver {
		httpx.RespondError(w, http.StatusForbidden, "Solo los conductores pueden verificar códigos de entrega")
		return
	}

	// Parse order ID from path
	orderIDStr := r.PathValue("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "ID de orden inválido",
		})
		return
	}

	// Verify order belongs to this driver
	if err := h.orderService.VerifyOrderBelongsToDriver(orderID, userID); err != nil {
		httpx.RespondError(w, http.StatusForbidden, err.Error())
		return
	}

	// Parse request body
	var req models.VerifyDeliveryCodeRequest
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

	// Verify delivery code
	isValid, err := h.orderService.VerifyDeliveryCode(orderID, req.DeliveryCode)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]bool{
		"valid": isValid,
	})
}

// CompleteDelivery godoc
//
//	@Summary		Complete delivery with code verification
//	@Description	Verify delivery code and mark order as delivered in one step
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Order ID (UUID)"
//	@Param			request	body		models.VerifyDeliveryCodeRequest	true	"Delivery code"
//	@Success		200		{object}	models.OrderResponse				"Order completed successfully"
//	@Failure		400		{object}	httpx.JSendFail						"Validation failed or incorrect code"
//	@Failure		401		{object}	httpx.JSendError					"Unauthorized"
//	@Failure		403		{object}	httpx.JSendError					"Forbidden"
//	@Failure		404		{object}	httpx.JSendFail						"Order not found"
//	@Failure		500		{object}	httpx.JSendError					"Internal server error"
//	@Security		BearerAuth
//	@Router			/orders/{id}/complete-delivery [post]
func (h *OrderHandler) CompleteDelivery(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Verify user is a driver
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != middleware.RoleDriver {
		httpx.RespondError(w, http.StatusForbidden, "Solo los conductores pueden completar entregas")
		return
	}

	// Parse order ID from path
	orderIDStr := r.PathValue("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "ID de orden inválido",
		})
		return
	}

	// Verify order belongs to this driver
	if err := h.orderService.VerifyOrderBelongsToDriver(orderID, userID); err != nil {
		httpx.RespondError(w, http.StatusForbidden, err.Error())
		return
	}

	// Parse request body
	var req models.VerifyDeliveryCodeRequest
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

	// Verify delivery code
	isValid, err := h.orderService.VerifyDeliveryCode(orderID, req.DeliveryCode)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !isValid {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"delivery_code": "Código de entrega incorrecto",
		})
		return
	}

	// Code is valid, update status to delivered
	if err := h.orderService.UpdateOrderStatus(orderID, "delivered"); err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get updated order
	order, err := h.orderService.GetOrderByID(orderID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, order)
}

// UpdateOrderStatus godoc
//
//	@Summary		Update order status
//	@Description	Update the status of an order (driver only). For 'delivered' status, delivery code must be verified first.
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Order ID (UUID)"
//	@Param			request	body		models.UpdateOrderStatusRequest	true	"New status"
//	@Success		200		{object}	models.OrderResponse			"Order status updated"
//	@Failure		400		{object}	httpx.JSendFail					"Validation failed"
//	@Failure		401		{object}	httpx.JSendError				"Unauthorized"
//	@Failure		403		{object}	httpx.JSendError				"Forbidden"
//	@Failure		404		{object}	httpx.JSendFail					"Order not found"
//	@Failure		500		{object}	httpx.JSendError				"Internal server error"
//	@Security		BearerAuth
//	@Router			/orders/{id} [patch]
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Verify user is a driver
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != middleware.RoleDriver {
		httpx.RespondError(w, http.StatusForbidden, "Solo los conductores pueden actualizar el estado de órdenes")
		return
	}

	// Parse order ID from path
	orderIDStr := r.PathValue("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"id": "ID de orden inválido",
		})
		return
	}

	// Verify order belongs to this driver
	if err := h.orderService.VerifyOrderBelongsToDriver(orderID, userID); err != nil {
		httpx.RespondError(w, http.StatusForbidden, err.Error())
		return
	}

	// Parse request body
	var req models.UpdateOrderStatusRequest
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

	// For delivered status, we need to verify delivery code first
	// This prevents drivers from marking orders as delivered without proper verification
	if req.Status == "delivered" {
		httpx.RespondError(w, http.StatusBadRequest, "Para marcar como entregado, primero debe verificar el código de entrega usando el endpoint /orders/{id}/verify-delivery-code")
		return
	}

	// Update order status
	if err := h.orderService.UpdateOrderStatus(orderID, req.Status); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated order
	order, err := h.orderService.GetOrderByID(orderID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, order)
}

// GetMyActiveOrder godoc
//
//	@Summary		Get my active order
//	@Description	Get the active order for the authenticated driver
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.OrderResponse	"Active order retrieved"
//	@Failure		401	{object}	httpx.JSendError		"Unauthorized"
//	@Failure		404	{object}	httpx.JSendFail			"No active order"
//	@Failure		500	{object}	httpx.JSendError		"Internal server error"
//	@Security		BearerAuth
//	@Router			/drivers/me/active-order [get]
func (h *OrderHandler) GetMyActiveOrder(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Verify user is a driver
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != middleware.RoleDriver {
		httpx.RespondError(w, http.StatusForbidden, "Solo los conductores pueden ver su orden activa")
		return
	}

	// Get active order
	order, err := h.orderService.GetActiveOrderByDriver(userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if order == nil {
		httpx.RespondFail(w, http.StatusNotFound, map[string]any{
			"order": "No tienes ninguna orden activa",
		})
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, order)
}

// GetMyPendingAssignments godoc
//
//	@Summary		Get my pending assignments
//	@Description	Get all pending order assignments for the authenticated driver
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	httpx.JSendSuccess{data=[]models.OrderAssignment}	"Pending assignments list"
//	@Failure		401	{object}	httpx.JSendError								"Unauthorized"
//	@Failure		403	{object}	httpx.JSendError								"Forbidden - not a driver"
//	@Failure		500	{object}	httpx.JSendError								"Internal server error"
//	@Router			/drivers/me/assignments [get]
func (h *OrderHandler) GetMyPendingAssignments(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Verify user is a driver
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != middleware.RoleDriver {
		httpx.RespondError(w, http.StatusForbidden, "Solo los conductores pueden ver sus asignaciones")
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

// ListOrders godoc
//
//	@Summary		List orders
//	@Description	Get paginated list of orders (filtered by role and optional status)
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			status	query		string	false	"Filter by status"
//	@Param			limit	query		int		false	"Number of items per page (default: 20, max: 100)"	minimum(1)	maximum(100)	default(20)
//	@Param			page	query		int		false	"Page number (default: 1)"	minimum(1)	default(1)
//	@Security		BearerAuth
//	@Success		200	{object}	models.OrderListResponse	"Paginated orders list"
//	@Failure		400	{object}	httpx.JSendFail				"Invalid parameters"
//	@Failure		401	{object}	httpx.JSendError			"Unauthorized"
//	@Failure		500	{object}	httpx.JSendError			"Internal server error"
//	@Router			/orders [get]
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	// Get user ID and role from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "Rol de usuario inválido")
		return
	}

	// Parse pagination parameters
	pagination, err := httpx.ParsePaginationParams(r)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"pagination": err.Error(),
		})
		return
	}

	// Get optional status filter
	status := r.URL.Query().Get("status")

	var orders []models.Order
	var total int

	// Filter by role
	switch userRole {
	case "driver":
		orders, total, err = h.orderService.GetOrdersByDriverPaginated(userID, status, pagination.Limit, pagination.Offset)
	case "merchant":
		// TODO: Get merchant_id from users table or merchant profile
		// For now, return empty list or all orders if admin
		orders, total, err = h.orderService.GetOrdersByDriverPaginated(userID, status, pagination.Limit, pagination.Offset)
	case "admin":
		// Admins can see all orders
		orders, total, err = h.orderService.GetAllOrdersPaginated(status, pagination.Limit, pagination.Offset)
	default:
		httpx.RespondError(w, http.StatusForbidden, "Acceso denegado")
		return
	}

	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build pagination metadata
	paginationMeta := httpx.BuildPaginationMetadata(pagination.Page, pagination.Limit, total, "/api/v1/orders")

	// Return paginated response
	httpx.RespondSuccessWithPagination(w, http.StatusOK, orders, paginationMeta)
}
