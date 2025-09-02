package collection_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/collection"
)

func TestIntersect(t *testing.T) {
	tests := []struct {
		name     string
		slice1   []string
		slice2   []string
		expected []string
	}{
		{
			name:     "both slices have common elements",
			slice1:   []string{"a", "b", "c", "d"},
			slice2:   []string{"c", "d", "e", "f"},
			expected: []string{"c", "d"},
		},
		{
			name:     "no common elements",
			slice1:   []string{"a", "b"},
			slice2:   []string{"c", "d"},
			expected: []string{},
		},
		{
			name:     "first slice is empty",
			slice1:   []string{},
			slice2:   []string{"a", "b"},
			expected: []string{},
		},
		{
			name:     "second slice is empty",
			slice1:   []string{"a", "b"},
			slice2:   []string{},
			expected: []string{},
		},
		{
			name:     "both slices are empty",
			slice1:   []string{},
			slice2:   []string{},
			expected: []string{},
		},
		{
			name:     "identical slices",
			slice1:   []string{"a", "b", "c"},
			slice2:   []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "duplicates in first slice",
			slice1:   []string{"a", "b", "a", "c"},
			slice2:   []string{"a", "c"},
			expected: []string{"a", "c"},
		},
		{
			name:     "duplicates in second slice",
			slice1:   []string{"a", "c"},
			slice2:   []string{"a", "b", "a", "c"},
			expected: []string{"a", "c"},
		},
		{
			name:     "single element intersection",
			slice1:   []string{"a"},
			slice2:   []string{"a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collection.Intersect(tt.slice1, tt.slice2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntersectWithInts(t *testing.T) {
	tests := []struct {
		name     string
		slice1   []int
		slice2   []int
		expected []int
	}{
		{
			name:     "common integers",
			slice1:   []int{1, 2, 3, 4},
			slice2:   []int{3, 4, 5, 6},
			expected: []int{3, 4},
		},
		{
			name:     "no common integers",
			slice1:   []int{1, 2},
			slice2:   []int{3, 4},
			expected: []int{},
		},
		{
			name:     "single integer intersection",
			slice1:   []int{1},
			slice2:   []int{1},
			expected: []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collection.Intersect(tt.slice1, tt.slice2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntersectOrderPreservation(t *testing.T) {
	slice1 := []string{"z", "a", "y", "b"}
	slice2 := []string{"a", "b", "c", "d"}
	expected := []string{"a", "b"}

	result := collection.Intersect(slice1, slice2)
	assert.Equal(t, expected, result)

	// Test reverse order
	result2 := collection.Intersect(slice2, slice1)
	assert.Equal(t, expected, result2)
}

func BenchmarkIntersect(b *testing.B) {
	slice1 := make([]int, 1000)
	slice2 := make([]int, 500)

	for i := 0; i < 1000; i++ {
		slice1[i] = i
	}
	for i := 0; i < 500; i++ {
		slice2[i] = i + 250 // Some overlap
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collection.Intersect(slice1, slice2)
	}
}

func BenchmarkIntersectSmallSlices(b *testing.B) {
	slice1 := []string{"USD", "EUR", "GBP", "JPY", "CAD"}
	slice2 := []string{"EUR", "JPY", "AUD"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collection.Intersect(slice1, slice2)
	}
}
