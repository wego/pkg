package to_test

import (
	"fmt"
	"github.com/wego/pkg/pointer"
	"github.com/wego/pkg/to"
	"testing"
	"unsafe"
)

func testFuncNameXXX(int, string) {
}

func TestTo(t *testing.T) {
	var (
		intChan = make(chan int, 1024)
	)
	testCases := map[string]struct {
		value    any
		expected string
	}{
		"Bool_True":  {true, "true"},
		"Bool_False": {false, "false"},
		"Int":        {value: 14, expected: "14"},
		"Int8":       {value: int8(15), expected: "15"},
		"Int16":      {value: int16(16), expected: "16"},
		"Int32":      {value: int32(17), expected: "17"},
		"Int64":      {value: int64(18), expected: "18"},
		"Uint":       {value: uint(19), expected: "19"},
		"Uint8":      {value: uint8(20), expected: "20"},
		"Uint16":     {value: uint16(21), expected: "21"},
		"Uint32":     {value: uint32(22), expected: "22"},
		"Uint64":     {value: uint64(23), expected: "23"},
		"Uintptr":    {value: uintptr(24), expected: "24"},
		"Float32":    {value: 25.30, expected: "25.3"},
		"Float64":    {value: 16.30, expected: "16.3"},
		"Complex64":  {value: complex(1, 2), expected: "1+2i"},
		"Complex128": {value: complex(3, 4), expected: "3+4i"},
		"String":     {value: "17", expected: "17"},
		"Struct": {value: struct {
			A int `json:"a"`
		}{1},
			expected: `{"a":1}`,
		},
		"StructWithPointer": {
			value: struct {
				A *int     `json:"a"`
				B *float64 `json:"b"`
			}{A: pointer.To(1), B: pointer.To(2.3)}, expected: `{"a":1,"b":2.3}`,
		},
		"Slice": {value: []int{1, 2, 3}, expected: "[1,2,3]"},
		"Array": {value: [3]string{"A", "B", "C"}, expected: `["A","B","C"]`},
		"Map":   {value: map[string]int{"a": 1, "b": 2}, expected: `{"a":1,"b":2}`},
		"StructSlice": {
			value: []struct {
				A int `json:"a"`
			}{{1}, {2}, {3}},
			expected: `[{"a":1},{"a":2},{"a":3}]`,
		},
		"Nil":  {value: nil, expected: "null"},
		"Func": {value: testFuncNameXXX, expected: fmt.Sprintf("func(int, string) at %p", testFuncNameXXX)},
		"Chan": {value: intChan, expected: fmt.Sprintf("chan int at %p with 0 elements", intChan)},
		"UnsafePointer": {
			value:    unsafe.Pointer(&intChan),
			expected: fmt.Sprintf("unsafe.Pointer at %p", unsafe.Pointer(&intChan)),
		},
		"Pointer": {
			value: &struct {
				A int `json:"a"`
			}{},
			expected: `{"a":0}`,
		},
		"PointerNil": {
			value:    (*int)(nil),
			expected: "nil",
		},
		"Interface": {
			value: interface{}(&struct {
				A int `json:"a"`
			}{1}),
			expected: `{"a":1}`,
		},
		"InterfaceSlice": {
			value:    []interface{}{1, 2, 3},
			expected: `[1,2,3]`,
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			result := to.String(tc.value)
			if result != tc.expected {
				t.Fatalf("fail to get string from value %+v of type %T, expected: %s, actual: %s", tc.value, tc.value, tc.expected, result)
			}
		})
	}
}
