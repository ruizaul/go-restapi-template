package models

import (
	"time"

	"github.com/google/uuid"
)

// AssignmentStatus represents the status of an order assignment attempt
type AssignmentStatus string

const (
	AssignmentStatusPending  AssignmentStatus = "pending"
	AssignmentStatusAccepted AssignmentStatus = "accepted"
	AssignmentStatusRejected AssignmentStatus = "rejected"
	AssignmentStatusTimeout  AssignmentStatus = "timeout"
	AssignmentStatusExpired  AssignmentStatus = "expired"
)

// OrderAssignment represents an attempt to assign an order to a driver
type OrderAssignment struct {
	ID                      uuid.UUID        `json:"id"`
	OrderID                 uuid.UUID        `json:"order_id"`
	DriverID                uuid.UUID        `json:"driver_id"`
	AttemptNumber           int              `json:"attempt_number"`
	SearchRadiusKm          float64          `json:"search_radius_km"`
	DistanceToPickupKm      float64          `json:"distance_to_pickup_km"`
	EstimatedArrivalMinutes *int             `json:"estimated_arrival_minutes,omitempty"`
	Status                  AssignmentStatus `json:"status"`
	CreatedAt               time.Time        `json:"created_at"`
	RespondedAt             *time.Time       `json:"responded_at,omitempty"`
	ExpiresAt               time.Time        `json:"expires_at"`
	RejectionReason         *string          `json:"rejection_reason,omitempty"`
}

// DriverWithDistance represents a driver and their distance to a location
type DriverWithDistance struct {
	DriverID                uuid.UUID
	DriverName              string
	Latitude                float64
	Longitude               float64
	DistanceToPickupKm      float64
	EstimatedArrivalMinutes int
}
