package queue

import (
	"reflect"
	"testing"
)

func TestFixedLengthVec(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		items    []interface{}
		wantOld  []interface{}
		wantNew  []interface{}
	}{
		{
			name:     "basic push and retrieve",
			capacity: 3,
			items:    []interface{}{1, 2, 3},
			wantOld:  []interface{}{1, 2, 3},
			wantNew:  []interface{}{3, 2, 1},
		},
		{
			name:     "overflow handling",
			capacity: 3,
			items:    []interface{}{1, 2, 3, 4, 5},
			wantOld:  []interface{}{3, 4, 5},
			wantNew:  []interface{}{5, 4, 3},
		},
		{
			name:     "mixed types",
			capacity: 4,
			items:    []interface{}{1, "two", 3.0, true},
			wantOld:  []interface{}{1, "two", 3.0, true},
			wantNew:  []interface{}{true, 3.0, "two", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewFixedQueue(tt.capacity)

			// Test initial state
			if v.Len() != 0 {
				t.Errorf("initial length = %v, want 0", v.Len())
			}
			if v.Cap() != tt.capacity {
				t.Errorf("capacity = %v, want %v", v.Cap(), tt.capacity)
			}

			// Push items
			for _, item := range tt.items {
				v.Push(item)
			}

			// Check length
			if got := v.Len(); got != min(len(tt.items), tt.capacity) {
				t.Errorf("Len() = %v, want %v", got, min(len(tt.items), tt.capacity))
			}

			// Check items old to new
			gotOld := v.Items()
			if !reflect.DeepEqual(gotOld, tt.wantOld) {
				t.Errorf("ItemsOldToNew() = %v, want %v", gotOld, tt.wantOld)
			}

			// Check items new to old
			gotNew := v.ReversedItems()
			if !reflect.DeepEqual(gotNew, tt.wantNew) {
				t.Errorf("ReversedItems() = %v, want %v", gotNew, tt.wantNew)
			}

			// Check deprecated Items() matches ItemsOldToNew()
			gotDeprecated := v.Items()
			if !reflect.DeepEqual(gotDeprecated, gotOld) {
				t.Errorf("Items() = %v, want %v", gotDeprecated, gotOld)
			}

			// Test Clear
			v.Clear()
			if v.Len() != 0 {
				t.Errorf("after Clear(), length = %v, want 0", v.Len())
			}
			if items := v.Items(); items != nil {
				t.Errorf("after Clear(), ItemsOldToNew() = %v, want nil", items)
			}
		})
	}
}

func TestPop(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		items    []interface{}
		pops     int
		want     []interface{}
	}{
		{
			name:     "pop from empty",
			capacity: 3,
			items:    []interface{}{},
			pops:     1,
			want:     []interface{}{nil},
		},
		{
			name:     "pop single item",
			capacity: 3,
			items:    []interface{}{1},
			pops:     1,
			want:     []interface{}{1},
		},
		{
			name:     "pop multiple items",
			capacity: 3,
			items:    []interface{}{1, 2, 3},
			pops:     2,
			want:     []interface{}{1, 2},
		},
		{
			name:     "pop with overflow",
			capacity: 3,
			items:    []interface{}{1, 2, 3, 4, 5},
			pops:     2,
			want:     []interface{}{3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewFixedQueue(tt.capacity)

			// Push items
			for _, item := range tt.items {
				v.Push(item)
			}

			// Pop items and verify
			for i := 0; i < tt.pops; i++ {
				got := v.Pop()
				if got != tt.want[i] {
					t.Errorf("Pop() = %v, want %v", got, tt.want[i])
				}
			}

			// Verify remaining length
			expectedLen := max(0, min(len(tt.items), tt.capacity)-tt.pops)
			if got := v.Len(); got != expectedLen {
				t.Errorf("after Pop(), Len() = %v, want %v", got, expectedLen)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
