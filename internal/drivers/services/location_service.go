package services

import (
	"context"
	"fmt"
	"tacoshare-delivery-api/internal/drivers/models"
	"tacoshare-delivery-api/internal/drivers/repositories"
	"tacoshare-delivery-api/internal/websockets/models/ws"

	"github.com/google/uuid"
)

// OrderRepository defines minimal interface needed for location service
type OrderRepository interface {
	FindActiveOrderByDriverID(driverID uuid.UUID) (*OrderInfo, error)
	UpdateRouteInfo(orderID uuid.UUID, distanceKm float64, durationMins int) error
}

// OrderInfo contains minimal order information needed for route recalculation
type OrderInfo struct {
	ID                uuid.UUID
	DriverID          *uuid.UUID
	Status            string
	DeliveryLatitude  float64
	DeliveryLongitude float64
}

// WebSocketHub defines minimal interface needed for broadcasting updates
type WebSocketHub interface {
	SendToUser(userID uuid.UUID, message *ws.WSMessage) error
}

// LocationService handles business logic for driver locations
type LocationService struct {
	repo           *repositories.LocationRepository
	orderRepo      OrderRepository
	routeRecalcSvc *RouteRecalculationService
	wsHub          WebSocketHub
}

// NewLocationService creates a new location service
func NewLocationService(
	repo *repositories.LocationRepository,
	orderRepo OrderRepository,
	routeRecalcSvc *RouteRecalculationService,
	wsHub WebSocketHub,
) *LocationService {
	return &LocationService{
		repo:           repo,
		orderRepo:      orderRepo,
		routeRecalcSvc: routeRecalcSvc,
		wsHub:          wsHub,
	}
}

// UpdateLocation updates or creates a driver's location
// and intelligently recalculates route if driver has an active order
func (s *LocationService) UpdateLocation(driverID uuid.UUID, req *models.UpdateLocationRequest) (*models.DriverLocation, error) {
	location := &models.DriverLocation{
		DriverID:       driverID,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Heading:        req.Heading,
		SpeedKmh:       req.SpeedKmh,
		AccuracyMeters: req.AccuracyMeters,
		IsAvailable:    true, // Default to available when updating location
	}

	if err := s.repo.Upsert(location); err != nil {
		return nil, fmt.Errorf("error updating driver location: %w", err)
	}

	// Attempt route recalculation if driver has active order (non-blocking)
	go s.attemptRouteRecalculation(driverID, req.Latitude, req.Longitude)

	return location, nil
}

// attemptRouteRecalculation attempts to recalculate route for active order (runs async)
func (s *LocationService) attemptRouteRecalculation(driverID uuid.UUID, currentLat, currentLng float64) {
	// Check if services are available (optional dependencies)
	if s.orderRepo == nil || s.routeRecalcSvc == nil {
		return // Skip if not configured
	}

	ctx := context.Background()

	// Find active order for this driver
	activeOrder, err := s.orderRepo.FindActiveOrderByDriverID(driverID)
	if err != nil {
		return
	}

	if activeOrder == nil {
		// No active order, clear cache if exists
		s.routeRecalcSvc.ClearCache(driverID)
		return
	}

	// Only recalculate for orders in transit (driver is on the way to delivery)
	if activeOrder.Status != "picked_up" && activeOrder.Status != "in_transit" {
		return
	}

	// Recalculate route from current position to delivery location
	result, err := s.routeRecalcSvc.RecalculateRoute(
		ctx,
		driverID,
		currentLat,
		currentLng,
		activeOrder.DeliveryLatitude,
		activeOrder.DeliveryLongitude,
	)
	if err != nil {
		return
	}

	// Only update DB and broadcast if change is significant
	if !result.ShouldUpdate {
		return
	}

	// Update order with new distance/ETA
	if err := s.orderRepo.UpdateRouteInfo(activeOrder.ID, result.NewDistanceKm, result.NewDurationMinutes); err != nil {
		return
	}

	// Broadcast ETA update via WebSocket (if hub is available)
	if s.wsHub != nil {
		s.broadcastETAUpdate(activeOrder, result)
	}
}

// broadcastETAUpdate sends ETA update to customer via WebSocket
func (s *LocationService) broadcastETAUpdate(order *OrderInfo, result *RecalculationResult) {
	message := &ws.WSMessage{
		Type: "eta_update",
		Data: map[string]interface{}{
			"order_id":           order.ID,
			"new_distance_km":    result.NewDistanceKm,
			"new_eta_minutes":    result.NewDurationMinutes,
			"distance_change_km": result.DistanceChange,
			"eta_change_minutes": result.DurationChange,
		},
	}

	// TODO: Send to customer (need customer_id from order)
	// For now, just send to driver
	if order.DriverID != nil {
		_ = s.wsHub.SendToUser(*order.DriverID, message)
	}
}

// GetMyLocation retrieves the current driver's location
func (s *LocationService) GetMyLocation(driverID uuid.UUID) (*models.DriverLocation, error) {
	location, err := s.repo.FindByDriverID(driverID)
	if err != nil {
		return nil, fmt.Errorf("error finding driver location: %w", err)
	}
	if location == nil {
		return nil, fmt.Errorf("ubicación del conductor no encontrada")
	}
	return location, nil
}

// UpdateAvailability updates a driver's availability status
func (s *LocationService) UpdateAvailability(driverID uuid.UUID, isAvailable bool) error {
	// Check if location exists first
	location, err := s.repo.FindByDriverID(driverID)
	if err != nil {
		return fmt.Errorf("error finding driver location: %w", err)
	}

	// If location doesn't exist, create it with default coordinates (0, 0)
	if location == nil {
		location = &models.DriverLocation{
			DriverID:    driverID,
			Latitude:    0.0,
			Longitude:   0.0,
			IsAvailable: isAvailable,
		}
		if err := s.repo.Upsert(location); err != nil {
			return fmt.Errorf("error creating driver location: %w", err)
		}
		return nil
	}

	// Update existing location
	if err := s.repo.UpdateAvailability(driverID, isAvailable); err != nil {
		return fmt.Errorf("error updating driver availability: %w", err)
	}

	return nil
}

// FindAvailableDriversNearby finds available drivers within a radius
func (s *LocationService) FindAvailableDriversNearby(lat, lng, radiusKm float64) ([]models.DriverWithInfo, error) {
	drivers, err := s.repo.FindAvailableInRadius(lat, lng, radiusKm)
	if err != nil {
		return nil, fmt.Errorf("error finding available drivers: %w", err)
	}
	return drivers, nil
}

// GetAllDriverLocations retrieves all driver locations (admin only)
func (s *LocationService) GetAllDriverLocations(availableOnly bool) ([]models.DriverWithInfo, error) {
	drivers, err := s.repo.FindAll(availableOnly)
	if err != nil {
		return nil, fmt.Errorf("error finding driver locations: %w", err)
	}
	return drivers, nil
}
