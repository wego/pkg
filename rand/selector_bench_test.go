package rand

import (
	"testing"
	"unsafe"
)

// BenchmarkNext focuses only on the Next function for more precise measurement
func BenchmarkNext(b *testing.B) {
	selector := NewSelector(50)
	b.ResetTimer()

	// Avoid compiler optimizations by storing result
	var result uint64

	for i := 0; i < b.N; i++ {
		if selector.Next() {
			result++
		}
	}

	// Force compiler to keep the result
	b.SetBytes(int64(unsafe.Sizeof(result)))
}

// BenchmarkNext_MultipleSelectors tests using multiple selectors with different percentages
func BenchmarkNext_MultipleSelectors(b *testing.B) {
	s25 := NewSelector(25)
	s50 := NewSelector(50)
	s75 := NewSelector(75)
	b.ResetTimer()

	var result uint64

	for i := 0; i < b.N; i++ {
		// Simulate a more complex usage pattern
		switch i % 3 {
		case 0:
			if s25.Next() {
				result++
			}
		case 1:
			if s50.Next() {
				result++
			}
		case 2:
			if s75.Next() {
				result++
			}
		}
	}

	b.SetBytes(int64(unsafe.Sizeof(result)))
}

// BenchmarkNextParallel measures performance with concurrent access
func BenchmarkNextParallel(b *testing.B) {
	selector := NewSelector(50)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine has its own counter to avoid atomic operations on the counter itself
		var localResult uint64

		for pb.Next() {
			if selector.Next() {
				localResult++
			}
		}

		// We don't need to aggregate results, this is just to prevent
		// compiler optimizations from eliminating the loop
		_ = localResult
	})
}

// BenchmarkNextParallel_ThreadCount varies the number of goroutines accessing the selector
func BenchmarkNextParallel_ThreadCount(b *testing.B) {
	threadCounts := []int{2, 4, 8, 16, 32, 64}

	for _, threads := range threadCounts {
		b.Run("Threads_"+string(rune('0'+threads)), func(b *testing.B) {
			selector := NewSelector(50)
			b.ResetTimer()

			b.SetParallelism(threads)
			b.RunParallel(func(pb *testing.PB) {
				var localResult uint64

				for pb.Next() {
					if selector.Next() {
						localResult++
					}
				}

				_ = localResult
			})
		})
	}
}
