package strings

import (
	goStr "strings"
)

// IsBlankP returns true if the string length is zero(excluding whitespaces) or null
func IsBlankP(s *string) bool {
	return !IsNotBlankP(s)
}

// IsBlank returns true if the string length is zero(excluding whitespaces)
func IsBlank(s string) bool {
	return !IsNotBlank(s)
}

// IsNotBlankP returns true if the string length is not zero(excluding whitespaces) and not null
func IsNotBlankP(s *string) bool {
	return s != nil && IsNotBlank(*s)
}

// IsNotBlank returns true if the string length is not zero(excluding whitespaces)
func IsNotBlank(s string) bool {
	return len(goStr.TrimSpace(s)) > 0
}

// IsEmptyP returns true if the string length is zero or null
func IsEmptyP(s *string) bool {
	return !IsNotEmptyP(s)
}

// IsEmpty returns true if the string length is zero
func IsEmpty(s string) bool {
	return !IsNotEmpty(s)
}

// IsNotEmptyP returns true if the string length is not zero and not null
func IsNotEmptyP(s *string) bool {
	return s != nil && IsNotEmpty(*s)
}

// IsNotEmpty returns true if the string length is not zero
func IsNotEmpty(s string) bool {
	return len(s) > 0
}

// PointerValue returns the string value from the pointer or empty string if nil
func PointerValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
