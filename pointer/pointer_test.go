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

func TestUpdateSelective(t *testing.T) {
	type testStruct struct {
		ID   int
		Name string
	}

	t.Run("Int", func(t *testing.T) {
		testCases := map[string]struct {
			oldValue *int
			newValue *int
			expected *int
		}{
			"Both nil": {
				oldValue: nil,
				newValue: nil,
				expected: nil,
			},
			"Old nil, new has value": {
				oldValue: nil,
				newValue: pointer.To(42),
				expected: pointer.To(42),
			},
			"Old has value, new nil": {
				oldValue: pointer.To(10),
				newValue: nil,
				expected: pointer.To(10),
			},
			"Both have same value": {
				oldValue: pointer.To(15),
				newValue: pointer.To(15),
				expected: pointer.To(15),
			},
			"Both have different values": {
				oldValue: pointer.To(20),
				newValue: pointer.To(30),
				expected: pointer.To(30),
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				result := pointer.UpdateSelective(tc.oldValue, tc.newValue)

				if tc.expected == nil {
					if result != nil {
						t.Errorf("expected nil, got %v", *result)
					}
					return
				}

				if result == nil {
					t.Errorf("expected %v, got nil", *tc.expected)
					return
				}

				if *result != *tc.expected {
					t.Errorf("expected %v, got %v", *tc.expected, *result)
				}
			})
		}
	})

	t.Run("String", func(t *testing.T) {
		testCases := map[string]struct {
			oldValue *string
			newValue *string
			expected *string
		}{
			"Both nil": {
				oldValue: nil,
				newValue: nil,
				expected: nil,
			},
			"Old nil, new has value": {
				oldValue: nil,
				newValue: pointer.To("hello"),
				expected: pointer.To("hello"),
			},
			"Old has value, new nil": {
				oldValue: pointer.To("world"),
				newValue: nil,
				expected: pointer.To("world"),
			},
			"Both have same value": {
				oldValue: pointer.To("test"),
				newValue: pointer.To("test"),
				expected: pointer.To("test"),
			},
			"Both have different values": {
				oldValue: pointer.To("old"),
				newValue: pointer.To("new"),
				expected: pointer.To("new"),
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				result := pointer.UpdateSelective(tc.oldValue, tc.newValue)

				if tc.expected == nil {
					if result != nil {
						t.Errorf("expected nil, got %v", *result)
					}
					return
				}

				if result == nil {
					t.Errorf("expected %v, got nil", *tc.expected)
					return
				}

				if *result != *tc.expected {
					t.Errorf("expected %v, got %v", *tc.expected, *result)
				}
			})
		}
	})

	t.Run("Struct", func(t *testing.T) {
		testCases := map[string]struct {
			oldValue *testStruct
			newValue *testStruct
			expected *testStruct
		}{
			"Both nil": {
				oldValue: nil,
				newValue: nil,
				expected: nil,
			},
			"Old nil, new has value": {
				oldValue: nil,
				newValue: &testStruct{ID: 1, Name: "test"},
				expected: &testStruct{ID: 1, Name: "test"},
			},
			"Old has value, new nil": {
				oldValue: &testStruct{ID: 2, Name: "old"},
				newValue: nil,
				expected: &testStruct{ID: 2, Name: "old"},
			},
			"Both have same value": {
				oldValue: &testStruct{ID: 3, Name: "same"},
				newValue: &testStruct{ID: 3, Name: "same"},
				expected: &testStruct{ID: 3, Name: "same"},
			},
			"Both have different values": {
				oldValue: &testStruct{ID: 4, Name: "old"},
				newValue: &testStruct{ID: 5, Name: "new"},
				expected: &testStruct{ID: 5, Name: "new"},
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				result := pointer.UpdateSelective(tc.oldValue, tc.newValue)

				if tc.expected == nil {
					if result != nil {
						t.Errorf("expected nil, got %v", *result)
					}
					return
				}

				if result == nil {
					t.Errorf("expected %v, got nil", *tc.expected)
					return
				}

				if *result != *tc.expected {
					t.Errorf("expected %v, got %v", *tc.expected, *result)
				}
			})
		}
	})
}
