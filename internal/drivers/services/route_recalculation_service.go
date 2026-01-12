package services

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"tacoshare-delivery-api/pkg/gmaps"

	"github.com/google/uuid"
)

// RecalculationCache stores the last recalculation time and location for throttling
type RecalculationCache struct {
	LastRecalcTime  time.Time
	LastLat         float64
	LastLng         float64
	LastDistanceKm  float64
	LastDurationMin int
}

// RouteRecalculationService handles intelligent route recalculation with throttling
type RouteRecalculationService struct {
	gmapsClient *gmaps.Client
	cache       map[uuid.UUID]*RecalculationCache // Driver ID -> cache
	mu          sync.RWMutex

	// Configuration thresholds
	minRecalcIntervalSeconds int     // Minimum time between recalculations (default: 30s)
	minDistanceMovedMeters   float64 // Minimum distance moved to trigger recalc (default: 200m)
	minETAChangeMins         int     // Minimum ETA change to update DB/broadcast (default: 2 mins)
}

// NewRouteRecalculationService creates a new route recalculation service
func NewRouteRecalculationService(gmapsClient *gmaps.Client) *RouteRecalculationService {
	return &RouteRecalculationService{
		gmapsClient:              gmapsClient,
		cache:                    make(map[uuid.UUID]*RecalculationCache),
		minRecalcIntervalSeconds: 30,
		minDistanceMovedMeters:   200.0,
		minETAChangeMins:         2,
	}
}

// RecalculationResult contains the result of a route recalculation
type RecalculationResult struct {
	ShouldUpdate       bool // Whether DB/WebSocket should be updated
	NewDistanceKm      float64
	NewDurationMinutes int
	DistanceChange     float64 // Change in km (positive = longer route)
	DurationChange     int     // Change in minutes (positive = slower)
}

// ShouldRecalculate checks if recalculation is needed based on throttling rules
func (s *RouteRecalculationService) ShouldRecalculate(driverID uuid.UUID, currentLat, currentLng float64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cached, exists := s.cache[driverID]
	if !exists {
		// No cache = first time = always recalculate
		return true
	}

	// Check time threshold (minimum 30 seconds between recalcs)
	timeSinceLastRecalc := time.Since(cached.LastRecalcTime).Seconds()
	if timeSinceLastRecalc < float64(s.minRecalcIntervalSeconds) {
		return false
	}

	// Check distance threshold (minimum 200 meters moved)
	distanceMoved := haversineDistance(cached.LastLat, cached.LastLng, currentLat, currentLng)
	return distanceMoved >= s.minDistanceMovedMeters
}

// RecalculateRoute recalculates the route from current position to destination
func (s *RouteRecalculationService) RecalculateRoute(
	ctx context.Context,
	driverID uuid.UUID,
	currentLat, currentLng float64,
	destLat, destLng float64,
) (*RecalculationResult, error) {
	// Check if recalculation is needed
	if !s.ShouldRecalculate(driverID, currentLat, currentLng) {
		// Return cached result without API call
		s.mu.RLock()
		cached := s.cache[driverID]
		s.mu.RUnlock()

		return &RecalculationResult{
			ShouldUpdate:       false,
			NewDistanceKm:      cached.LastDistanceKm,
			NewDurationMinutes: cached.LastDurationMin,
			DistanceChange:     0,
			DurationChange:     0,
		}, nil
	}

	// Call Google Maps API to get new distance/duration
	origin := gmaps.Location{Latitude: currentLat, Longitude: currentLng}
	destination := gmaps.Location{Latitude: destLat, Longitude: destLng}

	result, err := s.gmapsClient.CalculateDistance(ctx, origin, destination)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate distance: %w", err)
	}

	// Get previous cached values
	s.mu.RLock()
	cached, exists := s.cache[driverID]
	s.mu.RUnlock()

	var distanceChange float64
	var durationChange int
	shouldUpdate := true

	if exists {
		distanceChange = result.DistanceKm - cached.LastDistanceKm
		durationChange = result.DurationMinutes - cached.LastDurationMin

		// Only update if ETA changed significantly (>= 2 minutes)
		if absInt(durationChange) < s.minETAChangeMins {
			shouldUpdate = false
		}
	}

	// Update cache
	s.mu.Lock()
	s.cache[driverID] = &RecalculationCache{
		LastRecalcTime:  time.Now(),
		LastLat:         currentLat,
		LastLng:         currentLng,
		LastDistanceKm:  result.DistanceKm,
		LastDurationMin: result.DurationMinutes,
	}
	s.mu.Unlock()

	return &RecalculationResult{
		ShouldUpdate:       shouldUpdate,
		NewDistanceKm:      result.DistanceKm,
		NewDurationMinutes: result.DurationMinutes,
		DistanceChange:     distanceChange,
		DurationChange:     durationChange,
	}, nil
}

// ClearCache removes a driver from the cache (call when order is completed/cancelled)
func (s *RouteRecalculationService) ClearCache(driverID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, driverID)
}

// GetCacheStats returns the current cache size (for monitoring)
func (s *RouteRecalculationService) GetCacheStats() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.cache)
}

// haversineDistance calculates the distance between two coordinates in meters
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusMeters = 6371000.0

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}

// absInt returns the absolute value of an integer
func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
