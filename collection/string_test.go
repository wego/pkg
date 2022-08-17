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
	assertions := assert.New(t)

	indexZero := collection.Index(currencies, "USD")
	assertions.Equal(0, indexZero)

	nonIndex := collection.Index(currencies, "AED")
	assertions.Equal(-1, nonIndex)
}

func Test_Include(t *testing.T) {
	assertions := assert.New(t)
	assertions.True(collection.Include(currencies, "USD"))
	assertions.False(collection.Include(currencies, "AED"))
}

func Test_Any(t *testing.T) {
	assertions := assert.New(t)
	assertions.True(collection.Any(currencies, func(v string) bool {
		return strings.HasPrefix(v, "U")
	}))
	assertions.False(collection.Any(currencies, func(v string) bool {
		return strings.HasPrefix(v, "X")
	}))
}

func Test_Filter(t *testing.T) {
	assertions := assert.New(t)
	resultYes := collection.Filter(currencies, func(v string) bool {
		return strings.Contains(v, "USD")
	})
	assertions.Equal(2, len(resultYes))

	resultNo := collection.Filter(currencies, func(v string) bool {
		return strings.Contains(v, "JOD")
	})
	assertions.Equal(0, len(resultNo))
}

func Test_All(t *testing.T) {
	assertions := assert.New(t)
	resultNo := collection.All(strs, func(v string) bool {
		return strings.HasPrefix(v, "p")
	})
	resultYes := collection.All(allStrs, func(v string) bool {
		return strings.HasPrefix(v, "p")
	})
	assertions.False(resultNo)
	assertions.True(resultYes)
}

func Test_Map(t *testing.T) {
	assertions := assert.New(t)
	resultYes := collection.Map(strs, strings.ToUpper)
	assertions.Equal(resultYes[0], "PEACH")
}

func Test_MapI(t *testing.T) {
	assertions := assert.New(t)
	resultYes := collection.MapI(strs, func(s string) interface{} {
		return strings.ToUpper(s)
	})
	assertions.Equal(resultYes[0], "PEACH")
	assertions.Equal(resultYes[1], "APPLE")
}

func Test_Distinct(t *testing.T) {
	assertions := assert.New(t)
	distinct := collection.Distinct(currencies)
	assertions.Equal(distinct, []string{"USD", "MYR", "SGD", "INR"})
	distinct = collection.Distinct(distinct)
	assertions.Equal(distinct, []string{"USD", "MYR", "SGD", "INR"})
}

func Test_Equal(t *testing.T) {
	assertions := assert.New(t)
	s1 := []string{"1", "2"}
	s2 := []string{"1", "2", "3"}
	s3 := []string{"1"}
	s4 := []string{"1", "3"}
	s5 := []string{"2", "1"}
	assertions.True(collection.Equal(s1, s1))
	assertions.False(collection.Equal(s1, s2))
	assertions.False(collection.Equal(s1, s3))
	assertions.False(collection.Equal(s1, s4))
	assertions.True(collection.Equal(s1, s5))
}
