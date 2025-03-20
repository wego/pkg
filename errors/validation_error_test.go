package errors

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type RootStruct struct {
	TopLevel int    `json:"top_level" validate:"required"`
	Parent   Parent `json:"parent_field"`
}

type Parent struct {
	Child string `json:"child_field" validate:"required"`
}

func Test_FieldError_String_Ok(t *testing.T) {
	assert := assert.New(t)

	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	obj := RootStruct{}
	err := v.Struct(obj)
	validationErr, ok := err.(validator.ValidationErrors)
	assert.True(ok)

	errStrings := make([]string, len(validationErr))
	for i, fieldErr := range validationErr {
		errStrings[i] = FieldError{FieldError: fieldErr}.String()
	}

	assert.Len(errStrings, 2)
	assert.Contains(errStrings[0], "validation failed on field 'top_level', condition: required")
	assert.Contains(errStrings[1], "validation failed on field 'parent_field.child_field', condition: required")
}
