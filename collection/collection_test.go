package collection_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/collection"
)

func Test_Dedup(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "string empty",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "string no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "string duplicates",
			input:    []string{"a", "b", "c", "a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, collection.Dedup(test.input))
		})
	}

	for _, test := range []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "int empty",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "int no duplicates",
			input:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "int duplicates",
			input:    []int{1, 2, 3, 1, 2, 3},
			expected: []int{1, 2, 3},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, collection.Dedup(test.input))
		})
	}

	for _, test := range []struct {
		name     string
		input    []float64
		expected []float64
	}{
		{
			name:     "float64 empty",
			input:    []float64{},
			expected: []float64{},
		},
		{
			name:     "float64 no duplicates",
			input:    []float64{1.1, 2.2, 3.3},
			expected: []float64{1.1, 2.2, 3.3},
		},
		{
			name:     "float64 duplicates",
			input:    []float64{1.1, 2.2, 3.3, 1.1, 2.2, 3.3},
			expected: []float64{1.1, 2.2, 3.3},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, collection.Dedup(test.input))
		})
	}

	for _, test := range []struct {
		name     string
		input    []bool
		expected []bool
	}{
		{
			name:     "bool empty",
			input:    []bool{},
			expected: []bool{},
		},
		{
			name:     "bool duplicates",
			input:    []bool{true, false},
			expected: []bool{true, false},
		},
		{
			name:     "bool duplicates",
			input:    []bool{true, false, true, true, false, true},
			expected: []bool{true, false},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, collection.Dedup(test.input))
		})
	}

	type address struct {
		City       string
		State      string
		PostalCode string
	}

	type comparableStruct struct {
		name    string
		age     int
		address address
	}

	for _, test := range []struct {
		name     string
		input    []comparableStruct
		expected []comparableStruct
	}{
		{
			name:     "struct empty",
			input:    []comparableStruct{},
			expected: []comparableStruct{},
		},
		{
			name:     "struct no duplicates",
			input:    []comparableStruct{{"a", 1, address{"a", "a", "a"}}, {"b", 2, address{"b", "b", "b"}}, {"c", 3, address{"c", "c", "c"}}},
			expected: []comparableStruct{{"a", 1, address{"a", "a", "a"}}, {"b", 2, address{"b", "b", "b"}}, {"c", 3, address{"c", "c", "c"}}},
		},
		{
			name:     "struct duplicates",
			input:    []comparableStruct{{"a", 1, address{"a", "a", "a"}}, {"b", 2, address{"b", "b", "b"}}, {"c", 3, address{"c", "c", "c"}}, {"a", 1, address{"a", "a", "a"}}, {"b", 2, address{"b", "b", "b"}}, {"c", 3, address{"c", "c", "c"}}},
			expected: []comparableStruct{{"a", 1, address{"a", "a", "a"}}, {"b", 2, address{"b", "b", "b"}}, {"c", 3, address{"c", "c", "c"}}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, collection.Dedup(test.input))
		})
	}
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
	for _, test := range []struct {
		name     string
		mapper   func(string) string
		input    []string
		expected []string
	}{
		{
			name:     "to upper",
			mapper:   strings.ToUpper,
			input:    []string{"a", "b", "c"},
			expected: []string{"A", "B", "C"},
		},
		{
			name:     "to lower",
			mapper:   strings.ToLower,
			input:    []string{"A", "B", "C"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "to title",
			mapper:   strings.ToTitle,
			input:    []string{"abc", "你好, 世界", "Hello World"},
			expected: []string{"ABC", "你好, 世界", "HELLO WORLD"},
		},
		{
			name:     "trim",
			mapper:   strings.TrimSpace,
			input:    []string{" a ", " b ", " c "},
			expected: []string{"a", "b", "c"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			result := collection.Map(test.input, test.mapper)
			assertions.Equal(test.expected, result)
		})
	}

	for _, test := range []struct {
		name     string
		mapper   func(string) int
		input    []string
		expected []int
	}{
		{
			name:     "len",
			mapper:   func(s string) int { return len(s) },
			input:    []string{"abc", "你好, 世界", "Hello World"},
			expected: []int{3, 14, 11},
		},
		{
			name:     "atoi",
			mapper:   func(s string) int { i, _ := strconv.Atoi(s); return i },
			input:    []string{"1", "2", "3"},
			expected: []int{1, 2, 3},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			result := collection.Map(test.input, test.mapper)
			assertions.Equal(test.expected, result)
		})
	}

	type personal struct {
		name string
		age  int
	}

	for _, test := range []struct {
		name     string
		mapper   func(personal) int
		input    []personal
		expected []int
	}{
		{
			name:     "Age",
			mapper:   func(p personal) int { return p.age },
			input:    []personal{{"James Bond", 32}, {"John Doe", 42}, {"Jack Bauer", 52}},
			expected: []int{32, 42, 52},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			result := collection.Map(test.input, test.mapper)
			assertions.Equal(test.expected, result)
		})
	}
}

