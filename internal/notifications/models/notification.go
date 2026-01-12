package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type/category of notification
type NotificationType string

const (
	// NotificationTypeOrderCreated indicates a new order was created
	NotificationTypeOrderCreated NotificationType = "order_created"
	// NotificationTypeOrderUpdated indicates an order was updated
	NotificationTypeOrderUpdated NotificationType = "order_updated"
	// NotificationTypeOrderAssigned indicates a driver was assigned to an order
	NotificationTypeOrderAssigned NotificationType = "order_assigned"
	// NotificationTypeOrderInTransit indicates an order is being delivered
	NotificationTypeOrderInTransit NotificationType = "order_in_transit"
	// NotificationTypeOrderDelivered indicates an order was delivered
	NotificationTypeOrderDelivered NotificationType = "order_delivered"
	// NotificationTypeOrderCanceled indicates an order was canceled
	NotificationTypeOrderCanceled NotificationType = "order_canceled"
	// NotificationTypePaymentReceived indicates a payment was received
	NotificationTypePaymentReceived NotificationType = "payment_received"
	// NotificationTypePaymentFailed indicates a payment failed
	NotificationTypePaymentFailed NotificationType = "payment_failed"
	// NotificationTypeDriverAssigned indicates a driver was assigned
	NotificationTypeDriverAssigned NotificationType = "driver_assigned"
	// NotificationTypeDriverNearby indicates a driver is nearby
	NotificationTypeDriverNearby NotificationType = "driver_nearby"
	// NotificationTypeGeneral represents a general notification
	NotificationTypeGeneral NotificationType = "general"
	// NotificationTypePromotional represents a promotional notification
	NotificationTypePromotional NotificationType = "promotional"
)

// Notification represents a push notification sent to a user
type Notification struct {
	ID               uuid.UUID        `json:"id" example:"d53b655c-e833-400e-b0e8-ee68ea18e2cc"`
	UserID           uuid.UUID        `json:"user_id" example:"d53b655c-e833-400e-b0e8-ee68ea18e2cc"`
	CreatedAt        time.Time        `json:"created_at" example:"2025-01-15T10:00:00Z"`
	ReadAt           *time.Time       `json:"read_at,omitempty" example:"2025-01-15T10:30:00Z"`
	Title            string           `json:"title" example:"Pedido en camino"`
	Body             string           `json:"body" example:"Tu pedido #1234 está en camino"`
	NotificationType NotificationType `json:"notification_type" enums:"order_created,order_updated,order_assigned,order_in_transit,order_delivered,order_canceled,payment_received,payment_failed,driver_assigned,driver_nearby,general,promotional" example:"order_in_transit"`
	Data             json.RawMessage  `json:"data,omitempty" swaggertype:"string"`
	IsRead           bool             `json:"is_read" example:"false"`
}

// CreateNotificationRequest represents the request to create a notification
type CreateNotificationRequest struct {
	UserID           uuid.UUID        `json:"user_id" binding:"required" example:"d53b655c-e833-400e-b0e8-ee68ea18e2cc"`
	Data             json.RawMessage  `json:"data,omitempty" swaggertype:"string"`
	Title            string           `json:"title" binding:"required,max=255" example:"Pedido en camino"`
	Body             string           `json:"body" binding:"required" example:"Tu pedido #1234 está en camino"`
	NotificationType NotificationType `json:"notification_type" binding:"required" enums:"order_created,order_updated,order_assigned,order_in_transit,order_delivered,order_canceled,payment_received,payment_failed,driver_assigned,driver_nearby,general,promotional" example:"order_in_transit"`
}

// NotificationListResponse wraps the paginated list of notifications in JSend format
type NotificationListResponse struct {
	Status string `json:"status" example:"success"`
	Data   struct {
		Items      []Notification     `json:"items"`
		Pagination PaginationMetadata `json:"pagination"`
	} `json:"data"`
}

// NotificationResponse wraps a single notification in JSend format
type NotificationResponse struct {
	Status string       `json:"status" example:"success"`
	Data   Notification `json:"data"`
}

// UnreadCountResponse represents the unread notification count
type UnreadCountResponse struct {
	Status string `json:"status" example:"success"`
	Data   struct {
		Count int `json:"count" example:"5"`
	} `json:"data"`
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	NextURL     string `json:"next_url,omitempty" example:"/api/v1/notifications?page=2&limit=20"`
	PreviousURL string `json:"previous_url,omitempty"`
	CurrentPage int    `json:"current_page" example:"1"`
	PerPage     int    `json:"per_page" example:"20"`
	TotalItems  int    `json:"total_items" example:"100"`
	TotalPages  int    `json:"total_pages" example:"5"`
	HasNext     bool   `json:"has_next" example:"true"`
	HasPrevious bool   `json:"has_previous" example:"false"`
}
