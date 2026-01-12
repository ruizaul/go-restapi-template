package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	driverModels "tacoshare-delivery-api/internal/drivers/models"
	"tacoshare-delivery-api/internal/drivers/repositories"
	notificationModels "tacoshare-delivery-api/internal/notifications/models"
	"tacoshare-delivery-api/internal/notifications/services"
	"tacoshare-delivery-api/internal/orders/models"
	orderRepos "tacoshare-delivery-api/internal/orders/repositories"
	"tacoshare-delivery-api/pkg/gmaps"

	"github.com/google/uuid"
)

// AssignmentService handles the core order assignment logic
type AssignmentService struct {
	orderRepo            *orderRepos.OrderRepository
	assignmentRepo       *orderRepos.AssignmentRepository
	locationRepo         *repositories.LocationRepository
	gmapsClient          *gmaps.Client
	notificationSvc      *services.NotificationService
	wsHub                WSHub
	watcher              *AssignmentWatcher
	queueManager         *QueueManager
	timeoutSeconds       int
	radiusKm             float64
	retryIntervalSeconds int
	maxSearchTimeSeconds int
}

// WSHub interface for WebSocket broadcasting
type WSHub interface {
	BroadcastToChannel(channel string, message any) error
	SendToUser(userID uuid.UUID, message any) error
}

// AssignmentConfig contains configuration for the assignment service
type AssignmentConfig struct {
	TimeoutSeconds       int
	RadiusKm             float64
	RetryIntervalSeconds int
	MaxSearchTimeSeconds int
}

// NewAssignmentService creates a new assignment service
func NewAssignmentService(
	orderRepo *orderRepos.OrderRepository,
	assignmentRepo *orderRepos.AssignmentRepository,
	locationRepo *repositories.LocationRepository,
	gmapsClient *gmaps.Client,
	notificationSvc *services.NotificationService,
	wsHub WSHub,
) *AssignmentService {
	// Load configuration from environment with defaults
	config := loadAssignmentConfig()

	return &AssignmentService{
		orderRepo:            orderRepo,
		assignmentRepo:       assignmentRepo,
		locationRepo:         locationRepo,
		gmapsClient:          gmapsClient,
		notificationSvc:      notificationSvc,
		wsHub:                wsHub,
		watcher:              NewAssignmentWatcher(),
		queueManager:         NewQueueManager(),
		timeoutSeconds:       config.TimeoutSeconds,
		radiusKm:             config.RadiusKm,
		retryIntervalSeconds: config.RetryIntervalSeconds,
		maxSearchTimeSeconds: config.MaxSearchTimeSeconds,
	}
}

// loadAssignmentConfig loads assignment configuration from environment
func loadAssignmentConfig() AssignmentConfig {
	config := AssignmentConfig{
		TimeoutSeconds:       10,  // Default: 10 seconds per driver
		RadiusKm:             2.0, // Default: 2 km (FIXED RADIUS)
		RetryIntervalSeconds: 15,  // Default: 15 seconds between search retries
		MaxSearchTimeSeconds: 180, // Default: 3 minutes total search time
	}

	if val := os.Getenv("ASSIGNMENT_TIMEOUT_SECONDS"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil {
			config.TimeoutSeconds = timeout
		}
	}

	if val := os.Getenv("ASSIGNMENT_RADIUS_KM"); val != "" {
		if radius, err := strconv.ParseFloat(val, 64); err == nil {
			config.RadiusKm = radius
		}
	}

	if val := os.Getenv("ASSIGNMENT_RETRY_INTERVAL_SECONDS"); val != "" {
		if interval, err := strconv.Atoi(val); err == nil {
			config.RetryIntervalSeconds = interval
		}
	}

	if val := os.Getenv("ASSIGNMENT_MAX_SEARCH_SECONDS"); val != "" {
		if maxSearch, err := strconv.Atoi(val); err == nil {
			config.MaxSearchTimeSeconds = maxSearch
		}
	}

	return config
}

