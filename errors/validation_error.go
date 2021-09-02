package errors

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FieldError is wrapper of validator.FieldError
type FieldError struct {
	validator.FieldError
}

func (e FieldError) String() string {
	var sb strings.Builder

	sb.WriteString("validation failed on field '" + e.Field() + "'")
	sb.WriteString(", condition: " + e.ActualTag())

	if e.Param() != "" {
		sb.WriteString("=" + e.Param() + "")
	}

	if e.Value() != nil && e.Value() != "" {
		sb.WriteString(fmt.Sprintf(", actual: %v", e.Value()))
	}

	return sb.String()
}
