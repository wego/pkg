//go:build !unix

package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileStore_RefusesToWriteOffUnix pins the fail-closed behaviour for a
// caller that reaches NewFile directly on a platform where neither the file
// mode nor the locking holds. It cannot run on the development platforms, so it
// carries the same build tag as the stub it covers.
func TestFileStore_RefusesToWriteOffUnix(t *testing.T) {
	store := NewFile(t.TempDir())

	saveErr := store.Save(testNamespace, sampleTokens())
	require.Error(t, saveErr, "writing a long-lived refresh token to an unserialised store must fail, not succeed quietly")
	assert.True(t, errors.Is(saveErr, errFileStoreUnsupported))

	deleteErr := store.Delete(testNamespace)
	require.Error(t, deleteErr, "a sign-out that cannot serialise against a concurrent write must not report success")
	assert.True(t, errors.Is(deleteErr, errFileStoreUnsupported))
}