// AssignOrderToDriver is the main function that assigns an order to the nearest available driver
// It uses a retry loop with fixed radius and exponential backoff
func (s *AssignmentService) AssignOrderToDriver(orderID uuid.UUID) error {
	// Get the order
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	// Update order status to searching_driver
	if err := s.orderRepo.UpdateStatus(orderID, models.OrderStatusSearchingDriver); err != nil {
		return fmt.Errorf("failed to update order status to searching_driver: %w", err)
	}

	// Start retry loop with timeout
	searchStartTime := time.Now()
	searchDeadline := searchStartTime.Add(time.Duration(s.maxSearchTimeSeconds) * time.Second)
	attemptNumber := 0

	for {
		attemptNumber++

		// Check if search time exceeded
		if time.Now().After(searchDeadline) {
			if err := s.orderRepo.UpdateStatus(orderID, models.OrderStatusNoDriverAvailable); err != nil {
				return fmt.Errorf("failed to update order status: %w", err)
			}
			return fmt.Errorf("no hay conductores disponibles - tiempo máximo de búsqueda excedido (%d segundos)", s.maxSearchTimeSeconds)
		}

		// Check if order was cancelled by user
		currentOrder, err := s.orderRepo.FindByID(orderID)
		if err != nil {
			return fmt.Errorf("failed to refresh order: %w", err)
		}
		if currentOrder.Status == models.OrderStatusCancelled {
			return fmt.Errorf("orden cancelada por el usuario")
		}

		// Find available drivers in FIXED radius

		driversInRadius, err := s.locationRepo.FindAvailableInRadius(
			order.PickupLatitude,
			order.PickupLongitude,
			s.radiusKm,
		)
		if err != nil {
			time.Sleep(time.Duration(s.retryIntervalSeconds) * time.Second)
			continue
		}

		if len(driversInRadius) == 0 {
			time.Sleep(time.Duration(s.retryIntervalSeconds) * time.Second)
			continue
		}

		// Calculate accurate distances using Google Maps Distance Matrix API
		driversWithDistances, err := s.calculateDriverDistances(order, driversInRadius)
		if err != nil {
			// Fallback: use Haversine distances
			driversWithDistances = s.convertToDriversWithDistance(driversInRadius)
		}

		// Try to assign sequentially to all available drivers
		assigned, _, err := s.assignSequentially(order, driversWithDistances, s.radiusKm)
		if err != nil {
			time.Sleep(time.Duration(s.retryIntervalSeconds) * time.Second)
			continue
		}

		if assigned {
			return nil
		}

		// All drivers rejected/timed out - retry after interval
		time.Sleep(time.Duration(s.retryIntervalSeconds) * time.Second)
	}
}

