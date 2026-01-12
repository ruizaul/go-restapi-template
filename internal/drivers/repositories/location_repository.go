package repositories

import (
	"database/sql"
	"fmt"

	"tacoshare-delivery-api/internal/drivers/models"

	"github.com/google/uuid"
)

// LocationRepository handles database operations for driver locations
type LocationRepository struct {
	db *sql.DB
}

// NewLocationRepository creates a new location repository
func NewLocationRepository(db *sql.DB) *LocationRepository {
	return &LocationRepository{db: db}
}

// Upsert inserts or updates a driver's location (UPSERT pattern)
func (r *LocationRepository) Upsert(location *models.DriverLocation) error {
	query := `
		INSERT INTO driver_locations (
			driver_id, latitude, longitude, heading, speed_kmh, accuracy_meters, is_available
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (driver_id) DO UPDATE SET
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			heading = EXCLUDED.heading,
			speed_kmh = EXCLUDED.speed_kmh,
			accuracy_meters = EXCLUDED.accuracy_meters,
			is_available = EXCLUDED.is_available,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, updated_at
	`

	err := r.db.QueryRow(
		query,
		location.DriverID,
		location.Latitude,
		location.Longitude,
		location.Heading,
		location.SpeedKmh,
		location.AccuracyMeters,
		location.IsAvailable,
	).Scan(&location.ID, &location.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert driver location: %w", err)
	}

	return nil
}

// FindByDriverID finds a driver's current location
func (r *LocationRepository) FindByDriverID(driverID uuid.UUID) (*models.DriverLocation, error) {
	query := `
		SELECT id, driver_id, latitude, longitude, heading, speed_kmh,
			accuracy_meters, is_available, updated_at
		FROM driver_locations
		WHERE driver_id = $1
	`

	location := &models.DriverLocation{}
	err := r.db.QueryRow(query, driverID).Scan(
		&location.ID,
		&location.DriverID,
		&location.Latitude,
		&location.Longitude,
		&location.Heading,
		&location.SpeedKmh,
		&location.AccuracyMeters,
		&location.IsAvailable,
		&location.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find driver location: %w", err)
	}

	return location, nil
}

// UpdateAvailability updates a driver's availability status
func (r *LocationRepository) UpdateAvailability(driverID uuid.UUID, isAvailable bool) error {
	query := `
		UPDATE driver_locations
		SET is_available = $1, updated_at = CURRENT_TIMESTAMP
		WHERE driver_id = $2
	`

	result, err := r.db.Exec(query, isAvailable, driverID)
	if err != nil {
		return fmt.Errorf("failed to update driver availability: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("driver location not found")
	}

	return nil
}

// FindAvailableInRadius finds available drivers within a radius (in km) using Haversine formula
func (r *LocationRepository) FindAvailableInRadius(lat, lng, radiusKm float64) ([]models.DriverWithInfo, error) {
	// Haversine formula to calculate distance
	// Use subquery to calculate distance and filter in outer query
	query := `
		SELECT
			driver_id,
			name,
			phone,
			latitude,
			longitude,
			is_available,
			updated_at,
			distance_km
		FROM (
			SELECT
				dl.driver_id,
				u.name,
				u.phone,
				dl.latitude,
				dl.longitude,
				dl.is_available,
				dl.updated_at,
				(6371 * acos(
					cos(radians($1)) * cos(radians(dl.latitude)) *
					cos(radians(dl.longitude) - radians($2)) +
					sin(radians($1)) * sin(radians(dl.latitude))
				)) AS distance_km
			FROM driver_locations dl
			JOIN users u ON u.id = dl.driver_id
			WHERE dl.is_available = true
				AND u.role = 'driver'
				AND u.account_status = 'active'
		) AS drivers_with_distance
		WHERE distance_km <= $3
		ORDER BY distance_km ASC
	`

	rows, err := r.db.Query(query, lat, lng, radiusKm)
	if err != nil {
		return nil, fmt.Errorf("failed to find available drivers: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	drivers := []models.DriverWithInfo{}
	for rows.Next() {
		var driver models.DriverWithInfo
		var distanceKm float64

		err := rows.Scan(
			&driver.DriverID,
			&driver.Name,
			&driver.Phone,
			&driver.Latitude,
			&driver.Longitude,
			&driver.IsAvailable,
			&driver.UpdatedAt,
			&distanceKm,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver: %w", err)
		}

		drivers = append(drivers, driver)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return drivers, nil
}

// FindAll finds all driver locations (admin only)
func (r *LocationRepository) FindAll(availableOnly bool) ([]models.DriverWithInfo, error) {
	query := `
		SELECT
			dl.driver_id,
			u.name,
			u.phone,
			dl.latitude,
			dl.longitude,
			dl.is_available,
			dl.updated_at
		FROM driver_locations dl
		JOIN users u ON u.id = dl.driver_id
		WHERE u.role = 'driver'
	`

	if availableOnly {
		query += " AND dl.is_available = true"
	}

	query += " ORDER BY dl.updated_at DESC"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to find driver locations: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	drivers := []models.DriverWithInfo{}
	for rows.Next() {
		var driver models.DriverWithInfo

		err := rows.Scan(
			&driver.DriverID,
			&driver.Name,
			&driver.Phone,
			&driver.Latitude,
			&driver.Longitude,
			&driver.IsAvailable,
			&driver.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver: %w", err)
		}

		drivers = append(drivers, driver)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return drivers, nil
}
