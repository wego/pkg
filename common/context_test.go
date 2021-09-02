package common_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/common"
)

type TestStruct struct {
	Field1 string
	Field2 int
}

func Test_GetString_ReturnEmpty_WithNilContext(t *testing.T) {
	assert := assert.New(t)

	str := common.GetString(nil, common.CtxClientCode)
	assert.Empty(str)
}

func Test_GetString_ReturnEmpty_WhenKeyNotFound(t *testing.T) {
	assert := assert.New(t)

	str := common.GetString(context.Background(), common.CtxClientCode)
	assert.Empty(str)
}

func Test_GetString_ReturnEmpty_WhenValueIsNotString(t *testing.T) {
	assert := assert.New(t)

	value := 123
	ctx := context.WithValue(context.Background(), common.CtxClientCode, value)
	str := common.GetString(ctx, common.CtxClientCode)
	assert.Empty(str)
}

func Test_GetString_ReturnCorrectString_WhenKeyFound(t *testing.T) {
	assert := assert.New(t)

	value := "value"
	ctx := context.WithValue(context.Background(), common.CtxClientCode, value)
	str := common.GetString(ctx, common.CtxClientCode)
	assert.Equal(value, str)
}

func Test_GetExtras_ReturnNil_WithNilContext(t *testing.T) {
	assert := assert.New(t)

	extras := common.GetExtras(nil)
	assert.Nil(extras)
}

func Test_GetExtras_ReturnNil_WhenExtrasNotFound(t *testing.T) {
	assert := assert.New(t)

	extras := common.GetExtras(context.Background())
	assert.Nil(extras)
}

func Test_GetExtras_ReturnCorrectExtras(t *testing.T) {
	assert := assert.New(t)

	// test SetExtras with normal parent
	src := map[string]interface{}{"test": TestStruct{"yo", 1}}
	ctx := common.SetExtras(context.Background(), src)

	extras := common.GetExtras(ctx)
	assert.Len(extras, 1)
	value, ok := extras["test"]
	assert.True(ok)
	data, ok := value.(TestStruct)
	assert.True(ok)
	assert.Equal("yo", data.Field1)
	assert.Equal(1, data.Field2)

	// test SetExtras with nil parent
	src = map[string]interface{}{"test2": TestStruct{"yo2", 2}}
	ctx = common.SetExtras(nil, src)

	extras = common.GetExtras(ctx)
	assert.Len(extras, 1)
	value, ok = extras["test2"]
	assert.True(ok)
	data, ok = value.(TestStruct)
	assert.True(ok)
	assert.Equal("yo2", data.Field1)
	assert.Equal(2, data.Field2)
}
