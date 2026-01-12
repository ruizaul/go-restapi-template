package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"tacoshare-delivery-api/internal/orders/models"

	"github.com/google/uuid"
)

const (
	sqlAndStatusParam = " AND status = $2"
)

// OrderRepository handles database operations for orders
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create creates a new order
func (r *OrderRepository) Create(order *models.Order) error {
	// Marshal items to JSON
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	query := `
		INSERT INTO orders (
			external_order_id, merchant_id, customer_name, customer_phone,
			pickup_address, pickup_latitude, pickup_longitude, pickup_instructions,
			delivery_address, delivery_latitude, delivery_longitude, delivery_instructions,
			delivery_code, items, total_amount, delivery_fee, status, distance_km, estimated_duration_minutes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(
		query,
		order.ExternalOrderID,
		order.MerchantID,
		order.CustomerName,
		order.CustomerPhone,
		order.PickupAddress,
		order.PickupLatitude,
		order.PickupLongitude,
		order.PickupInstructions,
		order.DeliveryAddress,
		order.DeliveryLatitude,
		order.DeliveryLongitude,
		order.DeliveryInstructions,
		order.DeliveryCode,
		itemsJSON,
		order.TotalAmount,
		order.DeliveryFee,
		order.Status,
		order.DistanceKm,
		order.EstimatedDurationMinutes,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// FindByID finds an order by ID
func (r *OrderRepository) FindByID(id uuid.UUID) (*models.Order, error) {
	query := `
		SELECT id, external_order_id, merchant_id, driver_id, customer_name, customer_phone,
			pickup_address, pickup_latitude, pickup_longitude, pickup_instructions,
			delivery_address, delivery_latitude, delivery_longitude, delivery_instructions,
			delivery_code, items, total_amount, delivery_fee, status, distance_km, estimated_duration_minutes,
			created_at, updated_at, assigned_at, accepted_at, picked_up_at, delivered_at,
			cancelled_at, cancellation_reason, cancelled_by
		FROM orders
		WHERE id = $1
	`

	order := &models.Order{}
	var itemsJSON []byte
	var cancellationReason sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&order.ID,
		&order.ExternalOrderID,
		&order.MerchantID,
		&order.DriverID,
		&order.CustomerName,
		&order.CustomerPhone,
		&order.PickupAddress,
		&order.PickupLatitude,
		&order.PickupLongitude,
		&order.PickupInstructions,
		&order.DeliveryAddress,
		&order.DeliveryLatitude,
		&order.DeliveryLongitude,
		&order.DeliveryInstructions,
		&order.DeliveryCode,
		&itemsJSON,
		&order.TotalAmount,
		&order.DeliveryFee,
		&order.Status,
		&order.DistanceKm,
		&order.EstimatedDurationMinutes,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.AssignedAt,
		&order.AcceptedAt,
		&order.PickedUpAt,
		&order.DeliveredAt,
		&order.CancelledAt,
		&cancellationReason,
		&order.CancelledBy,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	order.Items = json.RawMessage(itemsJSON)

	// Convert nullable cancellation_reason
	if cancellationReason.Valid {
		order.CancellationReason = &cancellationReason.String
	}

	return order, nil
}

// UpdateStatus updates the order status
func (r *OrderRepository) UpdateStatus(id uuid.UUID, status models.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// AssignDriver assigns a driver to an order
func (r *OrderRepository) AssignDriver(orderID, driverID uuid.UUID) error {
	query := `
		UPDATE orders
		SET driver_id = $1, status = $2, assigned_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := r.db.Exec(query, driverID, models.OrderStatusAssigned, orderID)
	if err != nil {
		return fmt.Errorf("failed to assign driver: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// UpdateAccepted updates order status to accepted and sets accepted timestamp
func (r *OrderRepository) UpdateAccepted(id uuid.UUID) error {
	query := `
		UPDATE orders
		SET status = $1, accepted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(query, models.OrderStatusAccepted, id)
	if err != nil {
		return fmt.Errorf("failed to update order to accepted: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// Cancel cancels an order
func (r *OrderRepository) Cancel(id uuid.UUID, cancelledBy uuid.UUID, reason string) error {
	query := `
		UPDATE orders
		SET status = $1, cancelled_at = CURRENT_TIMESTAMP, cancelled_by = $2,
			cancellation_reason = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	result, err := r.db.Exec(query, models.OrderStatusCancelled, cancelledBy, reason, id)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// FindByMerchantID finds all orders for a merchant
func (r *OrderRepository) FindByMerchantID(merchantID uuid.UUID, status string) ([]models.Order, error) {
	query := `
		SELECT id, external_order_id, merchant_id, driver_id, customer_name, customer_phone,
			pickup_address, pickup_latitude, pickup_longitude, pickup_instructions,
			delivery_address, delivery_latitude, delivery_longitude, delivery_instructions,
			delivery_code, items, total_amount, delivery_fee, status, distance_km, estimated_duration_minutes,
			created_at, updated_at, assigned_at, accepted_at, picked_up_at, delivered_at,
			cancelled_at, cancellation_reason, cancelled_by
		FROM orders
		WHERE merchant_id = $1
	`

	args := []any{merchantID}
	if status != "" {
		query += sqlAndStatusParam
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	return r.scanOrders(rows)
}

// FindByDriverID finds all orders for a driver
func (r *OrderRepository) FindByDriverID(driverID uuid.UUID, status string) ([]models.Order, error) {
	query := `
		SELECT id, external_order_id, merchant_id, driver_id, customer_name, customer_phone,
			pickup_address, pickup_latitude, pickup_longitude, pickup_instructions,
			delivery_address, delivery_latitude, delivery_longitude, delivery_instructions,
			delivery_code, items, total_amount, delivery_fee, status, distance_km, estimated_duration_minutes,
			created_at, updated_at, assigned_at, accepted_at, picked_up_at, delivered_at,
			cancelled_at, cancellation_reason, cancelled_by
		FROM orders
		WHERE driver_id = $1
	`

	args := []any{driverID}
	if status != "" {
		query += sqlAndStatusParam
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	return r.scanOrders(rows)
}

// FindActiveOrderByDriverID finds the active order for a driver
// Only returns orders created within the last 24 hours to prevent returning zombie orders
func (r *OrderRepository) FindActiveOrderByDriverID(driverID uuid.UUID) (*models.Order, error) {
	query := `
		SELECT id, external_order_id, merchant_id, driver_id, customer_name, customer_phone,
			pickup_address, pickup_latitude, pickup_longitude, pickup_instructions,
			delivery_address, delivery_latitude, delivery_longitude, delivery_instructions,
			delivery_code, items, total_amount, delivery_fee, status, distance_km, estimated_duration_minutes,
			created_at, updated_at, assigned_at, accepted_at, picked_up_at, delivered_at,
			cancelled_at, cancellation_reason, cancelled_by
		FROM orders
		WHERE driver_id = $1
		  AND status IN ('assigned', 'accepted', 'picked_up', 'in_transit')
		  AND created_at > NOW() - INTERVAL '24 hours'
		ORDER BY created_at DESC
		LIMIT 1
	`

	order := &models.Order{}
	var itemsJSON []byte
	var cancellationReason sql.NullString

	err := r.db.QueryRow(query, driverID).Scan(
		&order.ID,
		&order.ExternalOrderID,
		&order.MerchantID,
		&order.DriverID,
		&order.CustomerName,
		&order.CustomerPhone,
		&order.PickupAddress,
		&order.PickupLatitude,
		&order.PickupLongitude,
		&order.PickupInstructions,
		&order.DeliveryAddress,
		&order.DeliveryLatitude,
		&order.DeliveryLongitude,
		&order.DeliveryInstructions,
		&order.DeliveryCode,
		&itemsJSON,
		&order.TotalAmount,
		&order.DeliveryFee,
		&order.Status,
		&order.DistanceKm,
		&order.EstimatedDurationMinutes,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.AssignedAt,
		&order.AcceptedAt,
		&order.PickedUpAt,
		&order.DeliveredAt,
		&order.CancelledAt,
		&cancellationReason,
		&order.CancelledBy,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error finding active order: %w", err)
	}

	order.Items = json.RawMessage(itemsJSON)

	// Convert nullable cancellation_reason
	if cancellationReason.Valid {
		order.CancellationReason = &cancellationReason.String
	}

	return order, nil
}

// scanOrders scans multiple orders from rows
func (r *OrderRepository) scanOrders(rows *sql.Rows) ([]models.Order, error) {
	orders := []models.Order{}

	for rows.Next() {
		var order models.Order
		var itemsJSON []byte
		var cancellationReason sql.NullString

		err := rows.Scan(
			&order.ID,
			&order.ExternalOrderID,
			&order.MerchantID,
			&order.DriverID,
			&order.CustomerName,
			&order.CustomerPhone,
			&order.PickupAddress,
			&order.PickupLatitude,
			&order.PickupLongitude,
			&order.PickupInstructions,
			&order.DeliveryAddress,
			&order.DeliveryLatitude,
			&order.DeliveryLongitude,
			&order.DeliveryInstructions,
			&order.DeliveryCode,
			&itemsJSON,
			&order.TotalAmount,
			&order.DeliveryFee,
			&order.Status,
			&order.DistanceKm,
			&order.EstimatedDurationMinutes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.AssignedAt,
			&order.AcceptedAt,
			&order.PickedUpAt,
			&order.DeliveredAt,
			&order.CancelledAt,
			&cancellationReason,
			&order.CancelledBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		order.Items = json.RawMessage(itemsJSON)

		// Convert nullable cancellation_reason
		if cancellationReason.Valid {
			order.CancellationReason = &cancellationReason.String
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return orders, nil
}

// FindByDriverIDPaginated finds paginated orders for a driver
func (r *OrderRepository) FindByDriverIDPaginated(driverID uuid.UUID, status string, limit, offset int) ([]models.Order, int, error) {
	// Count total matching records
	countQuery := `SELECT COUNT(*) FROM orders WHERE driver_id = $1`
	countArgs := []any{driverID}
	if status != "" {
		countQuery += sqlAndStatusParam
		countArgs = append(countArgs, status)
	}

	var total int
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Fetch paginated records
	query := `
		SELECT id, external_order_id, merchant_id, driver_id, customer_name, customer_phone,
			pickup_address, pickup_latitude, pickup_longitude, pickup_instructions,
			delivery_address, delivery_latitude, delivery_longitude, delivery_instructions,
			delivery_code, items, total_amount, delivery_fee, status, distance_km, estimated_duration_minutes,
			created_at, updated_at, assigned_at, accepted_at, picked_up_at, delivered_at,
			cancelled_at, cancellation_reason, cancelled_by
		FROM orders
		WHERE driver_id = $1
	`

	args := []any{driverID}
	if status != "" {
		query += sqlAndStatusParam
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find orders: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	orders, err := r.scanOrders(rows)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// UpdateRouteInfo updates the distance and estimated duration for an order
func (r *OrderRepository) UpdateRouteInfo(orderID uuid.UUID, distanceKm float64, durationMins int) error {
	query := `
		UPDATE orders
		SET distance_km = $1, estimated_duration_minutes = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := r.db.Exec(query, distanceKm, durationMins, orderID)
	if err != nil {
		return fmt.Errorf("failed to update route info: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// FindAllPaginated finds all orders with pagination (admin only)
func (r *OrderRepository) FindAllPaginated(status string, limit, offset int) ([]models.Order, int, error) {
	// Count total matching records
	countQuery := `SELECT COUNT(*) FROM orders`
	countArgs := []any{}
	if status != "" {
		countQuery += " WHERE status = $1"
		countArgs = append(countArgs, status)
	}

	var total int
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Fetch paginated records
	query := `
		SELECT id, external_order_id, merchant_id, driver_id, customer_name, customer_phone,
			pickup_address, pickup_latitude, pickup_longitude, pickup_instructions,
			delivery_address, delivery_latitude, delivery_longitude, delivery_instructions,
			delivery_code, items, total_amount, delivery_fee, status, distance_km, estimated_duration_minutes,
			created_at, updated_at, assigned_at, accepted_at, picked_up_at, delivered_at,
			cancelled_at, cancellation_reason, cancelled_by
		FROM orders
	`

	args := []any{}
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find orders: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	orders, err := r.scanOrders(rows)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
