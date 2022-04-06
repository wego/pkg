package collection_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/collection"
	"strings"
	"testing"
)

var (
	currencies = []string{"USD", "MYR", "SGD", "INR", "USD", "INR"}
	strs       = []string{"peach", "apple", "pear", "plum"}
	allStrs    = []string{"peach", "pear", "plum"}
)

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
	assert.Equal(2, len(resultYes))

	resultNo := collection.Filter(currencies, func(v string) bool {
		return strings.Contains(v, "JOD")
	})
	assert.Equal(0, len(resultNo))
}

func Test_All(t *testing.T) {
	assert := assert.New(t)
	resultNo := collection.All(strs, func(v string) bool {
		return strings.HasPrefix(v, "p")
	})
	resultYes := collection.All(allStrs, func(v string) bool {
		return strings.HasPrefix(v, "p")
	})
	assert.False(resultNo)
	assert.True(resultYes)
}

func Test_Map(t *testing.T) {
	assert := assert.New(t)
	resultYes := collection.Map(strs, strings.ToUpper)
	assert.Equal(resultYes[0], "PEACH")
}

func Test_MapI(t *testing.T) {
	assert := assert.New(t)
	resultYes := collection.MapI(strs, func(s string) interface{} {
		return strings.ToUpper(s)
	})
	assert.Equal(resultYes[0], "PEACH")
	assert.Equal(resultYes[1], "APPLE")
}

func Test_Distinct(t *testing.T) {
	assert := assert.New(t)
	distinct := collection.Distinct(currencies)
	assert.Equal(distinct, []string{"USD", "MYR", "SGD", "INR"})
	distinct = collection.Distinct(distinct)
	assert.Equal(distinct, []string{"USD", "MYR", "SGD", "INR"})
}

func Test_Equal(t *testing.T) {
	assert := assert.New(t)
	s1 := []string{"1", "2"}
	s2 := []string{"1", "2", "3"}
	s3 := []string{"1"}
	s4 := []string{"2", "1"}
	assert.True(collection.Equal(s1, s1))
	assert.False(collection.Equal(s1, s2))
	assert.False(collection.Equal(s1, s3))
	assert.True(collection.Equal(s1, s4))
}
