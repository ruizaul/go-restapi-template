package models

import (
	"time"

	"github.com/google/uuid"
)

// FiscalRegime represents the tax regime types in Mexico (SAT)
type FiscalRegime string

const (
	// FiscalRegimeGeneral represents Régimen General de Ley Personas Morales
	FiscalRegimeGeneral FiscalRegime = "general"
	// FiscalRegimeSimplificadoConfianza represents Régimen Simplificado de Confianza (RESICO)
	FiscalRegimeSimplificadoConfianza FiscalRegime = "simplificado_confianza"
	// FiscalRegimeActividadEmpresarial represents Actividad Empresarial y Profesional
	FiscalRegimeActividadEmpresarial FiscalRegime = "actividad_empresarial"
	// FiscalRegimeArrendamiento represents Arrendamiento
	FiscalRegimeArrendamiento FiscalRegime = "arrendamiento"
	// FiscalRegimeSalarios represents Sueldos y Salarios
	FiscalRegimeSalarios FiscalRegime = "salarios"
	// FiscalRegimeIncorporacionFiscal represents Régimen de Incorporación Fiscal
	FiscalRegimeIncorporacionFiscal FiscalRegime = "incorporacion_fiscal"
)

// UserDocument represents user verification documents and fiscal information
type UserDocument struct {
	ID                    uuid.UUID     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID                uuid.UUID     `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CreatedAt             time.Time     `json:"created_at" example:"2025-01-15T10:30:00Z"`
	UpdatedAt             time.Time     `json:"updated_at" example:"2025-01-15T11:00:00Z"`
	VehicleBrand          *string       `json:"vehicle_brand,omitempty" example:"Honda"`
	VehicleModel          *string       `json:"vehicle_model,omitempty" example:"CBR 250"`
	LicensePlate          *string       `json:"license_plate,omitempty" example:"ABC-123-XYZ"`
	CirculationCardURL    *string       `json:"circulation_card_url,omitempty" example:"https://storage.example.com/docs/circulation_card.jpg"`
	INEFrontURL           *string       `json:"ine_front_url,omitempty" example:"https://storage.example.com/docs/ine_front.jpg"`
	INEBackURL            *string       `json:"ine_back_url,omitempty" example:"https://storage.example.com/docs/ine_back.jpg"`
	DriverLicenseFrontURL *string       `json:"driver_license_front_url,omitempty" example:"https://storage.example.com/docs/license_front.jpg"`
	DriverLicenseBackURL  *string       `json:"driver_license_back_url,omitempty" example:"https://storage.example.com/docs/license_back.jpg"`
	ProfilePhotoURL       *string       `json:"profile_photo_url,omitempty" example:"https://storage.example.com/docs/profile.jpg"`
	FiscalName            *string       `json:"fiscal_name,omitempty" example:"Juan Pérez García"`
	FiscalRFC             *string       `json:"fiscal_rfc,omitempty" example:"PEGJ850101ABC"`
	FiscalZipCode         *string       `json:"fiscal_zip_code,omitempty" example:"06600"`
	FiscalRegime          *FiscalRegime `json:"fiscal_regime,omitempty" example:"actividad_empresarial" enums:"general,simplificado_confianza,actividad_empresarial,arrendamiento,salarios,incorporacion_fiscal"`
	FiscalStreet          *string       `json:"fiscal_street,omitempty" example:"Paseo de la Reforma"`
	FiscalExtNumber       *string       `json:"fiscal_ext_number,omitempty" example:"222"`
	FiscalIntNumber       *string       `json:"fiscal_int_number,omitempty" example:"Piso 5"`
	FiscalNeighborhood    *string       `json:"fiscal_neighborhood,omitempty" example:"Juárez"`
	FiscalCity            *string       `json:"fiscal_city,omitempty" example:"Ciudad de México"`
	FiscalState           *string       `json:"fiscal_state,omitempty" example:"CDMX"`
	FiscalCertificateURL  *string       `json:"fiscal_certificate_url,omitempty" example:"https://storage.example.com/docs/fiscal_cert.pdf"`
	Reviewed              bool          `json:"reviewed" example:"false"`
}

// CreateDocumentRequest represents the request body for creating user documents
type CreateDocumentRequest struct {
	// Vehicle Information
	VehicleBrand *string `json:"vehicle_brand,omitempty" example:"Honda"`
	VehicleModel *string `json:"vehicle_model,omitempty" example:"CBR 250"`
	LicensePlate *string `json:"license_plate,omitempty" example:"ABC-123-XYZ"`

	// Document URLs
	CirculationCardURL    *string `json:"circulation_card_url,omitempty" example:"https://storage.example.com/docs/circulation_card.jpg"`
	INEFrontURL           *string `json:"ine_front_url,omitempty" example:"https://storage.example.com/docs/ine_front.jpg"`
	INEBackURL            *string `json:"ine_back_url,omitempty" example:"https://storage.example.com/docs/ine_back.jpg"`
	DriverLicenseFrontURL *string `json:"driver_license_front_url,omitempty" example:"https://storage.example.com/docs/license_front.jpg"`
	DriverLicenseBackURL  *string `json:"driver_license_back_url,omitempty" example:"https://storage.example.com/docs/license_back.jpg"`
	ProfilePhotoURL       *string `json:"profile_photo_url,omitempty" example:"https://storage.example.com/docs/profile.jpg"`

	// Fiscal Information
	FiscalName           *string       `json:"fiscal_name,omitempty" example:"Juan Pérez García"`
	FiscalRFC            *string       `json:"fiscal_rfc,omitempty" example:"PEGJ850101ABC"`
	FiscalZipCode        *string       `json:"fiscal_zip_code,omitempty" example:"06600"`
	FiscalRegime         *FiscalRegime `json:"fiscal_regime,omitempty" example:"actividad_empresarial" enums:"general,simplificado_confianza,actividad_empresarial,arrendamiento,salarios,incorporacion_fiscal"`
	FiscalStreet         *string       `json:"fiscal_street,omitempty" example:"Paseo de la Reforma"`
	FiscalExtNumber      *string       `json:"fiscal_ext_number,omitempty" example:"222"`
	FiscalIntNumber      *string       `json:"fiscal_int_number,omitempty" example:"Piso 5"`
	FiscalNeighborhood   *string       `json:"fiscal_neighborhood,omitempty" example:"Juárez"`
	FiscalCity           *string       `json:"fiscal_city,omitempty" example:"Ciudad de México"`
	FiscalState          *string       `json:"fiscal_state,omitempty" example:"CDMX"`
	FiscalCertificateURL *string       `json:"fiscal_certificate_url,omitempty" example:"https://storage.example.com/docs/fiscal_cert.pdf"`
}

// UpdateDocumentRequest represents the request body for updating user documents (partial update)
type UpdateDocumentRequest struct {
	// Vehicle Information
	VehicleBrand *string `json:"vehicle_brand,omitempty" example:"Honda"`
	VehicleModel *string `json:"vehicle_model,omitempty" example:"CBR 250"`
	LicensePlate *string `json:"license_plate,omitempty" example:"ABC-123-XYZ"`

	// Document URLs
	CirculationCardURL    *string `json:"circulation_card_url,omitempty" example:"https://storage.example.com/docs/circulation_card.jpg"`
	INEFrontURL           *string `json:"ine_front_url,omitempty" example:"https://storage.example.com/docs/ine_front.jpg"`
	INEBackURL            *string `json:"ine_back_url,omitempty" example:"https://storage.example.com/docs/ine_back.jpg"`
	DriverLicenseFrontURL *string `json:"driver_license_front_url,omitempty" example:"https://storage.example.com/docs/license_front.jpg"`
	DriverLicenseBackURL  *string `json:"driver_license_back_url,omitempty" example:"https://storage.example.com/docs/license_back.jpg"`
	ProfilePhotoURL       *string `json:"profile_photo_url,omitempty" example:"https://storage.example.com/docs/profile.jpg"`

	// Fiscal Information
	FiscalName           *string       `json:"fiscal_name,omitempty" example:"Juan Pérez García"`
	FiscalRFC            *string       `json:"fiscal_rfc,omitempty" example:"PEGJ850101ABC"`
	FiscalZipCode        *string       `json:"fiscal_zip_code,omitempty" example:"06600"`
	FiscalRegime         *FiscalRegime `json:"fiscal_regime,omitempty" example:"actividad_empresarial" enums:"general,simplificado_confianza,actividad_empresarial,arrendamiento,salarios,incorporacion_fiscal"`
	FiscalStreet         *string       `json:"fiscal_street,omitempty" example:"Paseo de la Reforma"`
	FiscalExtNumber      *string       `json:"fiscal_ext_number,omitempty" example:"222"`
	FiscalIntNumber      *string       `json:"fiscal_int_number,omitempty" example:"Piso 5"`
	FiscalNeighborhood   *string       `json:"fiscal_neighborhood,omitempty" example:"Juárez"`
	FiscalCity           *string       `json:"fiscal_city,omitempty" example:"Ciudad de México"`
	FiscalState          *string       `json:"fiscal_state,omitempty" example:"CDMX"`
	FiscalCertificateURL *string       `json:"fiscal_certificate_url,omitempty" example:"https://storage.example.com/docs/fiscal_cert.pdf"`
}

// DocumentResponse wraps a single document in JSend format
type DocumentResponse struct {
	Data   UserDocument `json:"data"`
	Status string       `json:"status" example:"success"`
}

// DocumentStatusResponse wraps document verification status in JSend format
type DocumentStatusResponse struct {
	Status string             `json:"status" example:"success"`
	Data   DocumentStatusData `json:"data"`
}

// DocumentStatusData represents the verification status of user documents
type DocumentStatusData struct {
	HasDocuments bool `json:"has_documents" example:"true"`
	Reviewed     bool `json:"reviewed" example:"true"`
}

// DocumentListResponse wraps a paginated list of documents in JSend format
type DocumentListResponse struct {
	Status string `json:"status" example:"success"`
	Data   struct {
		Items      []UserDocument     `json:"items"`
		Pagination PaginationMetadata `json:"pagination"`
	} `json:"data"`
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	NextURL     string `json:"next_url,omitempty" example:"/api/v1/documents?page=2&limit=20"`
	PreviousURL string `json:"previous_url,omitempty" example:"/api/v1/documents?page=1&limit=20"`
	CurrentPage int    `json:"current_page" example:"1"`
	PerPage     int    `json:"per_page" example:"20"`
	TotalItems  int    `json:"total_items" example:"100"`
	TotalPages  int    `json:"total_pages" example:"5"`
	HasNext     bool   `json:"has_next" example:"true"`
	HasPrevious bool   `json:"has_previous" example:"false"`
}
