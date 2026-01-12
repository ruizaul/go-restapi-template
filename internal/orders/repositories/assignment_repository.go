package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tacoshare-delivery-api/internal/orders/models"

	"github.com/google/uuid"
)

// AssignmentRepository handles database operations for order assignments
type AssignmentRepository struct {
	db *sql.DB
}

// NewAssignmentRepository creates a new assignment repository
func NewAssignmentRepository(db *sql.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// Create creates a new order assignment attempt
func (r *AssignmentRepository) Create(assignment *models.OrderAssignment) error {
	query := `
		INSERT INTO order_assignments (
			order_id, driver_id, attempt_number, search_radius_km,
			distance_to_pickup_km, estimated_arrival_minutes, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		assignment.OrderID,
		assignment.DriverID,
		assignment.AttemptNumber,
		assignment.SearchRadiusKm,
		assignment.DistanceToPickupKm,
		assignment.EstimatedArrivalMinutes,
		assignment.Status,
		assignment.ExpiresAt,
	).Scan(&assignment.ID, &assignment.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	return nil
}

// FindByID finds an assignment by ID
func (r *AssignmentRepository) FindByID(id uuid.UUID) (*models.OrderAssignment, error) {
	query := `
		SELECT id, order_id, driver_id, attempt_number, search_radius_km,
			distance_to_pickup_km, estimated_arrival_minutes, status,
			created_at, responded_at, expires_at, rejection_reason
		FROM order_assignments
		WHERE id = $1
	`

	assignment := &models.OrderAssignment{}
	var rejectionReason sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&assignment.ID,
		&assignment.OrderID,
		&assignment.DriverID,
		&assignment.AttemptNumber,
		&assignment.SearchRadiusKm,
		&assignment.DistanceToPickupKm,
		&assignment.EstimatedArrivalMinutes,
		&assignment.Status,
		&assignment.CreatedAt,
		&assignment.RespondedAt,
		&assignment.ExpiresAt,
		&rejectionReason,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find assignment: %w", err)
	}

	if rejectionReason.Valid {
		assignment.RejectionReason = &rejectionReason.String
	}

	return assignment, nil
}

// FindPendingByOrderAndDriver finds a pending assignment for an order and driver
func (r *AssignmentRepository) FindPendingByOrderAndDriver(orderID, driverID uuid.UUID) (*models.OrderAssignment, error) {
	query := `
		SELECT id, order_id, driver_id, attempt_number, search_radius_km,
			distance_to_pickup_km, estimated_arrival_minutes, status,
			created_at, responded_at, expires_at, rejection_reason
		FROM order_assignments
		WHERE order_id = $1 AND driver_id = $2 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`

	assignment := &models.OrderAssignment{}
	var rejectionReason sql.NullString

	err := r.db.QueryRow(query, orderID, driverID).Scan(
		&assignment.ID,
		&assignment.OrderID,
		&assignment.DriverID,
		&assignment.AttemptNumber,
		&assignment.SearchRadiusKm,
		&assignment.DistanceToPickupKm,
		&assignment.EstimatedArrivalMinutes,
		&assignment.Status,
		&assignment.CreatedAt,
		&assignment.RespondedAt,
		&assignment.ExpiresAt,
		&rejectionReason,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find pending assignment: %w", err)
	}

	if rejectionReason.Valid {
		assignment.RejectionReason = &rejectionReason.String
	}

	return assignment, nil
}

// UpdateStatus updates the status of an assignment
func (r *AssignmentRepository) UpdateStatus(id uuid.UUID, status models.AssignmentStatus) error {
	query := `
		UPDATE order_assignments
		SET status = $1, responded_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update assignment status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("assignment not found")
	}

	return nil
}

// UpdateStatusWithReason updates the status and rejection reason
func (r *AssignmentRepository) UpdateStatusWithReason(id uuid.UUID, status models.AssignmentStatus, reason string) error {
	query := `
		UPDATE order_assignments
		SET status = $1, rejection_reason = $2, responded_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := r.db.Exec(query, status, reason, id)
	if err != nil {
		return fmt.Errorf("failed to update assignment status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("assignment not found")
	}

	return nil
}

// ExpireOldAssignments marks all pending assignments that have expired as "timeout"
func (r *AssignmentRepository) ExpireOldAssignments() error {
	query := `
		UPDATE order_assignments
		SET status = 'timeout', responded_at = CURRENT_TIMESTAMP
		WHERE status = 'pending' AND expires_at < CURRENT_TIMESTAMP
	`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to expire old assignments: %w", err)
	}

	return nil
}

// ExpirePendingByOrderID marks all pending assignments for an order as "expired"
func (r *AssignmentRepository) ExpirePendingByOrderID(orderID uuid.UUID) error {
	query := `
		UPDATE order_assignments
		SET status = 'expired', responded_at = CURRENT_TIMESTAMP
		WHERE order_id = $1 AND status = 'pending'
	`

	_, err := r.db.Exec(query, orderID)
	if err != nil {
		return fmt.Errorf("failed to expire pending assignments: %w", err)
	}

	return nil
}

// GetNextAttemptNumber gets the next attempt number for an order
func (r *AssignmentRepository) GetNextAttemptNumber(orderID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(MAX(attempt_number), 0) + 1
		FROM order_assignments
		WHERE order_id = $1
	`

	var nextAttempt int
	err := r.db.QueryRow(query, orderID).Scan(&nextAttempt)
	if err != nil {
		return 0, fmt.Errorf("failed to get next attempt number: %w", err)
	}

	return nextAttempt, nil
}

