package common

import (
	"time"
)

// BoolRef returns a reference to a bool value
func BoolRef(v bool) *bool {
	return &v
}

// StrRef returns a reference to a string value
func StrRef(v string) *string {
	return &v
}

// Int32Ref returns a reference to a int32 value
func Int32Ref(v int32) *int32 {
	return &v
}

// Int64Ref returns a reference to a int64 value
func Int64Ref(v int64) *int64 {
	return &v
}

// UintRef returns a reference to a uint value
func UintRef(v uint) *uint {
	return &v
}

// Uint32Ref returns a reference to a uint32 value
func Uint32Ref(v uint32) *uint32 {
	return &v
}

// TimeRef return a reference to time value
func TimeRef(v time.Time) *time.Time {
	if v.IsZero() {
		return nil
	}
	return &v
}
