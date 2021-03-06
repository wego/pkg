package binding

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wego/pkg/audit"
	"github.com/wego/pkg/errors"
)

// ShouldBindJSON binds general JSON request. It will ignore request without body.
func ShouldBindJSON(c *gin.Context, ctxKey string, request interface{}) error {
	if fromContext(c, ctxKey, request) {
		return nil
	}

	if c.Request.Body != nil && c.Request.Body != http.NoBody {
		return bindJSON(c, ctxKey, request)
	}
	return nil
}

// BindJSON binds general JSON request. It will return error if request doesn't have body.
func BindJSON(c *gin.Context, ctxKey string, request interface{}) error {
	if fromContext(c, ctxKey, request) {
		return nil
	}

	return bindJSON(c, ctxKey, request)
}

// BindQuery Bind general form request
func BindQuery(c *gin.Context, ctxKey string, request interface{}) (err error) {
	// try to get from context
	if fromContext(c, ctxKey, request) {
		return
	}

	// try to bind from request & set to context if ok
	if err = c.ShouldBindQuery(request); err != nil {
		err = errors.New(errors.BadRequest, err)
		return
	}
	c.Set(ctxKey, request)
	return
}

// BindChangeRequest Bind general change request(Update/Delete)
func BindChangeRequest(c *gin.Context, ctxKey string, request audit.IChangeRequest) (err error) {
	// try to get from context
	if fromContext(c, ctxKey, request) {
		return
	}

	// try to bind from request & set to context if ok
	if err = c.ShouldBindJSON(request); err != nil {
		err = errors.New(errors.BadRequest, err)
		return
	}
	var id uint
	if id, err = BindID(c); err != nil {
		return
	}
	request.SetID(id)
	c.Set(ctxKey, request)
	return
}

// BindID Bind param ID
func BindID(c *gin.Context) (id uint, err error) {
	idParam := c.Param("id")
	val, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || val == 0 {
		err = errors.New(errors.BadRequest, fmt.Sprintf("invalid id [%s]", idParam))
		return
	}
	id = uint(val)
	return
}

func fromContext(c *gin.Context, ctxKey string, value interface{}) bool {
	// try to get from context
	fromCtx, ok := c.Get(ctxKey)
	if ok && reflect.TypeOf(fromCtx) == reflect.TypeOf(value) {
		if val := reflect.ValueOf(fromCtx); val.Kind() == reflect.Ptr {
			reflect.ValueOf(value).Elem().Set(val.Elem())
			return true
		}
	}
	return false
}

// bindJSON tries to bind JSON object from request body & set to context if ok
func bindJSON(c *gin.Context, ctxKey string, request interface{}) (err error) {
	if err = c.ShouldBindJSON(request); err != nil {
		return errors.New(errors.BadRequest, err)
	}
	c.Set(ctxKey, request)
	return
}
