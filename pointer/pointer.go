package pointer

import "reflect"

// To returns the pointer to the value
func To[T any](value T) *T {
	return &value
}

// ToNonZero returns the pointer to the value if the value is not zero
func ToNonZero[T any](value T) *T {
	if reflect.ValueOf(value).IsZero() {
		return nil
	}
	return &value
}
