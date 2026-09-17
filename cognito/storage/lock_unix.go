//go:build unix

package storage

import (
	"fmt"
	"os"
	"syscall"
)

// lockPath is the advisory lock guarding one namespace's session file. It sits
// beside the session rather than covering the whole directory so that two
// environments never wait on each other.
func (f *fileStore) lockPath(namespace string) string {
	return f.path(namespace) + ".lock"
}

// lockNamespace takes an exclusive advisory lock for one namespace.
//
// Save commits with rename(2), which is atomic but says nothing about ordering
// against a concurrent Delete. Without this lock, logout can remove the session
// and report success while another process sits between writing its temp file
// and renaming it - and that rename then restores a live refresh token the
// operator has just been told was cleared. Two pay-admin invocations in two
// terminals is all it takes, and the warning this store prints tells operators
// to rely on logout precisely on the shared hosts where that is likeliest.
//
// flock is released by the kernel when the descriptor closes or the process
// dies, so a crash mid-write cannot strand it. The lock file is deliberately
// never removed: unlinking it while another process still holds a descriptor
// for it would leave the two of them holding locks on different inodes, which
// is the bug this is here to prevent.
func (f *fileStore) lockNamespace(namespace string) (func(), error) {
	file, err := os.OpenFile(f.lockPath(namespace), os.O_CREATE|os.O_RDWR, storeFileMode)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close()

		return nil, fmt.Errorf("lock the session for %q: %w", namespace, err)
	}

	return func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
	}, nil
}
