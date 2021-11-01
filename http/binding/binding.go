package binding

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wego/pkg/audit"
	"github.com/wego/pkg/errors"
	"reflect"
	"strconv"
)

// BindJSON Bind general json request
func BindJSON(c *gin.Context, ctxKey string, request interface{}) (err error) {
	// try to get from context
	if fromContext(c, ctxKey, request) {
		return
	}

	// try to bind from request & set to context if ok
	if err = c.ShouldBindJSON(request); err != nil {
		err = errors.New(errors.BadRequest, err)
		return
	}
	c.Set(ctxKey, request)
	return
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
		fmt.Println(err)
		err = errors.New(errors.BadRequest, err)
		return
	}
	fmt.Println(request)
	var id uint64
	idParam := c.Param("id")
	id, err = strconv.ParseUint(idParam, 10, 64)
	if err != nil || id == 0 {
		err = errors.New(errors.BadRequest, fmt.Sprintf("invalid id [%s]", idParam))
		return
	}
	request.SetID(uint(id))
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
