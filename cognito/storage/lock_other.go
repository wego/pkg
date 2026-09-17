//go:build !unix

package storage

// lockNamespace fails closed off Unix.
//
// Save and Delete both take this lock, so returning an error here makes the
// file store refuse to write rather than write without serialisation. That is
// the right direction: neither of the store's guarantees survives off Unix -
// the owner-only mode is a Unix one that Go maps onto Windows ACLs only
// approximately, and flock has no portable equivalent here - and a store that
// cannot order Save against Delete can report a successful sign-out and then
// have a concurrent write restore the session.
//
// NewAuto never reaches this code: it drops the file store entirely on a
// platform its allow-list does not cover, as the selected backend and as the
// one Delete also clears. This guards the other door, a caller using NewFile
// directly, which is exported and otherwise has nothing stopping it.
func (f *fileStore) lockNamespace(string) (func(), error) {
	return nil, errFileStoreUnsupported
}