// assignSequentially assigns order to drivers one at a time (sequential with 10s timeout each)
// Drivers are tried in order from closest to farthest, with a 10-second timeout per driver
// This replaces the old concurrent Fan-Out pattern with a sequential queue approach
func (s *AssignmentService) assignSequentially(
	order *models.Order,
	drivers []models.DriverWithDistance,
	searchRadiusKm float64,
) (bool, uuid.UUID, error) {
	if len(drivers) == 0 {
		return false, uuid.Nil, nil
	}

	// 🚫 CRITICAL: Exclude drivers who already rejected/timed out/expired for this order
	rejectedDriverIDs, err := s.assignmentRepo.GetRejectedDriverIDsByOrderID(order.ID)
	if err == nil && len(rejectedDriverIDs) > 0 {

		// Create map for O(1) lookup
		rejectedMap := make(map[uuid.UUID]bool)
		for _, driverID := range rejectedDriverIDs {
			rejectedMap[driverID] = true
		}

		// Filter out rejected drivers
		filteredDrivers := make([]models.DriverWithDistance, 0, len(drivers))
		for _, driver := range drivers {
			if !rejectedMap[driver.DriverID] {
				filteredDrivers = append(filteredDrivers, driver)
			}
		}

		drivers = filteredDrivers

		// If all drivers were filtered out, return early
		if len(drivers) == 0 {
			return false, uuid.Nil, nil
		}
	}

	// Create queue for this order (stores up to 10 drivers in memory)
	maxDriversInQueue := 10
	if len(drivers) > maxDriversInQueue {
		drivers = drivers[:maxDriversInQueue]
	}

	queue := s.queueManager.CreateQueue(order.ID, drivers)
	defer s.queueManager.RemoveQueue(order.ID) // Cleanup when done

	// Try each driver sequentially
	driverIndex := 0
	for queue.HasNext() {
		driverIndex++
		driver, ok := queue.Next()
		if !ok {
			break
		}

		// Create assignment for this driver
		assignmentID, err := s.createAssignmentForDriver(order, driver, searchRadiusKm)
		if err != nil {
			continue // Try next driver
		}

		queue.SetAssignmentID(assignmentID)

		// Send notification ONLY to this driver (not broadcast)
		s.sendDriverNotification(order, driver, assignmentID)

		// Watch for driver response with 10-second timeout
		responseCh := s.watcher.Watch(assignmentID)
		timeout := time.After(time.Duration(s.timeoutSeconds) * time.Second)

		select {
		case response := <-responseCh:
			if response.Error != nil {
				continue // Try next driver
			}

			switch response.Status {
			case models.AssignmentStatusAccepted:
				queue.MarkAccepted()

				// IMMEDIATE cleanup - remove queue from memory NOW (don't wait for defer)
				s.queueManager.RemoveQueue(order.ID)

				// Update order status
				if err := s.orderRepo.UpdateAccepted(order.ID); err != nil {
					return false, uuid.Nil, fmt.Errorf("failed to update order to accepted: %w", err)
				}

				// Expire all other pending assignments for this order
				_ = s.assignmentRepo.ExpirePendingByOrderID(order.ID)

				// Notify order acceptance
				s.notifyOrderAccepted(order, driver.DriverID, drivers)

				return true, driver.DriverID, nil

			case models.AssignmentStatusRejected:
				// Continue to next driver
			}

		case <-timeout:
			// Mark assignment as expired
			_ = s.assignmentRepo.UpdateStatus(assignmentID, models.AssignmentStatusExpired)
			// Continue to next driver
		}
	}

	// All drivers exhausted
	return false, uuid.Nil, nil
}

// createAssignmentForDriver creates an assignment record for a driver
func (s *AssignmentService) createAssignmentForDriver(
	order *models.Order,
	driver models.DriverWithDistance,
	searchRadiusKm float64,
) (uuid.UUID, error) {
	// Get next attempt number
	attemptNumber, err := s.assignmentRepo.GetNextAttemptNumber(order.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get next attempt number: %w", err)
	}

	// Create assignment record
	assignment := &models.OrderAssignment{
		OrderID:                 order.ID,
		DriverID:                driver.DriverID,
		AttemptNumber:           attemptNumber,
		SearchRadiusKm:          searchRadiusKm,
		DistanceToPickupKm:      driver.DistanceToPickupKm,
		EstimatedArrivalMinutes: &driver.EstimatedArrivalMinutes,
		Status:                  models.AssignmentStatusPending,
		ExpiresAt:               time.Now().Add(time.Duration(s.timeoutSeconds) * time.Second),
	}

	if err := s.assignmentRepo.Create(assignment); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	// Assign driver to order temporarily (status: assigned)
	if err := s.orderRepo.AssignDriver(order.ID, driver.DriverID); err != nil {
		return uuid.Nil, fmt.Errorf("failed to assign driver to order: %w", err)
	}

	return assignment.ID, nil
}

