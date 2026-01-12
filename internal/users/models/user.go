package models

import (
	"time"

	"github.com/google/uuid"
)

// UserProfile represents a user's profile information
type UserProfile struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-15T10:30:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-15T10:30:00Z"`
	Name      string    `json:"name" example:"Juan Pérez"`
	Email     string    `json:"email" example:"juan.perez@example.com"`
	Phone     string    `json:"phone,omitempty" example:"+525512345678"`
	Role      string    `json:"role" enums:"customer,merchant,driver,admin" example:"customer"`
}

// UserProfileResponse wraps user profile in JSend format
type UserProfileResponse struct {
	Data   UserProfile `json:"data"`
	Status string      `json:"status" example:"success"`
}
