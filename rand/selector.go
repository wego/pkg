package rand

import (
	"math"
	_ "unsafe" // Required for go:linkname
)

// Selector is a probabilistic selector
// It provides a way to randomly select an item with a given probability
// It uses a thread-local random number generator to avoid contention
// and improve performance in concurrent scenarios
type Selector interface {
	// Next returns true with the configured probability
	Next() bool
}

// selector uses the Go runtime fast RNG directly
type selector struct {
	threshold uint64    // Pre-scaled threshold
	_         [7]uint64 // Padding to prevent false sharing
}

//go:linkname runtimeRand runtime.rand
func runtimeRand() uint64

// NewSelector creates a new Selector with the given percentage
func NewSelector(percentage float64) Selector {
	rs := &selector{}
	rs.setThreshold(percentage)
	return rs
}

func (rs *selector) setThreshold(p float64) {
	p = math.Max(0.0, math.Min(100.0, p))
	switch {
	case p <= 0.0:
		rs.threshold = 0
	case p >= 100.0:
		rs.threshold = math.MaxUint64
	default:
		// Maintain 53-bit precision using two-stage scaling
		scaled := p * (1 << 53) / 100.0
		rs.threshold = uint64(scaled * (1 << 11))
	}
}

// Next returns true with the configured probability
//
//go:inline
func (rs *selector) Next() bool {
	switch rs.threshold {
	case 0:
		return false
	case math.MaxUint64:
		return true
	default:
		return runtimeRand() < rs.threshold
	}
}
