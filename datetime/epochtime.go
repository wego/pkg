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
	var epoch int64

	epoch, err = parseEpochTime(b)
	if err != nil || epoch == 0 {
		return
	}
	*e = EpochTime(time.UnixMilli(epoch))
	return
}

// EpochTimeSeconds represents time from string epoch format in seconds
type EpochTimeSeconds time.Time

// UnmarshalJSON Parses the json string epoch time to time.Time
func (e *EpochTimeSeconds) UnmarshalJSON(b []byte) (err error) {
	var epoch int64

	epoch, err = parseEpochTime(b)
	if err != nil || epoch == 0 {
		return
	}
	*e = EpochTimeSeconds(time.Unix(epoch, 0))
	return
}

func parseEpochTime(b []byte) (epoch int64, err error) {
	s := strings.Trim(string(b), "\"")
	if strings.ToLower(s) != nullString {
		epoch, err = strconv.ParseInt(s, 10, 64)
		return
	}
	return
}
