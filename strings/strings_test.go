package strings_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/pointer"
	"github.com/wego/pkg/strings"
)

func Test_IsBlankP(t *testing.T) {
	assert := assert.New(t)
	assert.True(strings.IsBlankP(nil))
	assert.True(strings.IsBlankP(pointer.To("")))
	assert.True(strings.IsBlankP(pointer.To("  ")))
	assert.False(strings.IsBlankP(pointer.To("str")))
	assert.False(strings.IsBlankP(pointer.To(" str ")))
}

func Test_IsBlank(t *testing.T) {
	assert := assert.New(t)
	assert.True(strings.IsBlank(""))
	assert.True(strings.IsBlank("  "))
	assert.False(strings.IsBlank("str"))
	assert.False(strings.IsBlank(" str "))
}

func Test_IsNotBlankP(t *testing.T) {
	assert := assert.New(t)
	assert.False(strings.IsNotBlankP(nil))
	assert.False(strings.IsNotBlankP(pointer.To("")))
	assert.False(strings.IsNotBlankP(pointer.To("  ")))
	assert.True(strings.IsNotBlankP(pointer.To("str")))
	assert.True(strings.IsNotBlankP(pointer.To(" str ")))
}

func Test_IsNotBlank(t *testing.T) {
	assert := assert.New(t)
	assert.False(strings.IsNotBlank(""))
	assert.False(strings.IsNotBlank("  "))
	assert.True(strings.IsNotBlank("str"))
	assert.True(strings.IsNotBlank(" str "))
}

func Test_IsEmptyP(t *testing.T) {
	assert := assert.New(t)
	assert.True(strings.IsEmptyP(nil))
	assert.True(strings.IsEmptyP(pointer.To("")))
	assert.False(strings.IsEmptyP(pointer.To("  ")))
	assert.False(strings.IsEmptyP(pointer.To("str")))
	assert.False(strings.IsEmptyP(pointer.To(" str ")))
}

func Test_IsEmpty(t *testing.T) {
	assert := assert.New(t)
	assert.True(strings.IsEmpty(""))
	assert.False(strings.IsEmpty("  "))
	assert.False(strings.IsEmpty("str"))
	assert.False(strings.IsEmpty(" str "))
}

func Test_IsNotEmptyP(t *testing.T) {
	assert := assert.New(t)
	assert.False(strings.IsNotEmptyP(nil))
	assert.False(strings.IsNotEmptyP(pointer.To("")))
	assert.True(strings.IsNotEmptyP(pointer.To("  ")))
	assert.True(strings.IsNotEmptyP(pointer.To("str")))
	assert.True(strings.IsNotEmptyP(pointer.To(" str ")))
}

func Test_IsNotEmpty(t *testing.T) {
	assert := assert.New(t)
	assert.False(strings.IsNotEmpty(""))
	assert.True(strings.IsNotEmpty("  "))
	assert.True(strings.IsNotEmpty("str"))
	assert.True(strings.IsNotEmpty(" str "))
}

func Test_PointerValue(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("", strings.PointerValue(nil))
	assert.Equal("", strings.PointerValue(pointer.To("")))
	assert.Equal("  ", strings.PointerValue(pointer.To("  ")))
	assert.Equal("str", strings.PointerValue(pointer.To("str")))
	assert.Equal(" str ", strings.PointerValue(pointer.To(" str ")))
}
