package wegin

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	alphaNumWithUnderscoreOrDashRegexString = `^[a-zA-Z0-9_\-]+$`
)

var (
	alphaNumWithUnderscoreOrDashRegex = regexp.MustCompile(alphaNumWithUnderscoreOrDashRegexString)
)

var (
	fieldValidators = map[string]validator.Func{
		"alphanum_with_underscore_or_dash": alphaNumWithDash,
		"one_of_or_blank":                  isOneOfOrBlank, // only for string or string pointer types
	}
	structValidators = map[any]validator.StructLevelFunc{}
)

var alphaNumWithDash validator.Func = func(fl validator.FieldLevel) bool {
	return alphaNumWithUnderscoreOrDashRegex.MatchString(fl.Field().String())
}

var isOneOfOrBlank validator.Func = func(fl validator.FieldLevel) bool {
	vals := parseOneOfParam(fl.Param())
	field := fl.Field()

	// Handle different types
	var v string

	// String pointer type (preferred usage)
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return true // Nil pointers are allowed
		}

		if field.Type().Elem().Kind() != reflect.String {
			return false // Non-string pointers are not supported
		}
		v = strings.TrimSpace(field.Elem().String())
	} else if field.Kind() == reflect.String {
		v = strings.TrimSpace(field.String())
	} else {
		// Unsupported type
		return false
	}

	// Check if v is empty
	if v == "" {
		return true
	}

	// Check if value is in the allowed list
	for _, val := range vals {
		if val == v {
			return true
		}
	}

	// If we get here, the value is not empty and not in the allowed list
	return false
}
