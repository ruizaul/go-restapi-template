package services

import (
	"sync"
	"time"

	"tacoshare-delivery-api/internal/orders/models"

	"github.com/google/uuid"
)

// DriverQueue manages a sequential queue of drivers for order assignment
// Drivers are tried one at a time with a timeout, optimizing for closest first
type DriverQueue struct {
	drivers         []models.DriverWithDistance
	currentIndex    int
	mu              sync.RWMutex
	assignmentID    uuid.UUID
	currentDriverID uuid.UUID
	status          QueueStatus
}

// QueueStatus represents the current state of the queue
type QueueStatus string

const (
	QueueStatusIdle      QueueStatus = "idle"
	QueueStatusWaiting   QueueStatus = "waiting"
	QueueStatusAccepted  QueueStatus = "accepted"
	QueueStatusExhausted QueueStatus = "exhausted"
)

// NewDriverQueue creates a new driver queue with sorted drivers (closest first)
func NewDriverQueue(drivers []models.DriverWithDistance) *DriverQueue {
	return &DriverQueue{
		drivers:      drivers,
		currentIndex: 0,
		status:       QueueStatusIdle,
	}
}

// HasNext returns true if there are more drivers in the queue
func (q *DriverQueue) HasNext() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.currentIndex < len(q.drivers)
}

// Next returns the next driver in the queue and advances the index
func (q *DriverQueue) Next() (models.DriverWithDistance, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.currentIndex >= len(q.drivers) {
		q.status = QueueStatusExhausted
		return models.DriverWithDistance{}, false
	}

	driver := q.drivers[q.currentIndex]
	q.currentDriverID = driver.DriverID
	q.currentIndex++
	q.status = QueueStatusWaiting

	return driver, true
}

// CurrentDriver returns the current driver being tried
func (q *DriverQueue) CurrentDriver() (uuid.UUID, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.currentDriverID, q.currentDriverID != uuid.Nil
}

// SetAssignmentID sets the assignment ID for the current driver
func (q *DriverQueue) SetAssignmentID(assignmentID uuid.UUID) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.assignmentID = assignmentID
}

// GetAssignmentID returns the current assignment ID
func (q *DriverQueue) GetAssignmentID() uuid.UUID {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.assignmentID
}

// MarkAccepted marks the queue as successfully assigned
func (q *DriverQueue) MarkAccepted() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.status = QueueStatusAccepted
}

// GetStatus returns the current queue status
func (q *DriverQueue) GetStatus() QueueStatus {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.status
}

// RemainingCount returns the number of drivers left in the queue
func (q *DriverQueue) RemainingCount() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.drivers) - q.currentIndex
}

// QueueManager manages multiple order queues in memory
// This is thread-safe and can handle multiple orders simultaneously
type QueueManager struct {
	queues map[uuid.UUID]*DriverQueue
	mu     sync.RWMutex
}

// NewQueueManager creates a new queue manager
func NewQueueManager() *QueueManager {
	return &QueueManager{
		queues: make(map[uuid.UUID]*DriverQueue),
	}
}

// CreateQueue creates a new queue for an order
func (qm *QueueManager) CreateQueue(orderID uuid.UUID, drivers []models.DriverWithDistance) *DriverQueue {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	queue := NewDriverQueue(drivers)
	qm.queues[orderID] = queue
	return queue
}

// GetQueue retrieves a queue for an order
func (qm *QueueManager) GetQueue(orderID uuid.UUID) (*DriverQueue, bool) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	queue, exists := qm.queues[orderID]
	return queue, exists
}

// RemoveQueue removes a queue from memory (cleanup after assignment completes)
func (qm *QueueManager) RemoveQueue(orderID uuid.UUID) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	delete(qm.queues, orderID)
}

// CleanupStaleQueues removes queues older than the specified duration
func (qm *QueueManager) CleanupStaleQueues(maxAge time.Duration) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	// In a production system, you'd track queue creation time
	// For now, this is a placeholder for future implementation
}
