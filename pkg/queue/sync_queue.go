package queue

import (
	"sync"
)

// SyncQueue is a thread-safe fixed-length queue that maintains a circular buffer of items
type SyncQueue struct {
	mu      sync.RWMutex
	notFull *sync.Cond
	queue   *FixedQueue
}

// NewSyncQueue creates a new thread-safe fixed-length queue with the specified capacity
func NewSyncQueue(cap int) *SyncQueue {
	q := &SyncQueue{
		queue: NewFixedQueue(cap),
	}
	q.notFull = sync.NewCond(&q.mu)
	return q
}

// Push adds a new item to the queue. If the queue is full, it will wait until space is available.
func (q *SyncQueue) Push(item interface{}) {
	q.mu.Lock()

	// Wait while the queue is full
	for q.queue.Len() >= q.queue.Cap() {
		q.notFull.Wait()
	}

	q.queue.Push(item)
	q.mu.Unlock()
}

// TryPush attempts to add an item to the queue without waiting.
// Returns true if successful, false if the queue is full.
func (q *SyncQueue) TryPush(item interface{}) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.queue.Len() >= q.queue.Cap() {
		return false
	}
	q.queue.Push(item)
	return true
}

// Pop removes and returns the oldest item in the queue.
// Returns nil if the queue is empty.
func (q *SyncQueue) Pop() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	item := q.queue.Pop()
	if item != nil {
		// Signal that there's now space in the queue
		q.notFull.Signal()
	}
	return item
}

// Items returns a slice of all items in chronological order (oldest to newest)
func (q *SyncQueue) Items() []interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.queue.Items()
}

// ReversedItems returns a slice of all items in reverse chronological order (newest to oldest)
func (q *SyncQueue) ReversedItems() []interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.queue.ReversedItems()
}

// Clear removes all items from the queue
func (q *SyncQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue.Clear()
	// Signal that there's now space in the queue
	q.notFull.Broadcast()
}

// Len returns the current number of items in the queue
func (q *SyncQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.queue.Len()
}

// Cap returns the maximum capacity of the queue
func (q *SyncQueue) Cap() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.queue.Cap()
}
