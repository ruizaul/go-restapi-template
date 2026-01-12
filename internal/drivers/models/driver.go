package models

import (
	"time"

	"github.com/google/uuid"
)

// DriverLocation represents a driver's real-time location
type DriverLocation struct {
	ID             uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	DriverID       uuid.UUID  `json:"driver_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Latitude       float64    `json:"latitude" example:"19.432608"`
	Longitude      float64    `json:"longitude" example:"-99.133209"`
	Heading        *float64   `json:"heading,omitempty" example:"45.5"`
	SpeedKmh       *float64   `json:"speed_kmh,omitempty" example:"35.2"`
	AccuracyMeters *float64   `json:"accuracy_meters,omitempty" example:"10.5"`
	IsAvailable    bool       `json:"is_available" example:"true"`
	UpdatedAt      time.Time  `json:"updated_at" example:"2025-01-15T10:00:00Z"`
}

// UpdateLocationRequest represents a request to update driver location
type UpdateLocationRequest struct {
	Latitude       float64  `json:"latitude" binding:"required,min=-90,max=90" example:"19.432608"`
	Longitude      float64  `json:"longitude" binding:"required,min=-180,max=180" example:"-99.133209"`
	Heading        *float64 `json:"heading,omitempty" binding:"omitempty,min=0,max=360" example:"45.5"`
	SpeedKmh       *float64 `json:"speed_kmh,omitempty" binding:"omitempty,min=0" example:"35.2"`
	AccuracyMeters *float64 `json:"accuracy_meters,omitempty" binding:"omitempty,min=0" example:"10.5"`
}

// UpdateAvailabilityRequest represents a request to update driver availability
type UpdateAvailabilityRequest struct {
	IsAvailable bool `json:"is_available" binding:"required" example:"true"`
}

// DriverLocationResponse wraps driver location in JSend format
type DriverLocationResponse struct {
	Status string         `json:"status" example:"success"`
	Data   DriverLocation `json:"data"`
}

// DriverWithInfo contains driver information with location
type DriverWithInfo struct {
	DriverID    uuid.UUID `json:"driver_id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	IsAvailable bool      `json:"is_available"`
	UpdatedAt   time.Time `json:"updated_at"`
}
