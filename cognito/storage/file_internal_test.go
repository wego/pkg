package storage

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// completeSession is the smallest body Load accepts.
const completeSession = `{"access_token":"a","id_token":"b","refresh_token":"c"}`

// TestFileStore_DeleteWaitsForAnInFlightSave reproduces the session
// resurrection: logout reports success, and a writer that was already inside
// Save then renames a live refresh token back into place.
//
// rename(2) is atomic but unordered against Delete, so the window is real
// whenever two pay-admin processes overlap - one refreshing an expired token
// while the operator signs out in another terminal. That is exactly the shared
// host this store's own warning tells people to run logout on.
func TestFileStore_DeleteWaitsForAnInFlightSave(t *testing.T) {
	dir := t.TempDir()
	store, ok := NewFile(dir).(*fileStore)
	require.True(t, ok)

	require.NoError(t, os.MkdirAll(dir, storeDirMode))

	// Stand in for the other process: inside Save, holding the namespace, not
	// yet renamed.
	release, err := store.lockNamespace(testNamespace)
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() { done <- store.Delete(testNamespace) }()

	select {
	case <-done:
		release()
		t.Fatal("Delete returned while another writer held the namespace: logout would report success, and the rename that writer is about to perform would restore the session")
	case <-time.After(100 * time.Millisecond):
	}

	// The in-flight Save commits its rename.
	require.NoError(t, os.WriteFile(store.path(testNamespace), []byte(completeSession), storeFileMode))
	release()

	select {
	case deleteErr := <-done:
		require.NoError(t, deleteErr)
	case <-time.After(5 * time.Second):
		t.Fatal("Delete never returned after the lock was released")
	}

	_, statErr := os.Stat(store.path(testNamespace))
	assert.True(t, errors.Is(statErr, fs.ErrNotExist),
		"once logout returns, no session may remain: a refresh token that survives it is the one thing logout promises to remove")
}

// TestFileStore_SaveWaitsForAnInFlightDelete is the other direction. Both
// operations have to take the lock or it serialises nothing - holding it in
// Delete alone would still let a Save commit straight through a logout.
func TestFileStore_SaveWaitsForAnInFlightDelete(t *testing.T) {
	dir := t.TempDir()
	store, ok := NewFile(dir).(*fileStore)
	require.True(t, ok)

	require.NoError(t, os.MkdirAll(dir, storeDirMode))

	release, err := store.lockNamespace(testNamespace)
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() { done <- store.Save(testNamespace, sampleTokens()) }()

	select {
	case <-done:
		release()
		t.Fatal("Save committed while another process held the namespace")
	case <-time.After(100 * time.Millisecond):
	}

	release()

	select {
	case saveErr := <-done:
		require.NoError(t, saveErr)
	case <-time.After(5 * time.Second):
		t.Fatal("Save never returned after the lock was released")
	}
}

// TestFileStore_DeleteWithoutADirectoryIsNotAnError covers the path the lock
// added: logging out of an environment that was never signed in to must not
// create the store directory just to take a lock on nothing.
func TestFileStore_DeleteWithoutADirectoryIsNotAnError(t *testing.T) {
	dir := t.TempDir() + "/never-created"
	store, ok := NewFile(dir).(*fileStore)
	require.True(t, ok)

	require.NoError(t, store.Delete(testNamespace))

	_, statErr := os.Stat(dir)
	assert.True(t, errors.Is(statErr, fs.ErrNotExist),
		"signing out of an environment that was never signed in to must leave no trace")
}
