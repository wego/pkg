package errors

import (
	"context"
	"net/http"

	goErrors "errors"

	"github.com/wego/pkg/collection"
	"github.com/wego/pkg/common"
	"gorm.io/gorm"
)

// Op : operation
type Op string

// Kind error kind
type Kind int

// Error custom error
type Error struct {
	Op   Op
	Kind Kind
	Err  error
	msg  string
	ctx  map[string]any
}

// error kinds
const (
	BadRequest      Kind = http.StatusBadRequest
	Conflict        Kind = http.StatusConflict
	Forbidden       Kind = http.StatusForbidden
	NotFound        Kind = http.StatusNotFound
	Unauthorized    Kind = http.StatusUnauthorized
	Unprocessable   Kind = http.StatusUnprocessableEntity
	TooManyRequests Kind = http.StatusTooManyRequests
	Unexpected      Kind = http.StatusInternalServerError
	Retry           Kind = -1 // Retry indicate an error we need to retry the action
	NotSupported    Kind = -2 // NotSupported THe requested action/resource is not supported
	NotImplemented  Kind = -3 // NotImplemented The requested action/resource is not implemented
)

const (
	ctxBasics = "basics"
	ctxExtras = "extras"
)

// sentry keys
const (
	SentryErrorCode  = "error_code"
	SentryOperations = "operations"
	SentryRequestID  = "request_id"
)

var (
	// ErrNotSupported the requested action/resource is not supported
	ErrNotSupported = New(nil, NotSupported, "not supported")
	// ErrNotImplemented the requested action/resource is not implemented
	ErrNotImplemented = New(nil, NotImplemented, "not implemented")
)

// New construct a new error, default having kind Unexpected
func New(ctx context.Context, args ...interface{}) error {
	e := &Error{
		ctx: map[string]any{},
	}

	basics := common.GetBasics(ctx)
	if basics != nil {
		e.ctx[ctxBasics] = basics
	}

	extras := common.GetExtras(ctx)
	if extras != nil {
		e.ctx[ctxExtras] = extras
	}

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
			e.propagateContexts()
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
	var res []Op
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
	if goErrors.Is(err, gorm.ErrRecordNotFound) {
		return New(nil, op, NotFound, err)
	}

	if goErrors.Is(err, gorm.ErrDuplicatedKey) {
		return New(nil, op, Conflict, err)
	}

	return New(nil, op, err)
}

// propagateContexts combines the "basics" and "extras" contexts from the child error into the parent, so that the
// key-values propagate upwards to the top level error.
func (e *Error) propagateContexts() {
	subErr, ok := e.Err.(*Error)
	if !ok {
		return
	}

	subBasics, _ := subErr.ctx[ctxBasics].(common.Basics)
	if subBasics != nil {
		basics, basicsOK := e.ctx[ctxBasics].(common.Basics)
		if !basicsOK {
			basics = common.Basics{}
		}

		collection.Copy(basics, subBasics)
		e.ctx[ctxBasics] = basics
	}

	subExtras, _ := subErr.ctx[ctxExtras].(common.Extras)
	if subExtras != nil {
		extras, extrasOK := e.ctx[ctxExtras].(common.Extras)
		if !extrasOK {
			extras = common.Extras{}
		}

		collection.Copy(extras, subExtras)
		e.ctx[ctxExtras] = extras
	}

	subErr.ctx = nil
}

func (e *Error) basics() common.Basics {
	basics, _ := e.ctx[ctxBasics].(common.Basics)
	return basics
}

func (e *Error) extras() common.Extras {
	extras, _ := e.ctx[ctxExtras].(common.Extras)
	return extras
}
