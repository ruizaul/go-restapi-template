package models

import (
	"time"

	"github.com/google/uuid"
)

// DeviceType represents the type of device
type DeviceType string

const (
	// DeviceTypeAndroid represents an Android device
	DeviceTypeAndroid DeviceType = "android"
	// DeviceTypeIOS represents an iOS device
	DeviceTypeIOS DeviceType = "ios"
	// DeviceTypeWeb represents a web browser
	DeviceTypeWeb DeviceType = "web"
)

// FCMToken represents a Firebase Cloud Messaging token for a user's device
type FCMToken struct {
	ID         uuid.UUID  `json:"id" example:"d53b655c-e833-400e-b0e8-ee68ea18e2cc"`
	UserID     uuid.UUID  `json:"user_id" example:"d53b655c-e833-400e-b0e8-ee68ea18e2cc"`
	CreatedAt  time.Time  `json:"created_at" example:"2025-01-15T10:00:00Z"`
	UpdatedAt  time.Time  `json:"updated_at" example:"2025-01-15T10:00:00Z"`
	LastUsedAt time.Time  `json:"last_used_at" example:"2025-01-15T10:00:00Z"`
	DeviceID   *string    `json:"device_id,omitempty" example:"device-abc-123"`
	Token      string     `json:"token" example:"fL8X9Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y8Z9A0B1C"`
	DeviceType DeviceType `json:"device_type" enums:"android,ios,web" example:"android"`
	IsActive   bool       `json:"is_active" example:"true"`
}

// RegisterTokenRequest represents the request to register a FCM token
type RegisterTokenRequest struct {
	DeviceID   *string    `json:"device_id,omitempty" example:"device-abc-123"`
	Token      string     `json:"token" binding:"required" example:"fL8X9Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y8Z9A0B1C"`
	DeviceType DeviceType `json:"device_type" binding:"required,oneof=android ios web" enums:"android,ios,web" example:"android"`
}

// UnregisterTokenRequest represents the request to unregister a FCM token
type UnregisterTokenRequest struct {
	Token string `json:"token" binding:"required" example:"fL8X9Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y8Z9A0B1C"`
}

// TokenResponse wraps a FCM token in JSend format
type TokenResponse struct {
	Status string   `json:"status" example:"success"`
	Data   FCMToken `json:"data"`
}
