package common_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/common"
	"strings"
	"testing"
)

var currencies = []string{"USD", "MYR", "SGD", "INR"}

func Test_Index(t *testing.T) {
	assert := assert.New(t)

	indexZero := common.Index(currencies, "USD")
	assert.Equal(0, indexZero)

	nonIndex := common.Index(currencies, "AED")
	assert.Equal(-1, nonIndex)
}

func Test_Include(t *testing.T) {
	assert := assert.New(t)
	assert.True(common.Include(currencies, "USD"))
	assert.False(common.Include(currencies, "AED"))
}

func Test_Any(t *testing.T) {
	assert := assert.New(t)
	assert.True(common.Any(currencies, func(v string) bool {
		return strings.HasPrefix(v, "U")
	}))
	assert.False(common.Any(currencies, func(v string) bool {
		return strings.HasPrefix(v, "X")
	}))
}

func Test_Filter(t *testing.T) {
	assert := assert.New(t)
	resultYes := common.Filter(currencies, func(v string) bool {
		return strings.Contains(v, "USD")
	})
	assert.Equal(1, len(resultYes))

	resultNo := common.Filter(currencies, func(v string) bool {
		return strings.Contains(v, "JOD")
	})
	assert.Equal(0, len(resultNo))
}
