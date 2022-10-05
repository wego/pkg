package collection

// Dedup returns the input slice after removing duplicated items
func Dedup[T comparable](in []T) (out []T) {
	keys := make(map[T]bool)
	out = make([]T, 0)
	for _, v := range in {
		if found, _ := keys[v]; !found {
			keys[v] = true
			out = append(out, v)
		}
	}
	return
}

// IndexOf returns the first index of the target element t, or -1 if no match is found
func IndexOf[T comparable](vs []T, t T) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// Contains returns true if the target element t is in the slice.
func Contains[T comparable](vs []T, t T) bool {
	return IndexOf(vs, t) >= 0
}

// Any returns true if one of the element in the slice satisfies the predicate f
func Any[T comparable](vs []T, f func(T) bool) bool {
	for _, v := range vs {
		if f(v) {
			return true
		}
	}
	return false
}

// All returns true if all the elements in the slice satisfy the predicate f
func All[T comparable](vs []T, f func(T) bool) bool {
	for _, v := range vs {
		if !f(v) {
			return false
		}
	}
	return true
}

// Filter returns a new slice containing all in the slice that satisfy the predicate f
func Filter[T comparable](vs []T, f func(T) bool) []T {
	vsf := make([]T, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// Map returns a new slice containing the results of applying the function f to each element in the original slice
func Map[I, R comparable](vs []I, f func(I) R) []R {
	vsm := make([]R, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

// MapI returns a new slice containing the results of applying the function f to each element in the original slice
// FIXME: will deprecate this function in the future
func MapI[T comparable](vs []T, f func(T) interface{}) []interface{} {
	vsm := make([]interface{}, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

// Distinct remove the duplicated items from the slice
func Distinct[T comparable](in []T) (out []T) {
	return Dedup(in)
}

// Equal checks whether 2 slices have the same items
func Equal[T comparable](s1 []T, s2 []T) bool {
	if len(s1) != len(s2) {
		return false
	}

	s1Map := make(map[T]bool)
	for _, v := range s1 {
		s1Map[v] = true
	}

	for _, v := range s2 {
		if !s1Map[v] {
			return false
		}
	}

	return true
}
