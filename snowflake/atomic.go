package snowflake

import "sync/atomic"

var lastTime int64
var lastSequence uint32

// AtomicResolver define as atomic sequence Resolver, base on standard sync/atomic.
func AtomicResolver(current int64) (uint16, error) {
	var last int64
	var currentSeq, lastSeq uint32

	for {
		last = atomic.LoadInt64(&lastTime)
		lastSeq = atomic.LoadUint32(&lastSequence)

		if last > current {
			return maxSequence, nil
		}

		if last == current {
			currentSeq = maxSequence & (lastSeq + 1)
			if currentSeq == 0 {
				return maxSequence, nil
			}
		}

		if atomic.CompareAndSwapInt64(&lastTime, last, current) &&
			atomic.CompareAndSwapUint32(&lastSequence, lastSeq, currentSeq) {
			return uint16(currentSeq), nil
		}
	}
}
