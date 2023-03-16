package collection_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/collection"
)

func Test_Chunk(t *testing.T) {
	for _, test := range []struct {
		name     string
		size     int
		slice    []int
		expected [][]int
	}{
		{
			name:     "nil slice",
			size:     2,
			slice:    nil,
			expected: nil,
		},
		{
			name:     "negative chunk size",
			size:     -1,
			slice:    []int{},
			expected: nil,
		},
		{
			name:     "zero chunk size",
			size:     0,
			slice:    []int{},
			expected: nil,
		},
		{
			name:     "empty slice",
			size:     2,
			slice:    []int{},
			expected: nil,
		},
		{
			name:     "int slice has less items than chunk size",
			size:     2,
			slice:    []int{1},
			expected: [][]int{{1}},
		},
		{
			name:     "int slice has number of items same as chunk size",
			size:     2,
			slice:    []int{1, 2},
			expected: [][]int{{1, 2}},
		},
		{
			name:     "int slice has more items than chunk size",
			size:     2,
			slice:    []int{1, 2, 3, 4, 5},
			expected: [][]int{{1, 2}, {3, 4}, {5}},
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			assert.Equal(tt, test.expected, collection.Chunk(test.slice, test.size))
		})
	}

	for _, test := range []struct {
		name     string
		size     int
		slice    []string
		expected [][]string
	}{
		{
			name:     "string slice has less items than chunk size",
			size:     2,
			slice:    []string{"1"},
			expected: [][]string{{"1"}},
		},
		{
			name:     "string slice has number of items same as chunk size",
			size:     2,
			slice:    []string{"1", "2"},
			expected: [][]string{{"1", "2"}},
		},
		{
			name:     "string slice has more items than chunk size",
			size:     2,
			slice:    []string{"1", "2", "3", "4", "5"},
			expected: [][]string{{"1", "2"}, {"3", "4"}, {"5"}},
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			assert.Equal(tt, test.expected, collection.Chunk(test.slice, test.size))
		})
	}

	for _, test := range []struct {
		name     string
		size     int
		slice    []float64
		expected [][]float64
	}{
		{
			name:     "float64 slice has less items than chunk size",
			size:     2,
			slice:    []float64{1.0},
			expected: [][]float64{{1.0}},
		},
		{
			name:     "float64 slice has number of items same as chunk size",
			size:     2,
			slice:    []float64{1.0, 2.0},
			expected: [][]float64{{1.0, 2.0}},
		},
		{
			name:     "float64 slice has more items than chunk size",
			size:     2,
			slice:    []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: [][]float64{{1.0, 2.0}, {3.0, 4.0}, {5.0}},
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			assert.Equal(tt, test.expected, collection.Chunk(test.slice, test.size))
		})
	}

	type testStruct struct {
		s string
		i int
		f float64
	}

	for _, test := range []struct {
		name     string
		size     int
		slice    []testStruct
		expected [][]testStruct
	}{
		{
			name:     "struct slice has less items than chunk size",
			size:     2,
			slice:    []testStruct{{"1", 1, 1.0}},
			expected: [][]testStruct{{{"1", 1, 1.0}}},
		},
		{
			name:     "struct slice has number of items same as chunk size",
			size:     2,
			slice:    []testStruct{{"1", 1, 1.0}, {"2", 2, 2.0}},
			expected: [][]testStruct{{{"1", 1, 1.0}, {"2", 2, 2.0}}},
		},
		{
			name:     "struct slice has more items than chunk size",
			size:     2,
			slice:    []testStruct{{"1", 1, 1.0}, {"2", 2, 2.0}, {"3", 3, 3.0}, {"4", 4, 4.0}, {"5", 5, 5.0}},
			expected: [][]testStruct{{{"1", 1, 1.0}, {"2", 2, 2.0}}, {{"3", 3, 3.0}, {"4", 4, 4.0}}, {{"5", 5, 5.0}}},
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			assert.Equal(tt, test.expected, collection.Chunk(test.slice, test.size))
		})
	}
}
