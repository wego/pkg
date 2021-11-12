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
	var field string

	// try to get full path of field if it's a nested field
	namespace := e.Namespace()
	if namespace != "" {
		// namespace has format RootStruct.parent_field.child_field
		// we don't want to display the RootStruct because it doesn't make sense to the client
		field = namespace[strings.Index(namespace, ".")+1:]
	}
	if field == "" {
		field = e.Field()
	}

	sb.WriteString("validation failed on field '" + field + "'")
	sb.WriteString(", condition: " + e.ActualTag())

	if e.Param() != "" {
		sb.WriteString("=" + e.Param() + "")
	}

	if e.Value() != nil && e.Value() != "" {
		sb.WriteString(fmt.Sprintf(", actual: %v", e.Value()))
	}

	return sb.String()
}
