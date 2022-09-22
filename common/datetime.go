package common

import (
	"time"

	"github.com/wego/pkg/pointer"
)

// CurrentUTCTime returns current date time in UTC
func CurrentUTCTime() time.Time {
	return time.Now().UTC()
}

// CurrentUTCTimeRef returns current date time reference in UTC
func CurrentUTCTimeRef() *time.Time {
	return pointer.To(time.Now().UTC())
}

// CurrentLocalTime returns current date time in local timezone
func CurrentLocalTime() time.Time {
	return time.Now()
}

// CurrentUnixTimestamp returns current Unix timestamp (seconds)
func CurrentUnixTimestamp() int64 {
	return time.Now().Unix()
}

// TimeChanged check if new time changed from the base, return the changed flag and time value
func TimeChanged(new *time.Time, base *time.Time) (bool, *time.Time) {
	if base != nil && new != nil {
		return !new.Equal(*base), new
	}

	if new != nil && base == nil {
		return true, new
	}

	if new == nil && base != nil {
		return false, base
	}
	return false, new
}
