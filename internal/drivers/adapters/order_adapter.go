package adapters

import (
	"tacoshare-delivery-api/internal/drivers/services"
	orderRepos "tacoshare-delivery-api/internal/orders/repositories"

	"github.com/google/uuid"
)

// OrderRepositoryAdapter adapts the full OrderRepository to the minimal interface needed by LocationService
type OrderRepositoryAdapter struct {
	orderRepo *orderRepos.OrderRepository
}

// NewOrderRepositoryAdapter creates a new order repository adapter
func NewOrderRepositoryAdapter(orderRepo *orderRepos.OrderRepository) *OrderRepositoryAdapter {
	return &OrderRepositoryAdapter{orderRepo: orderRepo}
}

// FindActiveOrderByDriverID finds the active order for a driver (adapter implementation)
func (a *OrderRepositoryAdapter) FindActiveOrderByDriverID(driverID uuid.UUID) (*services.OrderInfo, error) {
	order, err := a.orderRepo.FindActiveOrderByDriverID(driverID)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return nil, nil
	}

	// Map to minimal OrderInfo struct
	return &services.OrderInfo{
		ID:                order.ID,
		DriverID:          order.DriverID,
		Status:            string(order.Status),
		DeliveryLatitude:  order.DeliveryLatitude,
		DeliveryLongitude: order.DeliveryLongitude,
	}, nil
}

// UpdateRouteInfo updates the distance and duration for an order (adapter implementation)
func (a *OrderRepositoryAdapter) UpdateRouteInfo(orderID uuid.UUID, distanceKm float64, durationMins int) error {
	return a.orderRepo.UpdateRouteInfo(orderID, distanceKm, durationMins)
}

// Compile-time check to ensure OrderRepositoryAdapter implements services.OrderRepository
var _ services.OrderRepository = (*OrderRepositoryAdapter)(nil)
