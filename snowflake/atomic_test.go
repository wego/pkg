package snowflake

import (
	"testing"
)

func TestAtomicResolver(t *testing.T) {
	id, _ := AtomicResolver(1)

	if id != 0 {
		t.Error("Sequence should be equal 0")
	}
}

func BenchmarkCombinationParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = AtomicResolver(1)
		}
	})
}

func BenchmarkAtomicResolver(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = AtomicResolver(1)
	}
}
