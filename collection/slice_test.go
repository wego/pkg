package collection_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/collection"
	"testing"
)

func Test_EqualValues(t *testing.T) {
	assertions := assert.New(t)

	assertions.False(collection.EqualValuesI([]interface{}{"USD", "USD"}, []interface{}{"USD"}))
	assertions.False(collection.EqualValuesI([]interface{}{"USD", "AED", "SAR"}, []interface{}{"AED", "USD"}))
	assertions.True(collection.EqualValuesI([]interface{}{}, []interface{}{}))
	assertions.True(collection.EqualValuesI([]interface{}{"USD"}, []interface{}{"USD"}))
	assertions.True(collection.EqualValuesI([]interface{}{"USD", "AED"}, []interface{}{"AED", "USD"}))
	assertions.True(collection.EqualValuesI([]interface{}{"USD", "USD"}, []interface{}{"USD", "USD"}))
}
