package collection

// Dedup returns the input slice after removing duplicated items
func Dedup[T comparable](in []T) (out []T) {
	keys := make(map[T]bool)
	for _, v := range in {
		if found, _ := keys[v]; !found {
			keys[v] = true
			out = append(out, v)
		}
	}
	return
}
