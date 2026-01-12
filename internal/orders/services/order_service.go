package services

import (
	"context"
	"encoding/json"
	"fmt"

	"tacoshare-delivery-api/internal/orders/models"
	"tacoshare-delivery-api/internal/orders/repositories"
	"tacoshare-delivery-api/pkg/gmaps"

	"github.com/google/uuid"
)

// OrderService handles business logic for orders
type OrderService struct {
	repo        *repositories.OrderRepository
	gmapsClient *gmaps.Client
}

// NewOrderService creates a new order service
func NewOrderService(repo *repositories.OrderRepository, gmapsClient *gmaps.Client) *OrderService {
	return &OrderService{
		repo:        repo,
		gmapsClient: gmapsClient,
	}
}

// CreateExternalOrder creates a new order from an external backend
func (s *OrderService) CreateExternalOrder(req *models.CreateExternalOrderRequest) (*models.Order, error) {
	// Calculate distance first to validate it's within acceptable range
	var distanceKm *float64
	var estimatedDurationMinutes *int

	if s.gmapsClient != nil {
		ctx := context.Background()
		pickup := gmaps.Location{
			Latitude:  req.PickupLatitude,
			Longitude: req.PickupLongitude,
		}
		delivery := gmaps.Location{
			Latitude:  req.DeliveryLatitude,
			Longitude: req.DeliveryLongitude,
		}

		result, err := s.gmapsClient.CalculateDistance(ctx, pickup, delivery)
		if err != nil {
			return nil, fmt.Errorf("no se pudo obtener la distancia de entrega")
		}

		if result.DistanceKm > 3.0 {
			return nil, fmt.Errorf("la distancia de entrega (%.2f km) excede el límite máximo de 3 km", result.DistanceKm)
		}

		distanceKm = &result.DistanceKm
		estimatedDurationMinutes = &result.DurationMinutes
	} else {
		return nil, fmt.Errorf("servicio de validación de distancia no disponible")
	}

	// Marshal items to JSON
	itemsJSON, err := json.Marshal(req.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal items: %w", err)
	}

	order := &models.Order{
		ExternalOrderID:          req.ExternalOrderID,
		MerchantID:               req.MerchantID,
		CustomerName:             req.CustomerName,
		CustomerPhone:            req.CustomerPhone,
		PickupAddress:            req.PickupAddress,
		PickupLatitude:           req.PickupLatitude,
		PickupLongitude:          req.PickupLongitude,
		PickupInstructions:       req.PickupInstructions,
		DeliveryAddress:          req.DeliveryAddress,
		DeliveryLatitude:         req.DeliveryLatitude,
		DeliveryLongitude:        req.DeliveryLongitude,
		DeliveryInstructions:     req.DeliveryInstructions,
		DeliveryCode:             req.DeliveryCode,
		Items:                    json.RawMessage(itemsJSON),
		TotalAmount:              req.TotalAmount,
		DeliveryFee:              req.DeliveryFee,
		Status:                   models.OrderStatusSearchingDriver,
		DistanceKm:               distanceKm,
		EstimatedDurationMinutes: estimatedDurationMinutes,
	}

	if err := s.repo.Create(order); err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	return order, nil
}

// GetOrderByID retrieves an order by ID
func (s *OrderService) GetOrderByID(id uuid.UUID) (*models.Order, error) {
	order, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding order: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("orden no encontrada")
	}
	return order, nil
}

// GetOrdersByMerchant retrieves all orders for a merchant
func (s *OrderService) GetOrdersByMerchant(merchantID uuid.UUID, status string) ([]models.Order, error) {
	orders, err := s.repo.FindByMerchantID(merchantID, status)
	if err != nil {
		return nil, fmt.Errorf("error finding orders: %w", err)
	}
	return orders, nil
}

// GetOrdersByDriver retrieves all orders for a driver
func (s *OrderService) GetOrdersByDriver(driverID uuid.UUID, status string) ([]models.Order, error) {
	orders, err := s.repo.FindByDriverID(driverID, status)
	if err != nil {
		return nil, fmt.Errorf("error finding orders: %w", err)
	}
	return orders, nil
}

// GetActiveOrderByDriver retrieves the active order for a driver
func (s *OrderService) GetActiveOrderByDriver(driverID uuid.UUID) (*models.Order, error) {
	order, err := s.repo.FindActiveOrderByDriverID(driverID)
	if err != nil {
		return nil, fmt.Errorf("error finding active order: %w", err)
	}
	return order, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(orderID uuid.UUID, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"searching_driver":    true,
		"assigned":            true,
		"accepted":            true,
		"picked_up":           true,
		"in_transit":          true,
		"delivered":           true,
		"cancelled":           true,
		"no_driver_available": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("estado inválido: %s", status)
	}

	orderStatus := models.OrderStatus(status)
	if err := s.repo.UpdateStatus(orderID, orderStatus); err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}

	return nil
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(orderID, cancelledBy uuid.UUID, reason string) error {
	// Check if order exists
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("error finding order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("orden no encontrada")
	}

	// Check if order can be cancelled
	if order.Status == models.OrderStatusDelivered {
		return fmt.Errorf("no se puede cancelar una orden ya entregada")
	}
	if order.Status == models.OrderStatusCancelled {
		return fmt.Errorf("la orden ya está cancelada")
	}

	if err := s.repo.Cancel(orderID, cancelledBy, reason); err != nil {
		return fmt.Errorf("error cancelling order: %w", err)
	}

	return nil
}

// VerifyOrderBelongsToDriver verifies that an order belongs to a specific driver
func (s *OrderService) VerifyOrderBelongsToDriver(orderID, driverID uuid.UUID) error {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("error finding order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("orden no encontrada")
	}

	if order.DriverID == nil || *order.DriverID != driverID {
		return fmt.Errorf("esta orden no está asignada a este conductor")
	}

	return nil
}

// GetOrdersByDriverPaginated retrieves paginated orders for a driver
func (s *OrderService) GetOrdersByDriverPaginated(driverID uuid.UUID, status string, limit, offset int) ([]models.Order, int, error) {
	orders, total, err := s.repo.FindByDriverIDPaginated(driverID, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error finding orders: %w", err)
	}
	return orders, total, nil
}

// GetAllOrdersPaginated retrieves all orders with pagination (admin only)
func (s *OrderService) GetAllOrdersPaginated(status string, limit, offset int) ([]models.Order, int, error) {
	orders, total, err := s.repo.FindAllPaginated(status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error finding orders: %w", err)
	}
	return orders, total, nil
}

// VerifyDeliveryCode verifies if the provided delivery code matches the order's code
func (s *OrderService) VerifyDeliveryCode(orderID uuid.UUID, deliveryCode string) (bool, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return false, fmt.Errorf("error finding order: %w", err)
	}
	if order == nil {
		return false, fmt.Errorf("orden no encontrada")
	}

	// Check if order is in a valid state for delivery
	if order.Status != models.OrderStatusInTransit && order.Status != models.OrderStatusPickedUp {
		return false, fmt.Errorf("la orden no está en estado de entrega")
	}

	// Verify the code matches
	return order.DeliveryCode == deliveryCode, nil
}
