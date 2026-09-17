//go:build !unix

package storage

// lockNamespace does nothing off Unix.
//
// The file store is a Unix-only fallback - its owner-only guarantee is a Unix
// one, and NewAuto never selects it on Windows, where go-keyring drives a
// working credential manager. Keeping the method here lets the rest of the
// package stay platform-independent rather than growing build tags of its own.
func (f *fileStore) lockNamespace(string) (func(), error) {
	return func() {}, nil
}
