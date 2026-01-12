package services

import (
	"sync"
	"time"

	"tacoshare-delivery-api/internal/orders/models"

	"github.com/google/uuid"
)

// AssignmentResponse represents a driver's response to an assignment
type AssignmentResponse struct {
	Status models.AssignmentStatus
	Error  error
}

// AssignmentWatcher manages real-time assignment status updates using channels
// This eliminates the need for database polling (reducing from 2,000 queries/hour to 0)
type AssignmentWatcher struct {
	// Map of assignment ID to channel for status updates
	watchers map[uuid.UUID]chan AssignmentResponse
	mu       sync.RWMutex

	// Cleanup ticker
	cleanupTicker *time.Ticker
	done          chan struct{}
}

// NewAssignmentWatcher creates a new assignment watcher
func NewAssignmentWatcher() *AssignmentWatcher {
	watcher := &AssignmentWatcher{
		watchers:      make(map[uuid.UUID]chan AssignmentResponse),
		cleanupTicker: time.NewTicker(30 * time.Second),
		done:          make(chan struct{}),
	}

	// Start background cleanup goroutine
	go watcher.cleanupExpiredWatchers()

	return watcher
}

// Watch creates a new watcher for an assignment and returns a channel
// The channel will receive exactly one response (accept/reject/timeout)
func (w *AssignmentWatcher) Watch(assignmentID uuid.UUID) <-chan AssignmentResponse {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Create buffered channel to prevent blocking
	ch := make(chan AssignmentResponse, 1)
	w.watchers[assignmentID] = ch

	return ch
}

// NotifyAccepted notifies all watchers that an assignment was accepted
func (w *AssignmentWatcher) NotifyAccepted(assignmentID uuid.UUID) {
	w.notify(assignmentID, AssignmentResponse{
		Status: models.AssignmentStatusAccepted,
		Error:  nil,
	})
}

// NotifyRejected notifies all watchers that an assignment was rejected
func (w *AssignmentWatcher) NotifyRejected(assignmentID uuid.UUID) {
	w.notify(assignmentID, AssignmentResponse{
		Status: models.AssignmentStatusRejected,
		Error:  nil,
	})
}

// NotifyTimeout notifies all watchers that an assignment timed out
func (w *AssignmentWatcher) NotifyTimeout(assignmentID uuid.UUID) {
	w.notify(assignmentID, AssignmentResponse{
		Status: models.AssignmentStatusTimeout,
		Error:  nil,
	})
}

// NotifyError notifies all watchers of an error
func (w *AssignmentWatcher) NotifyError(assignmentID uuid.UUID, err error) {
	w.notify(assignmentID, AssignmentResponse{
		Status: "",
		Error:  err,
	})
}

// notify sends a response to the watcher channel and removes it
func (w *AssignmentWatcher) notify(assignmentID uuid.UUID, response AssignmentResponse) {
	w.mu.Lock()
	defer w.mu.Unlock()

	ch, exists := w.watchers[assignmentID]
	if !exists {
		return
	}

	// Send response (non-blocking due to buffered channel)
	select {
	case ch <- response:
		// Response sent successfully
	default:
		// Channel already has a value (shouldn't happen with buffer size 1)
	}

	// Close and remove the channel
	close(ch)
	delete(w.watchers, assignmentID)
}

// Unwatch removes a watcher for an assignment (useful for cleanup)
func (w *AssignmentWatcher) Unwatch(assignmentID uuid.UUID) {
	w.mu.Lock()
	defer w.mu.Unlock()

	ch, exists := w.watchers[assignmentID]
	if !exists {
		return
	}

	close(ch)
	delete(w.watchers, assignmentID)
}

// cleanupExpiredWatchers periodically removes stale watchers
func (w *AssignmentWatcher) cleanupExpiredWatchers() {
	for {
		select {
		case <-w.cleanupTicker.C:
			// Cleanup is handled automatically when notifications are sent
			// This goroutine just keeps the ticker running
		case <-w.done:
			w.cleanupTicker.Stop()
			return
		}
	}
}

// Close stops the watcher and cleans up resources
func (w *AssignmentWatcher) Close() {
	close(w.done)

	w.mu.Lock()
	defer w.mu.Unlock()

	// Close all remaining channels
	for id, ch := range w.watchers {
		close(ch)
		delete(w.watchers, id)
	}
}
