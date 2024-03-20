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
		"alphanum_with_underscore_or_dash": alphaNumWithUnderscoreOrDash,
		"value_if":                         valueIf,
	}
	structValidators = map[any]validator.StructLevelFunc{}
)

var alphaNumWithUnderscoreOrDash validator.Func = func(fl validator.FieldLevel) bool {
	return alphaNumWithUnderscoreOrDashRegex.MatchString(fl.Field().String())
}

// / value_if tag - to make this simple, we will only deal with strings
var valueIf validator.Func = func(fl validator.FieldLevel) bool {
	params := strings.Split(fl.Param(), ">")
	constraint := strings.Split(strings.TrimSpace(params[0]), " ")
	expectedValue := strings.TrimSpace(params[1])

	if len(constraint) != 2 {
		return false
	}

	types := strings.Split(strings.TrimSpace(constraint[0]), ".")
	typesLen := len(types)
	currentStruct := fl.Top()

	// var currentStructName string
	for i, typeName := range types {
		// we've already popped the Top above
		if i > 0 {
			currentStruct = currentStruct.FieldByName(typeName)
		}

		if !currentStruct.IsValid() {
			return false
		}

		// fail if we cannot find the field
		if currentStruct.Kind() == reflect.Ptr {
			currentStruct = currentStruct.Elem()
		}

		// we can only traverse on struct type
		if currentStruct.Kind().String() != "struct" && i < (typesLen-1) {
			return false
		}
	}

	if currentStruct.String() != strings.TrimSpace(constraint[1]) {
		return false
	}

	if fieldType := reflect.Kind(fl.Field().Kind()); fieldType == reflect.String {
		return fl.Field().String() == expectedValue
	}

	return false
}
