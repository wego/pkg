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

func TestToNonZero(t *testing.T) {
	testCases := map[string]struct {
		value any
		null  bool
	}{
		"Int NonZero":           {14, false},
		"Int Zero":              {0, true},
		"Uint32 NonZero":        {uint32(15), false},
		"Uint32 Zero":           {uint32(0), true},
		"Float64 NonZero":       {float64(16.30), false},
		"Float64 Zero":          {float64(0), true},
		"String NonZero":        {"17", false},
		"String Zero":           {"", true},
		"Bool NonZero":          {true, false},
		"Bool Zero":             {false, true},
		"Struct NonZero":        {time.Now(), false},
		"Struct Zero":           {time.Time{}, true},
		"Custom struct NonZero": {struct{ a int }{1}, false},
		"Custom struct Zero":    {struct{ a int }{}, true},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			p := pointer.ToNonZero(tc.value)
			if (p == nil) != tc.null {
				t.Fatalf("fail to get pointer to value %+v of type %T", tc.value, tc.value)
			}
		})
	}
}
