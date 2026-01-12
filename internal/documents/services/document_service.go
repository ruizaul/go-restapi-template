package services

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"tacoshare-delivery-api/internal/documents/models"
	"tacoshare-delivery-api/internal/documents/repositories"

	"github.com/google/uuid"
)

const (
	// Error messages
	errDocumentsNotFound = "documentos no encontrados"
)

// DocumentService handles business logic for user documents
type DocumentService struct {
	documentRepo *repositories.DocumentRepository
	userRepo     UserRepository
}

// UserRepository interface for user validation (adapter pattern)
type UserRepository interface {
	FindByID(id uuid.UUID) (*User, error)
}

// User represents a minimal user for validation purposes
type User struct {
	ID uuid.UUID
}

// NewDocumentService creates a new document service
func NewDocumentService(documentRepo *repositories.DocumentRepository, userRepo UserRepository) *DocumentService {
	return &DocumentService{
		documentRepo: documentRepo,
		userRepo:     userRepo,
	}
}

// CreateDocument creates a new user document record
//
//nolint:gocyclo // Complex document creation with multiple validation steps
func (s *DocumentService) CreateDocument(userID uuid.UUID, req *models.CreateDocumentRequest) (*models.UserDocument, error) {
	// Validate that user exists
	if s.userRepo != nil {
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			return nil, err
		}
		if user == nil || user.ID == uuid.Nil {
			return nil, errors.New("usuario no encontrado")
		}
	}

	// Check if document already exists for this user
	existing, err := s.documentRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("el usuario ya tiene documentos registrados")
	}

	// Validate RFC format if provided
	if req.FiscalRFC != nil && *req.FiscalRFC != "" {
		if err := validateRFC(*req.FiscalRFC); err != nil {
			return nil, err
		}
	}

	// Validate ZIP code format if provided
	if req.FiscalZipCode != nil && *req.FiscalZipCode != "" {
		if err := validateZipCode(*req.FiscalZipCode); err != nil {
			return nil, err
		}
	}

	// Validate fiscal regime if provided
	if req.FiscalRegime != nil {
		if err := validateFiscalRegime(*req.FiscalRegime); err != nil {
			return nil, err
		}
	}

	// Create document
	docID := uuid.New()
	doc := &models.UserDocument{
		ID:                    docID,
		UserID:                userID,
		VehicleBrand:          req.VehicleBrand,
		VehicleModel:          req.VehicleModel,
		LicensePlate:          req.LicensePlate,
		CirculationCardURL:    req.CirculationCardURL,
		INEFrontURL:           req.INEFrontURL,
		INEBackURL:            req.INEBackURL,
		DriverLicenseFrontURL: req.DriverLicenseFrontURL,
		DriverLicenseBackURL:  req.DriverLicenseBackURL,
		ProfilePhotoURL:       req.ProfilePhotoURL,
		FiscalName:            req.FiscalName,
		FiscalRFC:             req.FiscalRFC,
		FiscalZipCode:         req.FiscalZipCode,
		FiscalRegime:          req.FiscalRegime,
		FiscalStreet:          req.FiscalStreet,
		FiscalExtNumber:       req.FiscalExtNumber,
		FiscalIntNumber:       req.FiscalIntNumber,
		FiscalNeighborhood:    req.FiscalNeighborhood,
		FiscalCity:            req.FiscalCity,
		FiscalState:           req.FiscalState,
		FiscalCertificateURL:  req.FiscalCertificateURL,
		Reviewed:              false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := s.documentRepo.Create(userID, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

// GetDocumentByUserID retrieves a user's document record
func (s *DocumentService) GetDocumentByUserID(userID uuid.UUID) (*models.UserDocument, error) {
	doc, err := s.documentRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.New(errDocumentsNotFound)
	}
	return doc, nil
}

// GetDocumentStatus retrieves the verification status of a user's documents
func (s *DocumentService) GetDocumentStatus(userID uuid.UUID) (*models.DocumentStatusData, error) {
	doc, err := s.documentRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	// If no documents exist, return has_documents=false, reviewed=false
	if doc == nil {
		return &models.DocumentStatusData{
			HasDocuments: false,
			Reviewed:     false,
		}, nil
	}

	// Return status with actual review state
	return &models.DocumentStatusData{
		HasDocuments: true,
		Reviewed:     doc.Reviewed,
	}, nil
}

// UpdateDocument updates or creates a user's document record (upsert - partial update)
//
//nolint:gocyclo // Complex upsert logic with multiple validation paths
func (s *DocumentService) UpdateDocument(userID uuid.UUID, req *models.UpdateDocumentRequest) (*models.UserDocument, error) {
	// Validate user exists
	if s.userRepo != nil {
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			return nil, err
		}
		if user == nil || user.ID == uuid.Nil {
			return nil, errors.New("usuario no encontrado")
		}
	}

	// Check if document exists
	existing, err := s.documentRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	// If document doesn't exist, create it with provided data
	if existing == nil {
		// Validate RFC format if provided
		if req.FiscalRFC != nil && *req.FiscalRFC != "" {
			if err := validateRFC(*req.FiscalRFC); err != nil {
				return nil, err
			}
		}

		// Validate ZIP code format if provided
		if req.FiscalZipCode != nil && *req.FiscalZipCode != "" {
			if err := validateZipCode(*req.FiscalZipCode); err != nil {
				return nil, err
			}
		}

		// Validate fiscal regime if provided
		if req.FiscalRegime != nil {
			if err := validateFiscalRegime(*req.FiscalRegime); err != nil {
				return nil, err
			}
		}

		// Create new document with provided fields
		docID := uuid.New()
		doc := &models.UserDocument{
			ID:                    docID,
			UserID:                userID,
			VehicleBrand:          req.VehicleBrand,
			VehicleModel:          req.VehicleModel,
			LicensePlate:          req.LicensePlate,
			CirculationCardURL:    req.CirculationCardURL,
			INEFrontURL:           req.INEFrontURL,
			INEBackURL:            req.INEBackURL,
			DriverLicenseFrontURL: req.DriverLicenseFrontURL,
			DriverLicenseBackURL:  req.DriverLicenseBackURL,
			ProfilePhotoURL:       req.ProfilePhotoURL,
			FiscalName:            req.FiscalName,
			FiscalRFC:             req.FiscalRFC,
			FiscalZipCode:         req.FiscalZipCode,
			FiscalRegime:          req.FiscalRegime,
			FiscalStreet:          req.FiscalStreet,
			FiscalExtNumber:       req.FiscalExtNumber,
			FiscalIntNumber:       req.FiscalIntNumber,
			FiscalNeighborhood:    req.FiscalNeighborhood,
			FiscalCity:            req.FiscalCity,
			FiscalState:           req.FiscalState,
			FiscalCertificateURL:  req.FiscalCertificateURL,
			Reviewed:              false,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		if err := s.documentRepo.Create(userID, doc); err != nil {
			return nil, err
		}

		return doc, nil
	}

	// Document exists, update it

	// Validate RFC format if provided
	if req.FiscalRFC != nil && *req.FiscalRFC != "" {
		if err := validateRFC(*req.FiscalRFC); err != nil {
			return nil, err
		}
	}

	// Validate ZIP code format if provided
	if req.FiscalZipCode != nil && *req.FiscalZipCode != "" {
		if err := validateZipCode(*req.FiscalZipCode); err != nil {
			return nil, err
		}
	}

	// Validate fiscal regime if provided
	if req.FiscalRegime != nil {
		if err := validateFiscalRegime(*req.FiscalRegime); err != nil {
			return nil, err
		}
	}

	// Update document
	if err := s.documentRepo.Update(userID, req); err != nil {
		return nil, err
	}

	// Fetch updated document
	return s.documentRepo.FindByUserID(userID)
}

// DeleteDocument deletes a user's document record
func (s *DocumentService) DeleteDocument(userID uuid.UUID) error {
	// Check if document exists
	existing, err := s.documentRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New(errDocumentsNotFound)
	}

	return s.documentRepo.Delete(userID)
}

