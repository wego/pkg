package collection

// Chunk splits a slice into chunks of the given size.
// The last chunk may be smaller than the given size.
// Reference: https://github.com/golang/go/wiki/SliceTricks#batching-with-minimal-allocation
func Chunk[T any](slice []T, size int) (chunks [][]T) {
	if len(slice) == 0 || size < 1 {
		return
	}

	chunks = make([][]T, 0, (len(slice)+size-1)/size)

	for size < len(slice) {
		slice, chunks = slice[size:], append(chunks, slice[0:size:size])
	}

	chunks = append(chunks, slice)
	return chunks
}
