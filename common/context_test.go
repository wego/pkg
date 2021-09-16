package common_test

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/common"
)

type TestStruct struct {
	Field1 string
	Field2 int
}

var (
	key common.ContextKey = "key"
)

func Test_GetString_ReturnEmpty_WithNilContext(t *testing.T) {
	assert := assert.New(t)

	str := common.GetString(nil, key)
	assert.Empty(str)
}

func Test_GetString_ReturnEmpty_WhenKeyNotFound(t *testing.T) {
	assert := assert.New(t)

	str := common.GetString(context.Background(), key)
	assert.Empty(str)
}

func Test_GetString_ReturnEmpty_WhenValueIsNotString(t *testing.T) {
	assert := assert.New(t)

	value := 123
	ctx := context.WithValue(context.Background(), key, value)
	str := common.GetString(ctx, key)
	assert.Empty(str)
}

func Test_GetString_ReturnCorrectString_WhenKeyFound(t *testing.T) {
	assert := assert.New(t)

	value := "value"
	ctx := context.WithValue(context.Background(), key, value)
	str := common.GetString(ctx, key)
	assert.Equal(value, str)
}

func Test_SetBasic_ReturnNil_WithNilContext(t *testing.T) {
	assert := assert.New(t)
	key := "key"
	val := TestStruct{
		Field1: "f1",
		Field2: 1,
	}
	ctx := common.SetBasic(nil, key, val)
	assert.NotNil(ctx)

	assert.EqualValues(common.GetBasic(ctx, key), val)
}

func Test_SetBasic_ReturnNil_WithContext(t *testing.T) {
	assert := assert.New(t)
	key := "key"
	val := TestStruct{
		Field1: "f1",
		Field2: 1,
	}
	ctx := common.SetBasic(context.Background(), key, val)
	assert.NotNil(ctx)

	assert.EqualValues(common.GetBasic(ctx, key), val)
	assert.Nil(common.GetBasic(ctx, "1"))
}

func Test_GetBasics_ReturnNil_WithNilContext(t *testing.T) {
	assert := assert.New(t)

	basics := common.GetBasics(nil)
	assert.Nil(basics)
}

func Test_GetBasics_ReturnNil_WhenBasicsNotFound(t *testing.T) {
	assert := assert.New(t)

	basics := common.GetBasics(context.Background())
	assert.Nil(basics)
}

func Test_GetBasics_ReturnCorrectBasics(t *testing.T) {
	assert := assert.New(t)

	// test SetBasics with normal parent
	src := map[string]interface{}{"test": TestStruct{"yo", 1}}
	ctx := common.SetBasics(context.Background(), src)

	basics := common.GetBasics(ctx)
	assert.Len(basics, 1)
	value, ok := basics["test"]
	assert.True(ok)
	data, ok := value.(TestStruct)
	assert.True(ok)
	assert.Equal("yo", data.Field1)
	assert.Equal(1, data.Field2)
	// test SetBasics with nil parent
	src = map[string]interface{}{"test2": TestStruct{"yo2", 2}}
	ctx = common.SetBasics(nil, src)

	basics = common.GetBasics(ctx)
	assert.Len(basics, 1)
	value, ok = basics["test2"]
	assert.True(ok)
	data, ok = value.(TestStruct)
	assert.True(ok)
	assert.Equal("yo2", data.Field1)
	assert.Equal(2, data.Field2)
}

func Test_GetBasic_ReturnCorrectBasics(t *testing.T) {
	assert := assert.New(t)
	key1 := "test1"
	key2 := "test2"

	// test SetBasics with normal parent
	src := map[string]interface{}{key1: TestStruct{"yo", 1}}
	ctx := common.SetBasics(context.Background(), src)

	basic := common.GetBasic(ctx, key1)
	data, ok := basic.(TestStruct)
	assert.True(ok)
	assert.Equal("yo", data.Field1)
	assert.Equal(1, data.Field2)

	// test SetBasics with nil parent
	src = map[string]interface{}{key2: TestStruct{"yo2", 2}}
	ctx = common.SetBasics(nil, src)

	basic = common.GetBasic(ctx, key2)
	data, ok = basic.(TestStruct)
	assert.True(ok)
	assert.Equal("yo2", data.Field1)
	assert.Equal(2, data.Field2)
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

func Test_GetStatsD_ReturnNil_WithNilContext(t *testing.T) {
	assert := assert.New(t)

	statsD := common.GetStatsD(nil)
	assert.Nil(statsD)
}

func Test_GetStatsD_ReturnNil_WhenExtrasNotFound(t *testing.T) {
	assert := assert.New(t)

	statsD := common.GetStatsD(context.Background())
	assert.Nil(statsD)
}

func Test_GetStatsD_ReturnCorrect(t *testing.T) {
	assert := assert.New(t)

	src := &statsd.Client{
		Namespace: "Wego",
	}

	// test SetStatsD with normal parent
	ctx := common.SetStatsD(context.Background(), src)
	statsD := common.GetStatsD(ctx)
	assert.NotNil(statsD)
	assert.Equal(src.Namespace, statsD.Namespace)

	// test SetStatsD with nil parent
	ctx = common.SetStatsD(nil, src)
	statsD = common.GetStatsD(ctx)
	assert.NotNil(statsD)
	assert.Equal(src.Namespace, statsD.Namespace)
}