// MarkAsReviewed marks a user's documents as reviewed (admin only)
func (s *DocumentService) MarkAsReviewed(userID uuid.UUID, reviewed bool) error {
	// Check if document exists
	existing, err := s.documentRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New(errDocumentsNotFound)
	}

	return s.documentRepo.MarkAsReviewed(userID, reviewed)
}

// validateRFC validates Mexican RFC format (13 alphanumeric characters)
func validateRFC(rfc string) error {
	// RFC format: 4 letters + 6 digits + 3 alphanumeric characters
	// Example: PEGJ850101ABC
	pattern := `^[A-ZÑ&]{3,4}\d{6}[A-Z0-9]{3}$`
	matched, err := regexp.MatchString(pattern, rfc)
	if err != nil {
		return fmt.Errorf("error al validar RFC: %w", err)
	}
	if !matched {
		return fmt.Errorf("formato de RFC inválido (debe tener 13 caracteres alfanuméricos)")
	}
	return nil
}

// validateZipCode validates Mexican ZIP code format (5 digits)
func validateZipCode(zipCode string) error {
	pattern := `^\d{5}$`
	matched, err := regexp.MatchString(pattern, zipCode)
	if err != nil {
		return fmt.Errorf("error al validar código postal: %w", err)
	}
	if !matched {
		return fmt.Errorf("formato de código postal inválido (debe tener 5 dígitos)")
	}
	return nil
}

// validateFiscalRegime validates fiscal regime enum value
func validateFiscalRegime(regime models.FiscalRegime) error {
	validRegimes := map[models.FiscalRegime]bool{
		models.FiscalRegimeGeneral:               true,
		models.FiscalRegimeSimplificadoConfianza: true,
		models.FiscalRegimeActividadEmpresarial:  true,
		models.FiscalRegimeArrendamiento:         true,
		models.FiscalRegimeSalarios:              true,
		models.FiscalRegimeIncorporacionFiscal:   true,
	}

	if !validRegimes[regime] {
		return fmt.Errorf("régimen fiscal inválido - valores permitidos: general, simplificado_confianza, actividad_empresarial, arrendamiento, salarios, incorporacion_fiscal")
	}
	return nil
}

// GetAllDocuments retrieves all user documents with pagination (admin only)
func (s *DocumentService) GetAllDocuments(page, limit int) ([]*models.UserDocument, int, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	return s.documentRepo.FindAll(limit, offset)
}

// UpdateDocumentByID updates a document by its ID (admin only)
func (s *DocumentService) UpdateDocumentByID(docID uuid.UUID, reviewed bool) error {
	// Check if document exists
	existing, err := s.documentRepo.FindByID(docID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("documento no encontrado")
	}

	return s.documentRepo.UpdateByID(docID, reviewed)
}