func Test_MapI(t *testing.T) {
	assertions := assert.New(t)
	resultYes := collection.MapI(strs, func(s string) interface{} {
		return strings.ToUpper(s)
	})
	assertions.Equal(resultYes[0], "PEACH")
	assertions.Equal(resultYes[1], "APPLE")
}

func Test_Equal(t *testing.T) {

	type testStruct[T any] struct {
		name string
		a    []T
		b    []T
		want bool
	}

	for _, test := range []testStruct[string]{
		{
			name: "equal",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "c"},
			want: true,
		},
		{
			name: "not equal",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "d"},
			want: false,
		},
		{
			name: "not equal",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b"},
			want: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.Equal(test.want, collection.Equals(test.a, test.b))
		})
	}

	for _, test := range []testStruct[int]{
		{
			name: "equal",
			a:    []int{1, 2, 3},
			b:    []int{1, 2, 3},
			want: true,
		},
		{
			name: "not equal",
			a:    []int{1, 2, 3},
			b:    []int{1, 2, 4},
			want: false,
		},
		{
			name: "not equal",
			a:    []int{1, 2, 3},
			b:    []int{1, 2},
			want: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.Equal(test.want, collection.Equals(test.a, test.b))
		})
	}

	for _, test := range []testStruct[float64]{
		{
			name: "equal",
			a:    []float64{1.1, 2.2, 3.3},
			b:    []float64{1.1, 2.2, 3.3},
			want: true,
		},
		{
			name: "not equal",
			a:    []float64{1.1, 2.2, 3.3},
			b:    []float64{1.1, 2.2, 4.4},
			want: false,
		},
		{
			name: "not equal",
			a:    []float64{1.1, 2.2, 3.3},
			b:    []float64{1.1, 2.2},
			want: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.Equal(test.want, collection.Equals(test.a, test.b))
		})
	}

	for _, test := range []testStruct[struct {
		F1 string
		F2 int
	}]{
		{
			name: "equal",
			a: []struct {
				F1 string
				F2 int
			}{{"a", 1}, {"b", 2}, {"c", 3}},
			b: []struct {
				F1 string
				F2 int
			}{{"a", 1}, {"b", 2}, {"c", 3}},
			want: true,
		},
		{
			name: "not equal",
			a: []struct {
				F1 string
				F2 int
			}{{"a", 1}, {"b", 2}, {"c", 3}},
			b: []struct {
				F1 string
				F2 int
			}{{"a", 1}, {"b", 2}, {"d", 3}},
		},
		{
			name: "not equal",
			a: []struct {
				F1 string
				F2 int
			}{{"a", 1}, {"b", 2}, {"c", 3}},
			b: []struct {
				F1 string
				F2 int
			}{{"a", 1}, {"b", 2}},
			want: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.Equal(test.want, collection.Equals(test.a, test.b))
		})
	}
}

