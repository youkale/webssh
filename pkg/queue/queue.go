package queue

// FixedQueue is a fixed-length vector that maintains a circular buffer of items
type FixedQueue struct {
	items []interface{}
	head  int
	size  int
	cap   int
}

// NewFixedQueue creates a new fixed-length vector with the specified capacity
func NewFixedQueue(cap int) *FixedQueue {
	if cap <= 0 {
		cap = 1
	}
	return &FixedQueue{
		items: make([]interface{}, cap),
		cap:   cap,
	}
}

// Push adds a new item to the vector, overwriting the oldest item if at capacity
func (v *FixedQueue) Push(item interface{}) {
	v.items[v.head] = item
	v.head = (v.head + 1) % v.cap
	if v.size < v.cap {
		v.size++
	}
}

// Pop removes and returns the oldest item in the vector.
// Returns nil if the vector is empty.
func (v *FixedQueue) Pop() interface{} {
	if v.size == 0 {
		return nil
	}

	// Calculate the index of the oldest item
	oldestIdx := (v.head - v.size + v.cap) % v.cap
	oldestItem := v.items[oldestIdx]

	// Clear the item from the vector
	v.items[oldestIdx] = nil
	v.size--

	return oldestItem
}

// Items returns a slice of all items in chronological order (oldest to newest)
func (v *FixedQueue) Items() []interface{} {
	if v.size == 0 {
		return nil
	}

	result := make([]interface{}, v.size)
	start := (v.head - v.size + v.cap) % v.cap

	// Copy items in chronological order (oldest to newest)
	for i := 0; i < v.size; i++ {
		idx := (start + i) % v.cap
		result[i] = v.items[idx]
	}

	return result
}

// ReversedItems returns a slice of all items in reverse chronological order (newest to oldest)
func (v *FixedQueue) ReversedItems() []interface{} {
	if v.size == 0 {
		return nil
	}

	result := make([]interface{}, v.size)

	// Copy items in reverse chronological order (newest to oldest)
	for i := 0; i < v.size; i++ {
		idx := (v.head - 1 - i + v.cap) % v.cap
		result[i] = v.items[idx]
	}

	return result
}

// Clear removes all items from the vector
func (v *FixedQueue) Clear() {
	v.items = make([]interface{}, v.cap)
	v.head = 0
	v.size = 0
}

// Len returns the current number of items in the vector
func (v *FixedQueue) Len() int {
	return v.size
}

// Cap returns the maximum capacity of the vector
func (v *FixedQueue) Cap() int {
	return v.cap
}
