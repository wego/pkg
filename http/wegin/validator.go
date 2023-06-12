package wegin

import (
	"github.com/go-playground/validator/v10"
	"regexp"
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
	}
	structValidators = map[any]validator.StructLevelFunc{}
)

var alphaNumWithUnderscoreOrDash validator.Func = func(fl validator.FieldLevel) bool {
	return alphaNumWithUnderscoreOrDashRegex.MatchString(fl.Field().String())
}
