package services

import (
	"fmt"

	"github.com/google/uuid"
	"tacoshare-delivery-api/internal/merchants/models"
	"tacoshare-delivery-api/internal/merchants/repositories"
)

// MerchantService handles business logic for merchants
type MerchantService struct {
	repo *repositories.MerchantRepository
}

// NewMerchantService creates a new merchant service
func NewMerchantService(repo *repositories.MerchantRepository) *MerchantService {
	return &MerchantService{repo: repo}
}

// CreateMerchant creates a new merchant for a user
func (s *MerchantService) CreateMerchant(userID uuid.UUID, req *models.CreateMerchantRequest) (*models.Merchant, error) {
	// Check if user already has a merchant
	existing, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("error checking existing merchant: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("el usuario ya tiene un negocio registrado")
	}

	merchant := &models.Merchant{
		UserID:       userID,
		BusinessName: req.BusinessName,
		BusinessType: req.BusinessType,
		Phone:        req.Phone,
		Email:        req.Email,
		Address:      req.Address,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		Country:      "MX",
	}

	if err := s.repo.Create(merchant); err != nil {
		return nil, fmt.Errorf("error creating merchant: %w", err)
	}

	return merchant, nil
}

// GetMerchantByID retrieves a merchant by ID
func (s *MerchantService) GetMerchantByID(id uuid.UUID) (*models.Merchant, error) {
	merchant, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding merchant: %w", err)
	}
	if merchant == nil {
		return nil, fmt.Errorf("negocio no encontrado")
	}
	return merchant, nil
}

// GetMerchantByUserID retrieves a merchant by user ID
func (s *MerchantService) GetMerchantByUserID(userID uuid.UUID) (*models.Merchant, error) {
	merchant, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("error finding merchant: %w", err)
	}
	if merchant == nil {
		return nil, fmt.Errorf("negocio no encontrado")
	}
	return merchant, nil
}

// UpdateMerchant updates merchant information
func (s *MerchantService) UpdateMerchant(merchantID uuid.UUID, req *models.UpdateMerchantRequest) (*models.Merchant, error) {
	merchant, err := s.repo.FindByID(merchantID)
	if err != nil {
		return nil, fmt.Errorf("error finding merchant: %w", err)
	}
	if merchant == nil {
		return nil, fmt.Errorf("negocio no encontrado")
	}

	// Update only provided fields
	if req.BusinessName != "" {
		merchant.BusinessName = req.BusinessName
	}
	if req.Phone != "" {
		merchant.Phone = req.Phone
	}
	if req.Email != "" {
		merchant.Email = req.Email
	}
	if req.Address != "" {
		merchant.Address = req.Address
	}
	if req.Latitude != 0 {
		merchant.Latitude = req.Latitude
	}
	if req.Longitude != 0 {
		merchant.Longitude = req.Longitude
	}
	if req.PostalCode != "" {
		merchant.PostalCode = req.PostalCode
	}

	if err := s.repo.Update(merchant); err != nil {
		return nil, fmt.Errorf("error updating merchant: %w", err)
	}

	return merchant, nil
}

// GetAllMerchants retrieves all merchants with optional filters
func (s *MerchantService) GetAllMerchants(city, businessType, status string) ([]models.Merchant, error) {
	merchants, err := s.repo.FindAll(city, businessType, status)
	if err != nil {
		return nil, fmt.Errorf("error finding merchants: %w", err)
	}
	return merchants, nil
}

// UpdateMerchantStatus updates the status of a merchant
func (s *MerchantService) UpdateMerchantStatus(merchantID uuid.UUID, status string) error {
	validStatuses := map[string]bool{"active": true, "inactive": true, "suspended": true}
	if !validStatuses[status] {
		return fmt.Errorf("estado inválido: debe ser active, inactive o suspended")
	}

	if err := s.repo.UpdateStatus(merchantID, status); err != nil {
		return fmt.Errorf("error updating merchant status: %w", err)
	}

	return nil
}
