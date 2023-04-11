package datetime

import (
	"strconv"
	"strings"
	"time"
)

// EpochTime represents time from string epoch format
type EpochTime time.Time

// UnmarshalJSON Parses the json string epoch time to time.Time
func (e *EpochTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	epoch, err := strconv.ParseInt(s, 10, 64)
	convertedTime := time.UnixMilli(epoch)

	*e = EpochTime(convertedTime)

	return
}
