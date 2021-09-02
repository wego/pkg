package errors

import (
	"net/http"

	goerrors "errors"

	"gorm.io/gorm"
)

// Op : operation
type Op string

// Kind error kind
type Kind int

// Error : wego custom error
type Error struct {
	Op   Op
	Kind Kind
	Err  error
	msg  string
}

// error kinds
const (
	BadRequest    Kind = http.StatusBadRequest
	Conflict      Kind = http.StatusConflict
	Forbidden     Kind = http.StatusForbidden
	NotFound      Kind = http.StatusNotFound
	Unauthorized  Kind = http.StatusUnauthorized
	Unprocessable Kind = http.StatusUnprocessableEntity
	Unexpected    Kind = http.StatusInternalServerError
)

// New construct a new error, default having kind Unexpected
func New(args ...interface{}) error {
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case int:
			e.Kind = Kind(arg)
		case Op:
			e.Op = arg
		case Kind:
			e.Kind = arg
		case error:
			e.Err = arg
		case string:
			e.msg = arg
		}
	}
	return e
}

// Code return HTTP status code of the error
func Code(err error) int {
	e, ok := err.(*Error)
	if !ok {
		return int(Unexpected)
	}
	if e.Kind != 0 {
		return int(e.Kind)
	}
	return Code(e.Err)
}

func (e *Error) Error() string {
	var msg string
	if e.msg != "" {
		msg = e.msg
	}
	if e.Err != nil {
		if msg != "" {
			msg += ": "
		}
		msg += e.Err.Error()
	}

	if msg == "" {
		msg = "unknown error"
	}
	return msg
}

// ops return the stack of operation
func ops(e *Error) []Op {
	res := []Op{}
	if e.Op != "" {
		res = append(res, e.Op)
	}

	subErr, ok := e.Err.(*Error)
	if !ok {
		return res
	}

	res = append(res, ops(subErr)...)
	return res
}

// WrapGORMError wraps an GORM error into our error such as adding errors.Kind
func WrapGORMError(op Op, err error) error {
	if goerrors.Is(err, gorm.ErrRecordNotFound) {
		return New(op, NotFound, err)
	}
	return New(op, err)
}
