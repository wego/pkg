package wegin

import (
	"regexp"

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
		"flatvalue_if_gc_campaign":         giftCardFlatValue,
	}
	structValidators = map[any]validator.StructLevelFunc{}
)

var alphaNumWithUnderscoreOrDash validator.Func = func(fl validator.FieldLevel) bool {
	return alphaNumWithUnderscoreOrDashRegex.MatchString(fl.Field().String())
}

var giftCardFlatValue validator.Func = func(fl validator.FieldLevel) bool {
	if fl.Top().FieldByName("Type").String() == "GiftCard" {
		return fl.Field().String() == "FlatValue"
	}

	return true
}
