package snowflake

import "sync/atomic"

var (
	lastTimeSlot int64
	lastSequence uint32
)

// AtomicGenerator define as atomic sequence Resolver, base on standard sync/atomic.
func AtomicGenerator(currentTime int64) (uint16, error) {
	var lastTime int64
	var currentSeq, lastSeq uint32

	for {
		lastTime = atomic.LoadInt64(&lastTimeSlot)
		lastSeq = atomic.LoadUint32(&lastSequence)

		if lastTime > currentTime {
			return maxSequence, nil
		}

		if lastTime == currentTime {
			currentSeq = maxSequence & (lastSeq + 1)
			if currentSeq == 0 {
				return maxSequence, nil
			}
		}

		if atomic.CompareAndSwapInt64(&lastTimeSlot, lastTime, currentTime) &&
			atomic.CompareAndSwapUint32(&lastSequence, lastSeq, currentSeq) {
			return uint16(currentSeq), nil
		}
	}
}
