package collection_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/collection"
)

func Test_Dedup(t *testing.T) {
	as := assert.New(t)

	stringOut := collection.Dedup([]string{"b", "a", "b", "c"})
	as.ElementsMatch(stringOut, []string{"a", "b", "c"})

	intOut := collection.Dedup([]int{3, 1, 2, 1, 2, 3, 3})
	as.ElementsMatch(intOut, []int{1, 2, 3})

	floatOut := collection.Dedup([]float64{3.1, 1.2, 2.3, 1.2, 3.1})
	as.ElementsMatch(floatOut, []float64{1.2, 2.3, 3.1})

	boolOut := collection.Dedup([]bool{true, false, false, true, true})
	as.ElementsMatch(boolOut, []bool{true, false})

	type comparableStruct struct {
		i int
		f float64
		s string
		b bool
	}
	structIn := []comparableStruct{
		{1, 1.0, "1", true},
		{2, 2.0, "2", true},
		{1, 1.0, "1", true},
	}
	expectedStructOut := []comparableStruct{
		{1, 1.0, "1", true},
		{2, 2.0, "2", true},
	}
	structOut := collection.Dedup(structIn)
	as.ElementsMatch(structOut, expectedStructOut)
}
