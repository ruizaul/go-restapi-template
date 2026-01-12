package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusSearchingDriver   OrderStatus = "searching_driver"
	OrderStatusAssigned          OrderStatus = "assigned"
	OrderStatusAccepted          OrderStatus = "accepted"
	OrderStatusPickedUp          OrderStatus = "picked_up"
	OrderStatusInTransit         OrderStatus = "in_transit"
	OrderStatusDelivered         OrderStatus = "delivered"
	OrderStatusCancelled         OrderStatus = "cancelled"
	OrderStatusNoDriverAvailable OrderStatus = "no_driver_available"
)

// OrderItem represents an item in an order
type OrderItem struct {
	Name     string  `json:"name" example:"Tacos al Pastor"`
	Quantity int     `json:"quantity" example:"3"`
	Price    float64 `json:"price" example:"15.00"`
}

// Order represents a delivery order
type Order struct {
	ID                       uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ExternalOrderID          string          `json:"external_order_id,omitempty" example:"EXT-ORDER-12345"`
	MerchantID               uuid.UUID       `json:"merchant_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	DriverID                 *uuid.UUID      `json:"driver_id,omitempty" example:"987e6543-e21b-12d3-a456-426614174000"`
	CustomerName             string          `json:"customer_name" example:"María García"`
	CustomerPhone            string          `json:"customer_phone" example:"+525512345678"`
	PickupAddress            string          `json:"pickup_address" example:"Av. Insurgentes Sur 1234"`
	PickupLatitude           float64         `json:"pickup_latitude" example:"19.432608"`
	PickupLongitude          float64         `json:"pickup_longitude" example:"-99.133209"`
	PickupInstructions       string          `json:"pickup_instructions,omitempty" example:"Local 5, planta baja"`
	DeliveryAddress          string          `json:"delivery_address" example:"Calle Reforma 567"`
	DeliveryLatitude         float64         `json:"delivery_latitude" example:"19.426608"`
	DeliveryLongitude        float64         `json:"delivery_longitude" example:"-99.166209"`
	DeliveryInstructions     string          `json:"delivery_instructions,omitempty" example:"Tocar interfon 302"`
	DeliveryCode             string          `json:"delivery_code" example:"1234"`
	Items                    json.RawMessage `json:"items" swaggertype:"array,object"`
	TotalAmount              float64         `json:"total_amount" example:"250.50"`
	DeliveryFee              float64         `json:"delivery_fee" example:"25.00"`
	Status                   OrderStatus     `json:"status" example:"searching_driver"`
	DistanceKm               *float64        `json:"distance_km,omitempty" example:"5.2"`
	EstimatedDurationMinutes *int            `json:"estimated_duration_minutes,omitempty" example:"15"`
	CreatedAt                time.Time       `json:"created_at" example:"2025-01-15T10:00:00Z"`
	UpdatedAt                time.Time       `json:"updated_at" example:"2025-01-15T10:00:00Z"`
	AssignedAt               *time.Time      `json:"assigned_at,omitempty" example:"2025-01-15T10:01:00Z"`
	AcceptedAt               *time.Time      `json:"accepted_at,omitempty" example:"2025-01-15T10:01:30Z"`
	PickedUpAt               *time.Time      `json:"picked_up_at,omitempty" example:"2025-01-15T10:15:00Z"`
	DeliveredAt              *time.Time      `json:"delivered_at,omitempty" example:"2025-01-15T10:30:00Z"`
	CancelledAt              *time.Time      `json:"cancelled_at,omitempty"`
	CancellationReason       *string         `json:"cancellation_reason,omitempty"`
	CancelledBy              *uuid.UUID      `json:"cancelled_by,omitempty"`
}

// CreateExternalOrderRequest represents an order coming from an external backend
type CreateExternalOrderRequest struct {
	ExternalOrderID      string      `json:"external_order_id" binding:"required" example:"EXT-ORDER-12345"`
	MerchantID           uuid.UUID   `json:"merchant_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	CustomerName         string      `json:"customer_name" binding:"required" example:"María García"`
	CustomerPhone        string      `json:"customer_phone" binding:"required,e164" example:"+525512345678"`
	PickupAddress        string      `json:"pickup_address" binding:"required" example:"Av. Insurgentes Sur 1234"`
	PickupLatitude       float64     `json:"pickup_latitude" binding:"required,min=-90,max=90" example:"19.432608"`
	PickupLongitude      float64     `json:"pickup_longitude" binding:"required,min=-180,max=180" example:"-99.133209"`
	PickupInstructions   string      `json:"pickup_instructions,omitempty" example:"Local 5"`
	DeliveryAddress      string      `json:"delivery_address" binding:"required" example:"Calle Reforma 567"`
	DeliveryLatitude     float64     `json:"delivery_latitude" binding:"required,min=-90,max=90" example:"19.426608"`
	DeliveryLongitude    float64     `json:"delivery_longitude" binding:"required,min=-180,max=180" example:"-99.166209"`
	DeliveryInstructions string      `json:"delivery_instructions,omitempty" example:"Tocar 302"`
	DeliveryCode         string      `json:"delivery_code" binding:"required,len=4,numeric" example:"1234"`
	Items                []OrderItem `json:"items" binding:"required,min=1"`
	TotalAmount          float64     `json:"total_amount" binding:"required,gt=0" example:"250.50"`
	DeliveryFee          float64     `json:"delivery_fee,omitempty" example:"25.00"`
}

// UpdateOrderStatusRequest represents a request to update order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required" example:"accepted"`
}

// CancelOrderRequest represents a request to cancel an order
type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required" example:"Cliente solicitó cancelación"`
}

// VerifyDeliveryCodeRequest represents a request to verify delivery code
type VerifyDeliveryCodeRequest struct {
	DeliveryCode string `json:"delivery_code" binding:"required,len=4,numeric" example:"1234"`
}

// OrderResponse wraps an order in JSend format
type OrderResponse struct {
	Status string `json:"status" example:"success"`
	Data   Order  `json:"data"`
}

// OrderListResponse wraps a list of orders in JSend format
type OrderListResponse struct {
	Status string  `json:"status" example:"success"`
	Data   []Order `json:"data"`
}
