package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Order events
	MessageTypeNewOrder          MessageType = "new_order"
	MessageTypeOrderAssigned     MessageType = "order_assigned"
	MessageTypeOrderAccepted     MessageType = "order_accepted"
	MessageTypeOrderRejected     MessageType = "order_rejected"
	MessageTypeOrderPickedUp     MessageType = "order_picked_up"
	MessageTypeOrderInTransit    MessageType = "order_in_transit"
	MessageTypeOrderDelivered    MessageType = "order_delivered"
	MessageTypeOrderCancelled    MessageType = "order_cancelled"
	MessageTypeOrderTimeout      MessageType = "order_timeout"
	MessageTypeNoDriverAvailable MessageType = "no_driver_available"

	// Driver events
	MessageTypeDriverLocationUpdate MessageType = "driver_location_update"
	MessageTypeDriverAvailable      MessageType = "driver_available"
	MessageTypeDriverUnavailable    MessageType = "driver_unavailable"

	// Connection events
	MessageTypeConnected    MessageType = "connected"
	MessageTypeDisconnected MessageType = "disconnected"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
	MessageTypeError        MessageType = "error"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type           MessageType     `json:"type"`
	Data           json.RawMessage `json:"data,omitempty"`
	Timestamp      time.Time       `json:"timestamp"`
	MessageID      string          `json:"message_id"`
	TimeoutSeconds *int            `json:"timeout_seconds,omitempty"` // Time in seconds before the message/action expires
}

// NewOrderData represents data for new_order event
type NewOrderData struct {
	OrderID              string  `json:"order_id"`
	AssignmentID         string  `json:"assignment_id"`
	CustomerName         string  `json:"customer_name"`
	PickupAddress        string  `json:"pickup_address"`
	DeliveryAddress      string  `json:"delivery_address"`
	DistanceKm           float64 `json:"distance_km"`
	EstimatedTimeMinutes int     `json:"estimated_time_minutes"`
	TotalAmount          float64 `json:"total_amount"`
	ExpiresAt            string  `json:"expires_at"`
	TimeoutSeconds       int     `json:"timeout_seconds"` // Time in seconds before assignment expires
	PickupLatitude       float64 `json:"pickup_latitude"`
	PickupLongitude      float64 `json:"pickup_longitude"`
	DeliveryLatitude     float64 `json:"delivery_latitude"`
	DeliveryLongitude    float64 `json:"delivery_longitude"`
	PickupInstructions   string  `json:"pickup_instructions,omitempty"`
	DeliveryInstructions string  `json:"delivery_instructions,omitempty"`
}

// OrderStatusData represents data for order status change events
type OrderStatusData struct {
	OrderID     string  `json:"order_id"`
	Status      string  `json:"status"`
	DriverID    string  `json:"driver_id,omitempty"`
	DriverName  string  `json:"driver_name,omitempty"`
	DriverPhone string  `json:"driver_phone,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	Message     string  `json:"message,omitempty"`
}

// DriverLocationData represents data for driver location updates
type DriverLocationData struct {
	DriverID  string  `json:"driver_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Heading   float64 `json:"heading,omitempty"`
	SpeedKmh  float64 `json:"speed_kmh,omitempty"`
	UpdatedAt string  `json:"updated_at"`
}

// ErrorData represents error message data
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ConnectedData represents connection confirmation data
type ConnectedData struct {
	ClientID string `json:"client_id"`
	UserID   string `json:"user_id,omitempty"`
	Role     string `json:"role,omitempty"`
	Message  string `json:"message"`
}

// NewWSMessage creates a new WebSocket message
func NewWSMessage(msgType MessageType, data any) (*WSMessage, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &WSMessage{
		Type:      msgType,
		Data:      dataJSON,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
	}, nil
}

// NewErrorMessage creates an error WebSocket message
func NewErrorMessage(code, message string) (*WSMessage, error) {
	return NewWSMessage(MessageTypeError, ErrorData{
		Code:    code,
		Message: message,
	})
}

// NewConnectedMessage creates a connected confirmation message
func NewConnectedMessage(clientID, userID, role string) (*WSMessage, error) {
	return NewWSMessage(MessageTypeConnected, ConnectedData{
		ClientID: clientID,
		UserID:   userID,
		Role:     role,
		Message:  "Conectado exitosamente al servidor WebSocket",
	})
}
