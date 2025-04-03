package rand

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	"sync"
	"time"
)

// Selector is a probabilistic selector
// It provides a way to randomly select an item with a given probability
// It uses a thread-local random number generator to avoid contention
// and improve performance in concurrent scenarios
type Selector interface {
	// Next returns true with the configured probability
	Next() bool
}

// selector provides probabilistic selection with a fixed percentage
type selector struct {
	threshold uint64    // Pre-scaled threshold (cache line aligned)
	_         [7]uint64 // Padding to prevent false sharing
	rngPool   sync.Pool // Thread-local RNG states
}

// Constants for the SplitMix64 algorithm
const (
	phi           = 0x9E3779B97F4A7C15 // Golden ratio constant
	splitmixMult1 = 0xBF58476D1CE4E5B9 // Multiplication constant 1
	splitmixMult2 = 0x94D049BB133111EB // Multiplication constant 2
)

func NewSelector(percentage float64) Selector {
	rs := &selector{}
	rs.setThreshold(percentage)

	// Initialize thread-local RNG pool
	rs.rngPool = sync.Pool{
		New: func() interface{} {
			return newRNGState()
		},
	}

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

// newRNGState creates a new random state for the PRNG
func newRNGState() *uint64 {
	var seed [8]byte
	state := new(uint64)
	if _, err := rand.Read(seed[:]); err != nil {
		*state = uint64(time.Now().UnixNano())
	} else {
		*state = binary.LittleEndian.Uint64(seed[:])
	}
	return state
}

// nextRNG generates the next random number using the SplitMix64 algorithm
//
//go:inline
func nextRNG(state *uint64) uint64 {
	// Update state using SplitMix64 algorithm
	*state += phi
	z := *state
	z ^= z >> 30
	z *= splitmixMult1
	z *= splitmixMult2
	z ^= z >> 31
	return z
}

// Next returns true with the configured probability
//
//go:inline
func (rs *selector) Next() bool {
	// Handle edge cases with a switch for better readability
	switch rs.threshold {
	case 0:
		return false
	case math.MaxUint64:
		return true
	default:
		// Get thread-local RNG state
		state := rs.rngPool.Get().(*uint64)
		defer rs.rngPool.Put(state)

		// Generate random number and compare with threshold
		return nextRNG(state) < rs.threshold
	}
}
