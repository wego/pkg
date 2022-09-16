package pointer_test

import (
	"testing"
	"time"

	"github.com/wego/pkg/pointer"
)

func TestTo(t *testing.T) {
	testCases := map[string]struct {
		value any
	}{
		"Int":     {14},
		"Uint32":  {uint32(15)},
		"Float64": {float64(16.30)},
		"String":  {"17"},
		"Bool":    {true},
		"Struct":  {time.Now()},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			p := pointer.To(tc.value)
			if p == nil {
				t.Fatalf("fail to get pointer to value %+v of type %T", tc.value, tc.value)
			}
		})
	}
}
