package collection_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/collection"
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
	assertions.Equal(indexZero, collection.IndexOf(currencies, "USD"))

	nonIndex := collection.Index(currencies, "AED")
	assertions.Equal(-1, nonIndex)
	assertions.Equal(nonIndex, collection.IndexOf(currencies, "AED"))
}

func Test_Include(t *testing.T) {
	assertions := assert.New(t)
	assertions.True(collection.Include(currencies, "USD"))
	assertions.True(collection.Contains(currencies, "USD"))
	assertions.False(collection.Include(currencies, "AED"))
	assertions.False(collection.Contains(currencies, "AED"))
}
