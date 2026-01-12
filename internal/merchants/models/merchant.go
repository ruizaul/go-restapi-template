package models

import (
	"time"

	"github.com/google/uuid"
)

// Merchant represents a business/merchant in the system
type Merchant struct {
	ID           uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID       uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	BusinessName string    `json:"business_name" example:"Tacos El Güero"`
	BusinessType string    `json:"business_type" example:"restaurant"`
	Phone        string    `json:"phone" example:"+525512345678"`
	Email        string    `json:"email,omitempty" example:"contacto@tacoselguero.com"`
	Address      string    `json:"address" example:"Av. Insurgentes Sur 1234, Col. Del Valle"`
	Latitude     float64   `json:"latitude" example:"19.432608"`
	Longitude    float64   `json:"longitude" example:"-99.133209"`
	City         string    `json:"city" example:"Ciudad de México"`
	State        string    `json:"state" example:"CDMX"`
	PostalCode   string    `json:"postal_code,omitempty" example:"03100"`
	Country      string    `json:"country" example:"MX"`
	Status       string    `json:"status" example:"active" enums:"active,inactive,suspended"`
	Rating       float64   `json:"rating" example:"4.5"`
	TotalOrders  int       `json:"total_orders" example:"125"`
	CreatedAt    time.Time `json:"created_at" example:"2025-01-15T10:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2025-01-15T10:00:00Z"`
}

// CreateMerchantRequest represents the request to create a merchant
type CreateMerchantRequest struct {
	BusinessName string  `json:"business_name" binding:"required,min=2,max=255" example:"Tacos El Güero"`
	BusinessType string  `json:"business_type" binding:"required" example:"restaurant"`
	Phone        string  `json:"phone" binding:"required,e164" example:"+525512345678"`
	Email        string  `json:"email,omitempty" binding:"omitempty,email" example:"contacto@tacoselguero.com"`
	Address      string  `json:"address" binding:"required" example:"Av. Insurgentes Sur 1234, Col. Del Valle"`
	Latitude     float64 `json:"latitude" binding:"required,min=-90,max=90" example:"19.432608"`
	Longitude    float64 `json:"longitude" binding:"required,min=-180,max=180" example:"-99.133209"`
	City         string  `json:"city" binding:"required" example:"Ciudad de México"`
	State        string  `json:"state" binding:"required" example:"CDMX"`
	PostalCode   string  `json:"postal_code,omitempty" example:"03100"`
}

// UpdateMerchantRequest represents the request to update merchant information
type UpdateMerchantRequest struct {
	BusinessName string  `json:"business_name,omitempty" binding:"omitempty,min=2,max=255" example:"Tacos El Güero"`
	Phone        string  `json:"phone,omitempty" binding:"omitempty,e164" example:"+525512345678"`
	Email        string  `json:"email,omitempty" binding:"omitempty,email" example:"contacto@tacoselguero.com"`
	Address      string  `json:"address,omitempty" example:"Av. Insurgentes Sur 1234, Col. Del Valle"`
	Latitude     float64 `json:"latitude,omitempty" binding:"omitempty,min=-90,max=90" example:"19.432608"`
	Longitude    float64 `json:"longitude,omitempty" binding:"omitempty,min=-180,max=180" example:"-99.133209"`
	PostalCode   string  `json:"postal_code,omitempty" example:"03100"`
}

// MerchantResponse wraps a merchant in JSend format
type MerchantResponse struct {
	Status string   `json:"status" example:"success"`
	Data   Merchant `json:"data"`
}

// MerchantListResponse wraps a list of merchants in JSend format
type MerchantListResponse struct {
	Status string     `json:"status" example:"success"`
	Data   []Merchant `json:"data"`
}
