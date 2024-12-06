package queue

import (
	"sync"
	"testing"
	"time"
)

func TestSyncQueue(t *testing.T) {
	t.Run("concurrent push and pop", func(t *testing.T) {
		q := NewSyncQueue(5)
		var wg sync.WaitGroup
		results := make(chan interface{}, 10) // Buffer for all possible results

		// Start poppers first
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if item := q.Pop(); item != nil {
					results <- item
				}
			}()
		}

		// Then start pushers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()
				q.TryPush(val) // Use TryPush to avoid blocking
			}(i)
		}

		wg.Wait()
		close(results)

		// Count received items
		count := 0
		for range results {
			count++
		}

		if count > 5 {
			t.Errorf("received %d items, want <= 5", count)
		}
	})

	t.Run("push waiting behavior", func(t *testing.T) {
		q := NewSyncQueue(2)
		var wg sync.WaitGroup

		// Fill the queue
		q.Push(1)
		q.Push(2)

		// Try to push to full queue
		pushed := make(chan bool, 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			q.Push(3) // This should wait
			duration := time.Since(start)
			if duration < 100*time.Millisecond {
				t.Errorf("Push didn't wait long enough: %v", duration)
			}
			pushed <- true
		}()

		// Wait a bit and then pop
		time.Sleep(100 * time.Millisecond)
		q.Pop() // This should unblock the push

		select {
		case <-pushed:
			// Push completed successfully
		case <-time.After(500 * time.Millisecond):
			t.Error("Push didn't complete after Pop")
		}

		wg.Wait()
		if q.Len() != 2 {
			t.Errorf("queue length = %d, want 2", q.Len())
		}
	})

	t.Run("try push behavior", func(t *testing.T) {
		q := NewSyncQueue(2)

		// First two pushes should succeed
		if !q.TryPush(1) {
			t.Error("First TryPush failed")
		}
		if !q.TryPush(2) {
			t.Error("Second TryPush failed")
		}

		// Third push should fail
		if q.TryPush(3) {
			t.Error("Third TryPush succeeded when queue was full")
		}

		// After a pop, push should succeed again
		q.Pop()
		if !q.TryPush(4) {
			t.Error("TryPush failed after Pop")
		}
	})

	t.Run("concurrent read operations", func(t *testing.T) {
		q := NewSyncQueue(3)
		q.Push(1)
		q.Push(2)
		q.Push(3)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				items := q.Items()
				if len(items) != 3 {
					t.Errorf("Items() length = %d, want 3", len(items))
				}
				reversed := q.ReversedItems()
				if len(reversed) != 3 {
					t.Errorf("ReversedItems() length = %d, want 3", len(reversed))
				}
				l := q.Len()
				if l != 3 {
					t.Errorf("Len() = %d, want 3", l)
				}
				c := q.Cap()
				if c != 3 {
					t.Errorf("Cap() = %d, want 3", c)
				}
			}()
		}
		wg.Wait()
	})

	t.Run("concurrent clear", func(t *testing.T) {
		q := NewSyncQueue(5)
		var wg sync.WaitGroup

		// Start concurrent operations
		for i := 0; i < 10; i++ {
			wg.Add(3)
			// Pusher
			go func() {
				defer wg.Done()
				q.TryPush(1) // Use TryPush to avoid blocking
			}()
			// Popper
			go func() {
				defer wg.Done()
				q.Pop()
			}()
			// Clearer
			go func() {
				defer wg.Done()
				q.Clear()
			}()
		}
		wg.Wait()

		// Queue should be in a valid state after concurrent operations
		if q.Len() < 0 || q.Len() > q.Cap() {
			t.Errorf("invalid queue length after concurrent operations: len=%d, cap=%d", q.Len(), q.Cap())
		}
	})
}
