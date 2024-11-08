package common

import (
	"fmt"
	"strings"
	"time"
)

// Date : custom date in format yyyy-MM-dd
type Date time.Time

const layout = "2006-01-02"

// UnmarshalJSON Parses the json string to the Date
func (d *Date) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(layout, s)
	*d = Date(nt)
	return
}

// MarshalJSON marshall Date into JSON
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(d.quote()), nil
}

// UnmarshalText Parses the text string to the Date
func (d *Date) UnmarshalText(text []byte) (err error) {
	s := strings.Trim(string(text), `"`)
	parsedTime, err := time.Parse(layout, s)
	*d = Date(parsedTime)
	return
}

// MarshalText marshall Date into Text
func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.quote()), nil
}

// UnmarshalParam Parses the form's value to the Date
func (d *Date) UnmarshalParam(param string) (err error) {
	s := strings.Trim(param, `"`)
	parsedTime, err := time.Parse(layout, s)
	*d = Date(parsedTime)
	return
}

// String returns the time in the custom format
func (d Date) String() string {
	t := time.Time(d)
	return t.Format(layout)
}

// Before checks if the date is before the other date
func (d Date) Before(other Date) bool {
	return time.Time(d).Before(time.Time(other))
}

// IsZero check if date is present zero time instance
func (d Date) IsZero() bool {
	return time.Time(d).IsZero()
}

// Equal check if date is present the same time instant with other
func (d Date) Equal(other time.Time) bool {
	return time.Time(d).Equal(other)
}

func (d Date) quote() string {
	return fmt.Sprintf("%q", d.String())
}
