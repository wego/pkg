package datetime

import (
	"strconv"
	"strings"
	"time"
)

const (
	nullString = "null"
)

// EpochTime represents time from string epoch format in milliseconds
type EpochTime time.Time

// UnmarshalJSON Parses the json string epoch time to time.Time
func (e *EpochTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if strings.ToLower(s) == nullString {
		return
	}
	epoch, err := strconv.ParseInt(s, 10, 64)
	convertedTime := time.UnixMilli(epoch)

	*e = EpochTime(convertedTime)

	return
}

// EpochTimeSeconds represents time from string epoch format in seconds
type EpochTimeSeconds time.Time

// UnmarshalJSON Parses the json string epoch time to time.Time
func (e *EpochTimeSeconds) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if strings.ToLower(s) == nullString {
		return
	}
	epoch, err := strconv.ParseInt(s, 10, 64)
	convertedTime := time.Unix(epoch, 0)

	*e = EpochTimeSeconds(convertedTime)

	return
}
