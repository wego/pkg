package storage_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wego/pkg/cognito"
	"github.com/wego/pkg/cognito/storage"
)

func fileTokens() *cognito.TokenSet {
	return &cognito.TokenSet{
		AccessToken:  "access-value",
		IDToken:      "id-value",
		RefreshToken: "refresh-value",
		ExpiresAt:    time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC),
	}
}

func TestFileStore_RoundTrip(t *testing.T) {
	store := storage.NewFile(t.TempDir())
	want := fileTokens()

	require.NoError(t, store.Save("pay-admin/staging", want))

	got, err := store.Load("pay-admin/staging")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, want.AccessToken, got.AccessToken)
	assert.Equal(t, want.IDToken, got.IDToken)
	assert.Equal(t, want.RefreshToken, got.RefreshToken)
	assert.True(t, want.ExpiresAt.Equal(got.ExpiresAt), "expiry must survive the round trip")
}

// TestFileStore_NothingStoredIsNotAnError pins the contract shared with the
// keyring store: not signed in is a normal state, not a failure.
func TestFileStore_NothingStoredIsNotAnError(t *testing.T) {
	got, err := storage.NewFile(t.TempDir()).Load("pay-admin/staging")

	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestFileStore_IsOwnerOnly is the whole security story of this store. The
// refresh token is long-lived and sits in plaintext, so the file mode is the
// only thing protecting it.
func TestFileStore_IsOwnerOnly(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, storage.NewFile(dir).Save("pay-admin/production", fileTokens()))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1, "exactly one session file, and no temp file left behind")

	info, err := entries[0].Info()
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "session file must not be readable by group or other")

	dirInfo, err := os.Stat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm(), "session directory must not be listable by group or other")
}

// TestFileStore_NamespacesAreIsolated covers the reason namespaces exist: a
// staging login must never overwrite a production one.
func TestFileStore_NamespacesAreIsolated(t *testing.T) {
	store := storage.NewFile(t.TempDir())

	staging := fileTokens()
	staging.AccessToken = "staging-access"
	production := fileTokens()
	production.AccessToken = "production-access"

	require.NoError(t, store.Save("pay-admin/staging", staging))
	require.NoError(t, store.Save("pay-admin/production", production))

	gotStaging, err := store.Load("pay-admin/staging")
	require.NoError(t, err)
	assert.Equal(t, "staging-access", gotStaging.AccessToken)

	gotProduction, err := store.Load("pay-admin/production")
	require.NoError(t, err)
	assert.Equal(t, "production-access", gotProduction.AccessToken)
}

// TestFileStore_NamespaceCannotEscapeTheDirectory is the reason the filename is
// derived rather than used directly. A namespace is opaque caller input, so a
// traversal sequence in it must not reach outside dir.
func TestFileStore_NamespaceCannotEscapeTheDirectory(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "sessions")
	outside := filepath.Join(root, "escaped")

	require.NoError(t, storage.NewFile(dir).Save("../escaped", fileTokens()))

	_, err := os.Stat(outside)
	assert.True(t, os.IsNotExist(err), "a traversing namespace must not write outside the store directory")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "it must land inside the store directory instead")
}

func TestFileStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFile(dir)
	require.NoError(t, store.Save("pay-admin/staging", fileTokens()))

	require.NoError(t, store.Delete("pay-admin/staging"))

	got, err := store.Load("pay-admin/staging")
	require.NoError(t, err)
	assert.Nil(t, got, "a deleted session reads as not signed in")

	assert.NoError(t, store.Delete("pay-admin/staging"), "deleting nothing is not an error")
}

// TestFileStore_RejectsAnIncompleteSession covers the same rule the keyring
// store applies: a session missing a token is an inconsistency worth reporting,
// not a logged-out state to swallow.
func TestFileStore_RejectsAnIncompleteSession(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFile(dir)
	require.NoError(t, store.Save("pay-admin/staging", fileTokens()))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	path := filepath.Join(dir, entries[0].Name())
	require.NoError(t, os.WriteFile(path, []byte(`{"access_token":"a","id_token":"b"}`), 0o600))

	_, err = store.Load("pay-admin/staging")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "incomplete")
	assert.Contains(t, err.Error(), "sign in again", "the error must name the recovery")
}

func TestFileStore_RejectsUnreadableContent(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFile(dir)
	require.NoError(t, store.Save("pay-admin/staging", fileTokens()))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, entries[0].Name()), []byte("not json"), 0o600))

	_, err = store.Load("pay-admin/staging")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "sign in again", "the error must name the recovery")
}

// TestFileStore_SaveReplacesAtomically pins that a second Save leaves exactly
// one file and no temp debris, which is what makes a crash mid-write safe.
func TestFileStore_SaveReplacesAtomically(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFile(dir)

	first := fileTokens()
	first.AccessToken = "first"
	require.NoError(t, store.Save("pay-admin/staging", first))

	second := fileTokens()
	second.AccessToken = "second"
	require.NoError(t, store.Save("pay-admin/staging", second))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "replacing a session must not leave the old file or a temp file behind")

	got, err := store.Load("pay-admin/staging")
	require.NoError(t, err)
	assert.Equal(t, "second", got.AccessToken)
}

func TestFileStore_RejectsBlankNamespace(t *testing.T) {
	store := storage.NewFile(t.TempDir())

	_, loadErr := store.Load("")
	assert.Error(t, loadErr)
	assert.Error(t, store.Save("", fileTokens()))
	assert.Error(t, store.Delete(""))
}

func TestFileStore_RejectsNilTokens(t *testing.T) {
	assert.Error(t, storage.NewFile(t.TempDir()).Save("pay-admin/staging", nil))
}

func TestDefaultFileDir_FollowsXDG(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/xdg-state")

	dir, err := storage.DefaultFileDir("pay-admin")

	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/tmp/xdg-state", "pay-admin"), dir)
}

func TestDefaultFileDir_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")

	dir, err := storage.DefaultFileDir("pay-admin")

	require.NoError(t, err)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".local", "state", "pay-admin"), dir)
}