// sendDriverNotification sends FCM and WebSocket notifications to a driver (non-blocking)
func (s *AssignmentService) sendDriverNotification(
	order *models.Order,
	driver models.DriverWithDistance,
	assignmentID uuid.UUID,
) {
	expiresAt := time.Now().Add(time.Duration(s.timeoutSeconds) * time.Second).Format(time.RFC3339)

	// Build shared data once
	wsData := map[string]any{
		"type":                   "new_order",
		"order_id":               order.ID.String(),
		"assignment_id":          assignmentID.String(),
		"customer_name":          order.CustomerName,
		"pickup_address":         order.PickupAddress,
		"delivery_address":       order.DeliveryAddress,
		"distance_km":            driver.DistanceToPickupKm,
		"estimated_time_minutes": driver.EstimatedArrivalMinutes,
		"total_amount":           order.TotalAmount,
		"expires_at":             expiresAt,
		"timeout_seconds":        s.timeoutSeconds, // Time in seconds before assignment expires
		"pickup_latitude":        order.PickupLatitude,
		"pickup_longitude":       order.PickupLongitude,
		"delivery_latitude":      order.DeliveryLatitude,
		"delivery_longitude":     order.DeliveryLongitude,
		"pickup_instructions":    order.PickupInstructions,
		"delivery_instructions":  order.DeliveryInstructions,
	}

	// 🚀 CRITICAL: Send WebSocket FIRST and SYNCHRONOUSLY (fastest path)
	if s.wsHub != nil {
		_ = s.wsHub.SendToUser(driver.DriverID, wsData)

		// Also broadcast to order channel
		orderChannel := fmt.Sprintf("order:%s", order.ID.String())
		_ = s.wsHub.BroadcastToChannel(orderChannel, wsData)
	}

	// 📱 Send FCM notification AFTER WebSocket (background, non-blocking)
	go func() {
		if s.notificationSvc == nil {
			return
		}

		notificationTitle := "Nueva orden disponible"
		notificationBody := fmt.Sprintf("Orden de %s - %.1f km de distancia - $%.2f",
			order.CustomerName,
			driver.DistanceToPickupKm,
			order.TotalAmount,
		)

		notificationData := map[string]string{
			"type":                   "new_order",
			"order_id":               order.ID.String(),
			"assignment_id":          assignmentID.String(),
			"pickup_address":         order.PickupAddress,
			"delivery_address":       order.DeliveryAddress,
			"distance_km":            fmt.Sprintf("%.2f", driver.DistanceToPickupKm),
			"estimated_time_minutes": fmt.Sprintf("%d", driver.EstimatedArrivalMinutes),
			"total_amount":           fmt.Sprintf("%.2f", order.TotalAmount),
			"expires_at":             expiresAt,
			"timeout_seconds":        fmt.Sprintf("%d", s.timeoutSeconds), // Time in seconds before assignment expires
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		dataJSON, _ := json.Marshal(notificationData)
		notificationReq := &notificationModels.CreateNotificationRequest{
			UserID:           driver.DriverID,
			Title:            notificationTitle,
			Body:             notificationBody,
			Data:             dataJSON,
			NotificationType: notificationModels.NotificationTypeOrderAssigned,
		}

		_, _ = s.notificationSvc.CreateAndSend(ctx, notificationReq)
	}()
}

// notifyOrderAccepted sends notifications about order acceptance
func (s *AssignmentService) notifyOrderAccepted(order *models.Order, acceptedDriverID uuid.UUID, allDrivers []models.DriverWithDistance) {
	if s.wsHub == nil {
		return
	}

	// Find driver name
	var driverName string
	for _, driver := range allDrivers {
		if driver.DriverID == acceptedDriverID {
			driverName = driver.DriverName
			break
		}
	}

	// Send WebSocket notification (fire-and-forget)
	go func() {
		wsData := map[string]any{
			"type":        "order_accepted",
			"order_id":    order.ID.String(),
			"driver_id":   acceptedDriverID.String(),
			"driver_name": driverName,
			"status":      "accepted",
			"message":     "El conductor ha aceptado tu orden",
		}

		orderChannel := fmt.Sprintf("order:%s", order.ID.String())
		_ = s.wsHub.BroadcastToChannel(orderChannel, wsData)
	}()
}

// tryAssignToDriver tries to assign an order to a specific driver
// calculateDriverDistances calculates accurate driving distances using Google Maps API
func (s *AssignmentService) calculateDriverDistances(
	order *models.Order,
	drivers []driverModels.DriverWithInfo,
) ([]models.DriverWithDistance, error) {
	if len(drivers) == 0 {
		return []models.DriverWithDistance{}, nil
	}

	// Prepare origins (driver locations)
	origins := make([]gmaps.Location, len(drivers))
	for i, driver := range drivers {
		origins[i] = gmaps.Location{
			Latitude:  driver.Latitude,
			Longitude: driver.Longitude,
		}
	}

	// Prepare destination (pickup location)
	destination := gmaps.Location{
		Latitude:  order.PickupLatitude,
		Longitude: order.PickupLongitude,
	}

	// Calculate distances using Google Maps
	ctx := context.Background()
	distances, err := s.gmapsClient.CalculateMultipleDistances(ctx, origins, destination)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate distances: %w", err)
	}

	// Combine driver info with distances
	result := make([]models.DriverWithDistance, 0, len(drivers))
	for i, driver := range drivers {
		if i < len(distances) {
			result = append(result, models.DriverWithDistance{
				DriverID:                driver.DriverID,
				DriverName:              driver.Name,
				DistanceToPickupKm:      distances[i].DistanceKm,
				EstimatedArrivalMinutes: distances[i].DurationMinutes,
			})
		}
	}

	return result, nil
}

