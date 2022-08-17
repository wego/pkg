package collection

// EqualValuesI returns true if the two slices are values equal ignoring the order
func EqualValuesI(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	// create a map of []interface{} -> int
	diff := make(map[interface{}]int, len(a))
	for _, aVal := range a {
		// 0 value for int is 0, so just increment a counter for the value
		diff[aVal]++
	}
	for _, bVal := range b {
		// If the count is not in diff bail out early
		if _, ok := diff[bVal]; !ok {
			return false
		}
		diff[bVal]--
		if diff[bVal] == 0 {
			delete(diff, bVal)
		}
	}
	return len(diff) == 0
}