// FindByOrderID finds all assignments for an order
func (r *AssignmentRepository) FindByOrderID(orderID uuid.UUID) ([]models.OrderAssignment, error) {
	query := `
		SELECT id, order_id, driver_id, attempt_number, search_radius_km,
			distance_to_pickup_km, estimated_arrival_minutes, status,
			created_at, responded_at, expires_at, rejection_reason
		FROM order_assignments
		WHERE order_id = $1
		ORDER BY attempt_number ASC, created_at ASC
	`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find assignments: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	assignments := []models.OrderAssignment{}
	for rows.Next() {
		var a models.OrderAssignment
		var rejectionReason sql.NullString

		err := rows.Scan(
			&a.ID,
			&a.OrderID,
			&a.DriverID,
			&a.AttemptNumber,
			&a.SearchRadiusKm,
			&a.DistanceToPickupKm,
			&a.EstimatedArrivalMinutes,
			&a.Status,
			&a.CreatedAt,
			&a.RespondedAt,
			&a.ExpiresAt,
			&rejectionReason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}

		if rejectionReason.Valid {
			a.RejectionReason = &rejectionReason.String
		}

		assignments = append(assignments, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return assignments, nil
}

// GetRejectedDriverIDsByOrderID returns a list of driver IDs who have already rejected,
// timed out, or had expired assignments for a specific order
func (r *AssignmentRepository) GetRejectedDriverIDsByOrderID(orderID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT DISTINCT driver_id
		FROM order_assignments
		WHERE order_id = $1
		  AND status IN ('rejected', 'timeout', 'expired')
		ORDER BY driver_id
	`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query rejected drivers: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	driverIDs := []uuid.UUID{}
	for rows.Next() {
		var driverID uuid.UUID
		if err := rows.Scan(&driverID); err != nil {
			return nil, fmt.Errorf("failed to scan driver ID: %w", err)
		}
		driverIDs = append(driverIDs, driverID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return driverIDs, nil
}

// WaitForResponse waits for a driver response or timeout
// NOTE: This method is DEPRECATED and will be removed. Use AssignmentWatcher instead.
// This method uses database polling which is inefficient (20 queries per assignment).
func (r *AssignmentRepository) WaitForResponse(assignmentID uuid.UUID, timeout time.Duration) (models.AssignmentStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Mark as timeout
			if err := r.UpdateStatus(assignmentID, models.AssignmentStatusTimeout); err != nil {
				return "", fmt.Errorf("failed to mark assignment as timeout: %w", err)
			}
			return models.AssignmentStatusTimeout, nil
		case <-ticker.C:
			assignment, err := r.FindByID(assignmentID)
			if err != nil {
				return "", fmt.Errorf("failed to check assignment status: %w", err)
			}
			if assignment == nil {
				return "", fmt.Errorf("assignment not found")
			}

			// Check if status changed from pending
			if assignment.Status != models.AssignmentStatusPending {
				return assignment.Status, nil
			}
		}
	}
}

// FindPendingByDriverID finds all pending assignments for a driver
func (r *AssignmentRepository) FindPendingByDriverID(driverID uuid.UUID) ([]*models.OrderAssignment, error) {
	query := `
		SELECT id, order_id, driver_id, attempt_number, search_radius_km,
			distance_to_pickup_km, estimated_arrival_minutes, status,
			created_at, responded_at, expires_at, rejection_reason
		FROM order_assignments
		WHERE driver_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, driverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending assignments: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	assignments := []*models.OrderAssignment{}
	for rows.Next() {
		assignment := &models.OrderAssignment{}
		var rejectionReason sql.NullString

		err := rows.Scan(
			&assignment.ID,
			&assignment.OrderID,
			&assignment.DriverID,
			&assignment.AttemptNumber,
			&assignment.SearchRadiusKm,
			&assignment.DistanceToPickupKm,
			&assignment.EstimatedArrivalMinutes,
			&assignment.Status,
			&assignment.CreatedAt,
			&assignment.RespondedAt,
			&assignment.ExpiresAt,
			&rejectionReason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}

		if rejectionReason.Valid {
			assignment.RejectionReason = &rejectionReason.String
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return assignments, nil
}
