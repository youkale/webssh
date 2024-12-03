package tui

import (
	"testing"
)

func TestNewFixedLengthVec(t *testing.T) {
	maxLen := 32
	vec := newFixedLengthVec(maxLen)
	if vec.size != 0 {
		t.Errorf("Expected initial size to be 0, got %d", vec.size)
	}
	if vec.head != 0 {
		t.Errorf("Expected initial head to be 0, got %d", vec.head)
	}
	if vec.tail != 0 {
		t.Errorf("Expected initial tail to be 0, got %d", vec.tail)
	}
}

func TestPushAndPop(t *testing.T) {
	maxLen := 32
	vec := newFixedLengthVec(maxLen)

	// Test pushing items
	vec.Push("first")
	if vec.size != 1 {
		t.Errorf("Expected size 1 after push, got %d", vec.size)
	}

	vec.Push("second")
	if vec.size != 2 {
		t.Errorf("Expected size 2 after push, got %d", vec.size)
	}

	// Test popping items
	item, ok := vec.Pop()
	if !ok {
		t.Error("Expected Pop to return true")
	}
	if item != "first" {
		t.Errorf("Expected 'first', got %v", item)
	}
	if vec.size != 1 {
		t.Errorf("Expected size 1 after pop, got %d", vec.size)
	}

	item, ok = vec.Pop()
	if !ok {
		t.Error("Expected Pop to return true")
	}
	if item != "second" {
		t.Errorf("Expected 'second', got %v", item)
	}

	// Test popping from empty vector
	_, ok = vec.Pop()
	if ok {
		t.Error("Expected Pop to return false for empty vector")
	}
}

func TestMaxLength(t *testing.T) {
	maxLen := 32

	vec := newFixedLengthVec(maxLen)

	// Fill vector to max capacity
	for i := 0; i < maxLen+5; i++ {
		vec.Push(i)
	}

	if vec.Len() != maxLen {
		t.Errorf("Expected length %d, got %d", maxLen, vec.Len())
	}

	// Check that only the last maxLen items are retained
	items := vec.Items()
	if len(items) != maxLen {
		t.Errorf("Expected items length %d, got %d", maxLen, len(items))
	}

	// Verify the items are in correct order
	for i := 0; i < maxLen; i++ {
		expected := i + 5 // Since we pushed maxLen+5 items
		if items[i] != expected {
			t.Errorf("Expected item at index %d to be %d, got %v", i, expected, items[i])
		}
	}
}

func TestItems(t *testing.T) {
	maxLen := 32
	vec := newFixedLengthVec(maxLen)

	// Test empty vector
	items := vec.Items()
	if items != nil {
		t.Error("Expected nil items for empty vector")
	}

	// Test with some items
	testItems := []string{"one", "two", "three"}
	for _, item := range testItems {
		vec.Push(item)
	}

	items = vec.Items()
	if len(items) != len(testItems) {
		t.Errorf("Expected %d items, got %d", len(testItems), len(items))
	}

	for i, item := range items {
		if item != testItems[i] {
			t.Errorf("Expected item %v at index %d, got %v", testItems[i], i, item)
		}
	}
}