func Test_Contains(t *testing.T) {
	type testStruct[K, V comparable] struct {
		name string
		m    map[K]V
		keys []K
		vals []V
		want bool
	}

	// ContainsKeys
	for _, test := range []testStruct[string, string]{
		{
			name: "ContainsKeys - contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			keys: []string{"a", "b"},
			want: true,
		},
		{
			name: "ContainsKeys - not contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			keys: []string{"a", "d"},
			want: false,
		},
		{
			name: "ContainsKeys - not contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			keys: []string{"a", "b", "d"},
			want: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.Equal(test.want, collection.ContainsKeys(test.m, test.keys))
		})
	}

	// ContainsAnyKeys
	for _, test := range []testStruct[string, string]{
		{
			name: "ContainsAnyKeys - contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			keys: []string{"a", "b"},
			want: true,
		},
		{
			name: "ContainsAnyKeys - contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			keys: []string{"a", "d"},
			want: true,
		},
		{
			name: "ContainsAnyKeys - not contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			keys: []string{"d", "e"},
			want: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.Equal(test.want, collection.ContainsAnyKeys(test.m, test.keys))
		})
	}

	// ContainsNoneKeys
	for _, test := range []testStruct[string, string]{
		{
			name: "ContainsNoneKeys - not contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			keys: []string{"a", "b"},
			want: false,
		},
		{
			name: "ContainsNoneKeys - not contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			keys: []string{"a", "d"},
			want: false,
		},
		{
			name: "ContainsNoneKeys - contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			keys: []string{"d", "e"},
			want: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.Equal(test.want, collection.ContainsNoneKeys(test.m, test.keys))
		})
	}

	// ContainsValues
	for _, test := range []testStruct[string, string]{
		{
			name: "ContainsValues - contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			vals: []string{"a", "b"},
			want: true,
		},
		{
			name: "ContainsValues - not contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			vals: []string{"a", "d"},
			want: false,
		},
		{
			name: "ContainsValues - not contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			vals: []string{"a", "b", "d"},
			want: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.Equal(test.want, collection.ContainsValues(test.m, test.vals))
		})
	}

	// ContainsAnyValues
	for _, test := range []testStruct[string, string]{
		{
			name: "ContainsAnyValues - contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			vals: []string{"a", "b"},
			want: true,
		},
		{
			name: "ContainsAnyValues - contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			vals: []string{"a", "d"},
			want: true,
		},
		{
			name: "ContainsAnyValues - not contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			vals: []string{"d", "e"},
			want: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.Equal(test.want, collection.ContainsAnyValues(test.m, test.vals))
		})
	}

	// ContainsNoneValues
	for _, test := range []testStruct[string, string]{
		{
			name: "ContainsNoneValues - not contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			vals: []string{"a", "b"},
			want: false,
		},
		{
			name: "ContainsNoneValues - not contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			vals: []string{"a", "d"},
			want: false,
		},
		{
			name: "ContainsNoneValues - contains",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			vals: []string{"d", "e"},
			want: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.Equal(test.want, collection.ContainsNoneValues(test.m, test.vals))
		})
	}

	// Keys
	for _, test := range []testStruct[string, string]{
		{
			name: "Keys - empty",
			m:    map[string]string{},
			keys: []string{},
		},
		{
			name: "Keys - not empty",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			keys: []string{"a", "b", "c"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.True(collection.Equals(test.keys, collection.Keys(test.m)))
		})
	}

	for _, test := range []testStruct[int, string]{
		{
			name: "Keys - empty",
			m:    map[int]string{},
			keys: []int{},
		},
		{
			name: "Keys - not empty",
			m:    map[int]string{1: "a", 2: "b", 3: "c"},
			keys: []int{1, 2, 3},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.True(collection.Equals(test.keys, collection.Keys(test.m)))
		})
	}

	// Values
	for _, test := range []testStruct[string, string]{
		{
			name: "Values - empty",
			m:    map[string]string{},
			vals: []string{},
		},
		{
			name: "Values - not empty",
			m:    map[string]string{"a": "a", "b": "b", "c": "c"},
			vals: []string{"a", "b", "c"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.True(collection.Equals(test.vals, collection.Values(test.m)))
		})
	}

	for _, test := range []testStruct[string, float64]{
		{
			name: "Values - empty",
			m:    map[string]float64{},
			vals: []float64{},
		},
		{
			name: "Values - not empty",
			m:    map[string]float64{"a": 1.1, "b": 2.2, "c": 3.3},
			vals: []float64{1.1, 2.2, 3.3},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			assertions.True(collection.Equals(test.vals, collection.Values(test.m)))
		})
	}
}

func TestCopy(t *testing.T) {
	for _, test := range []struct {
		name     string
		dst      map[string]any
		src      map[string]any
		expected map[string]any
	}{
		{
			name: "maps with different keys",
			dst: map[string]any{
				"a": "a",
			},
			src: map[string]any{
				"b": "b",
			},
			expected: map[string]any{
				"a": "a",
				"b": "b",
			},
		},
		{
			name: "maps with same keys",
			dst: map[string]any{
				"a": "a",
			},
			src: map[string]any{
				"a": "b",
			},
			expected: map[string]any{
				"a": "b",
			},
		},
		{
			name: "dst map with no keys",
			dst:  map[string]any{},
			src: map[string]any{
				"a": "a",
			},
			expected: map[string]any{
				"a": "a",
			},
		},
		{
			name: "src map with no keys",
			dst: map[string]any{
				"a": "a",
			},
			src: map[string]any{},
			expected: map[string]any{
				"a": "a",
			},
		},
		{
			name: "nil dst map",
			dst:  nil,
			src: map[string]any{
				"a": "a",
			},
			expected: nil,
		},
		{
			name: "nil src map",
			dst: map[string]any{
				"a": "a",
			},
			src: nil,
			expected: map[string]any{
				"a": "a",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			collection.Copy(test.dst, test.src)
			assert.Equal(t, test.expected, test.dst)
		})
	}
}
