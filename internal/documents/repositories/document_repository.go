package repositories

import (
	"database/sql"
	"tacoshare-delivery-api/internal/documents/models"

	"github.com/google/uuid"
)

// DocumentRepository handles data access for user documents
type DocumentRepository struct {
	db *sql.DB
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create creates a new user document record
func (r *DocumentRepository) Create(userID uuid.UUID, doc *models.UserDocument) error {
	query := `
		INSERT INTO user_documents (
			id, user_id, vehicle_brand, vehicle_model, license_plate,
			circulation_card_url, ine_front_url, ine_back_url,
			driver_license_front_url, driver_license_back_url, profile_photo_url,
			fiscal_name, fiscal_rfc, fiscal_zip_code, fiscal_regime,
			fiscal_street, fiscal_ext_number, fiscal_int_number,
			fiscal_neighborhood, fiscal_city, fiscal_state, fiscal_certificate_url,
			reviewed, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25
		)
	`

	_, err := r.db.Exec(
		query,
		doc.ID,
		userID,
		doc.VehicleBrand,
		doc.VehicleModel,
		doc.LicensePlate,
		doc.CirculationCardURL,
		doc.INEFrontURL,
		doc.INEBackURL,
		doc.DriverLicenseFrontURL,
		doc.DriverLicenseBackURL,
		doc.ProfilePhotoURL,
		doc.FiscalName,
		doc.FiscalRFC,
		doc.FiscalZipCode,
		doc.FiscalRegime,
		doc.FiscalStreet,
		doc.FiscalExtNumber,
		doc.FiscalIntNumber,
		doc.FiscalNeighborhood,
		doc.FiscalCity,
		doc.FiscalState,
		doc.FiscalCertificateURL,
		doc.Reviewed,
		doc.CreatedAt,
		doc.UpdatedAt,
	)

	return err
}

// FindByUserID finds a user's document record
func (r *DocumentRepository) FindByUserID(userID uuid.UUID) (*models.UserDocument, error) {
	query := `
		SELECT
			id, user_id, vehicle_brand, vehicle_model, license_plate,
			circulation_card_url, ine_front_url, ine_back_url,
			driver_license_front_url, driver_license_back_url, profile_photo_url,
			fiscal_name, fiscal_rfc, fiscal_zip_code, fiscal_regime,
			fiscal_street, fiscal_ext_number, fiscal_int_number,
			fiscal_neighborhood, fiscal_city, fiscal_state, fiscal_certificate_url,
			reviewed, created_at, updated_at
		FROM user_documents
		WHERE user_id = $1
	`

	doc := &models.UserDocument{}
	var fiscalRegime sql.NullString

	err := r.db.QueryRow(query, userID).Scan(
		&doc.ID,
		&doc.UserID,
		&doc.VehicleBrand,
		&doc.VehicleModel,
		&doc.LicensePlate,
		&doc.CirculationCardURL,
		&doc.INEFrontURL,
		&doc.INEBackURL,
		&doc.DriverLicenseFrontURL,
		&doc.DriverLicenseBackURL,
		&doc.ProfilePhotoURL,
		&doc.FiscalName,
		&doc.FiscalRFC,
		&doc.FiscalZipCode,
		&fiscalRegime,
		&doc.FiscalStreet,
		&doc.FiscalExtNumber,
		&doc.FiscalIntNumber,
		&doc.FiscalNeighborhood,
		&doc.FiscalCity,
		&doc.FiscalState,
		&doc.FiscalCertificateURL,
		&doc.Reviewed,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Convert fiscal_regime from string to enum
	if fiscalRegime.Valid {
		regime := models.FiscalRegime(fiscalRegime.String)
		doc.FiscalRegime = &regime
	}

	return doc, nil
}

// Update updates a user's document record
func (r *DocumentRepository) Update(userID uuid.UUID, doc *models.UpdateDocumentRequest) error {
	query := `
		UPDATE user_documents SET
			vehicle_brand = COALESCE($1, vehicle_brand),
			vehicle_model = COALESCE($2, vehicle_model),
			license_plate = COALESCE($3, license_plate),
			circulation_card_url = COALESCE($4, circulation_card_url),
			ine_front_url = COALESCE($5, ine_front_url),
			ine_back_url = COALESCE($6, ine_back_url),
			driver_license_front_url = COALESCE($7, driver_license_front_url),
			driver_license_back_url = COALESCE($8, driver_license_back_url),
			profile_photo_url = COALESCE($9, profile_photo_url),
			fiscal_name = COALESCE($10, fiscal_name),
			fiscal_rfc = COALESCE($11, fiscal_rfc),
			fiscal_zip_code = COALESCE($12, fiscal_zip_code),
			fiscal_regime = COALESCE($13, fiscal_regime),
			fiscal_street = COALESCE($14, fiscal_street),
			fiscal_ext_number = COALESCE($15, fiscal_ext_number),
			fiscal_int_number = COALESCE($16, fiscal_int_number),
			fiscal_neighborhood = COALESCE($17, fiscal_neighborhood),
			fiscal_city = COALESCE($18, fiscal_city),
			fiscal_state = COALESCE($19, fiscal_state),
			fiscal_certificate_url = COALESCE($20, fiscal_certificate_url),
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $21
	`

	_, err := r.db.Exec(
		query,
		doc.VehicleBrand,
		doc.VehicleModel,
		doc.LicensePlate,
		doc.CirculationCardURL,
		doc.INEFrontURL,
		doc.INEBackURL,
		doc.DriverLicenseFrontURL,
		doc.DriverLicenseBackURL,
		doc.ProfilePhotoURL,
		doc.FiscalName,
		doc.FiscalRFC,
		doc.FiscalZipCode,
		doc.FiscalRegime,
		doc.FiscalStreet,
		doc.FiscalExtNumber,
		doc.FiscalIntNumber,
		doc.FiscalNeighborhood,
		doc.FiscalCity,
		doc.FiscalState,
		doc.FiscalCertificateURL,
		userID,
	)

	return err
}

// Delete deletes a user's document record
func (r *DocumentRepository) Delete(userID uuid.UUID) error {
	query := `DELETE FROM user_documents WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

// MarkAsReviewed marks a user's documents as reviewed (admin only)
func (r *DocumentRepository) MarkAsReviewed(userID uuid.UUID, reviewed bool) error {
	query := `
		UPDATE user_documents
		SET reviewed = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`
	_, err := r.db.Exec(query, reviewed, userID)
	return err
}

// FindAll retrieves all user documents with pagination (admin only)
func (r *DocumentRepository) FindAll(limit, offset int) ([]*models.UserDocument, int, error) {
	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM user_documents`
	err := r.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT
			id, user_id, vehicle_brand, vehicle_model, license_plate,
			circulation_card_url, ine_front_url, ine_back_url,
			driver_license_front_url, driver_license_back_url, profile_photo_url,
			fiscal_name, fiscal_rfc, fiscal_zip_code, fiscal_regime,
			fiscal_street, fiscal_ext_number, fiscal_int_number,
			fiscal_neighborhood, fiscal_city, fiscal_state, fiscal_certificate_url,
			reviewed, created_at, updated_at
		FROM user_documents
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	//nolint:errcheck // rows.Close() error is not critical in defer
	defer func() { _ = rows.Close() }()

	var documents []*models.UserDocument

	for rows.Next() {
		doc := &models.UserDocument{}
		var fiscalRegime sql.NullString

		err := rows.Scan(
			&doc.ID,
			&doc.UserID,
			&doc.VehicleBrand,
			&doc.VehicleModel,
			&doc.LicensePlate,
			&doc.CirculationCardURL,
			&doc.INEFrontURL,
			&doc.INEBackURL,
			&doc.DriverLicenseFrontURL,
			&doc.DriverLicenseBackURL,
			&doc.ProfilePhotoURL,
			&doc.FiscalName,
			&doc.FiscalRFC,
			&doc.FiscalZipCode,
			&fiscalRegime,
			&doc.FiscalStreet,
			&doc.FiscalExtNumber,
			&doc.FiscalIntNumber,
			&doc.FiscalNeighborhood,
			&doc.FiscalCity,
			&doc.FiscalState,
			&doc.FiscalCertificateURL,
			&doc.Reviewed,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Convert fiscal_regime from string to enum
		if fiscalRegime.Valid {
			regime := models.FiscalRegime(fiscalRegime.String)
			doc.FiscalRegime = &regime
		}

		documents = append(documents, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return documents, totalCount, nil
}

// FindByID retrieves a document by its ID (admin only)
func (r *DocumentRepository) FindByID(docID uuid.UUID) (*models.UserDocument, error) {
	query := `
		SELECT
			id, user_id, vehicle_brand, vehicle_model, license_plate,
			circulation_card_url, ine_front_url, ine_back_url,
			driver_license_front_url, driver_license_back_url, profile_photo_url,
			fiscal_name, fiscal_rfc, fiscal_zip_code, fiscal_regime,
			fiscal_street, fiscal_ext_number, fiscal_int_number,
			fiscal_neighborhood, fiscal_city, fiscal_state, fiscal_certificate_url,
			reviewed, created_at, updated_at
		FROM user_documents
		WHERE id = $1
	`

	doc := &models.UserDocument{}
	var fiscalRegime sql.NullString

	err := r.db.QueryRow(query, docID).Scan(
		&doc.ID,
		&doc.UserID,
		&doc.VehicleBrand,
		&doc.VehicleModel,
		&doc.LicensePlate,
		&doc.CirculationCardURL,
		&doc.INEFrontURL,
		&doc.INEBackURL,
		&doc.DriverLicenseFrontURL,
		&doc.DriverLicenseBackURL,
		&doc.ProfilePhotoURL,
		&doc.FiscalName,
		&doc.FiscalRFC,
		&doc.FiscalZipCode,
		&fiscalRegime,
		&doc.FiscalStreet,
		&doc.FiscalExtNumber,
		&doc.FiscalIntNumber,
		&doc.FiscalNeighborhood,
		&doc.FiscalCity,
		&doc.FiscalState,
		&doc.FiscalCertificateURL,
		&doc.Reviewed,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Convert fiscal_regime from string to enum
	if fiscalRegime.Valid {
		regime := models.FiscalRegime(fiscalRegime.String)
		doc.FiscalRegime = &regime
	}

	return doc, nil
}

// UpdateByID updates a document by its ID (admin only)
func (r *DocumentRepository) UpdateByID(docID uuid.UUID, reviewed bool) error {
	query := `
		UPDATE user_documents
		SET reviewed = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	result, err := r.db.Exec(query, reviewed, docID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
