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

// UpdateSelective updates the value of the pointer if the value is not nil and not equal to the old value
func UpdateSelective[T comparable](oldValue, newValue *T) *T {
	if newValue == nil {
		return oldValue
	}

	if oldValue == nil {
		return newValue
	}

	if *oldValue == *newValue {
		return oldValue
	}
	return newValue
}