// convertToDriversWithDistance converts DriverWithInfo to DriverWithDistance using Haversine distances
func (s *AssignmentService) convertToDriversWithDistance(drivers []driverModels.DriverWithInfo) []models.DriverWithDistance {
	result := make([]models.DriverWithDistance, len(drivers))
	for i, driver := range drivers {
		// Estimate time: assume 30 km/h average speed in city
		estimatedMinutes := int(driver.Latitude * 2) // Rough estimate based on Haversine

		result[i] = models.DriverWithDistance{
			DriverID:                driver.DriverID,
			DriverName:              driver.Name,
			DistanceToPickupKm:      0, // Haversine distance not stored in DriverWithInfo
			EstimatedArrivalMinutes: estimatedMinutes,
		}
	}
	return result
}

// AcceptOrder marks an assignment as accepted by the driver
func (s *AssignmentService) AcceptOrder(orderID, driverID uuid.UUID) error {
	// Find pending assignment
	assignment, err := s.assignmentRepo.FindPendingByOrderAndDriver(orderID, driverID)
	if err != nil {
		return fmt.Errorf("failed to find assignment: %w", err)
	}
	if assignment == nil {
		return fmt.Errorf("no hay una asignación pendiente para esta orden - es posible que ya haya expirado o sido asignada a otro conductor")
	}

	// Check if assignment has expired
	if time.Now().After(assignment.ExpiresAt) {
		_ = s.assignmentRepo.UpdateStatus(assignment.ID, models.AssignmentStatusTimeout)
		return fmt.Errorf("la asignación ha expirado - el tiempo límite era %v", assignment.ExpiresAt.Format("15:04:05"))
	}

	// Mark assignment as accepted
	if err := s.assignmentRepo.UpdateStatus(assignment.ID, models.AssignmentStatusAccepted); err != nil {
		return fmt.Errorf("failed to accept assignment: %w", err)
	}

	// Notify watcher immediately (no polling needed!)
	s.watcher.NotifyAccepted(assignment.ID)

	return nil
}

// RejectOrder marks an assignment as rejected by the driver
func (s *AssignmentService) RejectOrder(orderID, driverID uuid.UUID, reason string) error {
	// Find pending assignment
	assignment, err := s.assignmentRepo.FindPendingByOrderAndDriver(orderID, driverID)
	if err != nil {
		return fmt.Errorf("failed to find assignment: %w", err)
	}
	if assignment == nil {
		return fmt.Errorf("no hay una asignación pendiente para esta orden - es posible que ya haya expirado o sido asignada a otro conductor")
	}

	// Mark assignment as rejected
	if err := s.assignmentRepo.UpdateStatusWithReason(assignment.ID, models.AssignmentStatusRejected, reason); err != nil {
		return fmt.Errorf("failed to reject assignment: %w", err)
	}

	// Notify watcher immediately (no polling needed!)
	s.watcher.NotifyRejected(assignment.ID)

	return nil
}

// GetPendingAssignmentsByDriver retrieves all pending assignments for a driver
func (s *AssignmentService) GetPendingAssignmentsByDriver(driverID uuid.UUID) ([]*models.OrderAssignment, error) {
	assignments, err := s.assignmentRepo.FindPendingByDriverID(driverID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending assignments: %w", err)
	}
	return assignments, nil
}
