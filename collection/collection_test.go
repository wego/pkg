package collection_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/collection"
	"strings"
	"testing"
)

var currencies = []string{"USD", "MYR", "SGD", "INR"}

func Test_Index(t *testing.T) {
	assert := assert.New(t)

	indexZero := collection.Index(currencies, "USD")
	assert.Equal(0, indexZero)

	nonIndex := collection.Index(currencies, "AED")
	assert.Equal(-1, nonIndex)
}

func Test_Include(t *testing.T) {
	assert := assert.New(t)
	assert.True(collection.Include(currencies, "USD"))
	assert.False(collection.Include(currencies, "AED"))
}

func Test_Any(t *testing.T) {
	assert := assert.New(t)
	assert.True(collection.Any(currencies, func(v string) bool {
		return strings.HasPrefix(v, "U")
	}))
	assert.False(collection.Any(currencies, func(v string) bool {
		return strings.HasPrefix(v, "X")
	}))
}

func Test_Filter(t *testing.T) {
	assert := assert.New(t)
	resultYes := collection.Filter(currencies, func(v string) bool {
		return strings.Contains(v, "USD")
	})
	assert.Equal(1, len(resultYes))

	resultNo := collection.Filter(currencies, func(v string) bool {
		return strings.Contains(v, "JOD")
	})
	assert.Equal(0, len(resultNo))
}
