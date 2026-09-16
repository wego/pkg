package storage_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/wego/pkg/cognito/storage"
)

// These tests drive go-keyring's package-level mock, which is global state, so
// they must not run in parallel with anything else touching the keyring.

// TestNewAuto_UsesTheKeychainWhenAvailable is the path every workstation takes.
func TestNewAuto_UsesTheKeychainWhenAvailable(t *testing.T) {
	keyring.MockInit()
	t.Cleanup(keyring.MockInit)

	dir := t.TempDir()
	store, downgrade := storage.NewAuto("pay-admin-test", dir)

	require.NotNil(t, store)
	assert.Nil(t, downgrade, "a working keychain must not report a downgrade")

	require.NoError(t, store.Save("pay-admin-test/staging", fileTokens()))
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries, "with a keychain available nothing may be written to disk")
}

// TestNewAuto_FallsBackWhenTheSecretsServiceIsMissing reproduces the reported
// failure: a headless Linux host over SSH with no org.freedesktop.secrets on
// the bus. Before the fallback this discarded a completed sign-in.
func TestNewAuto_FallsBackWhenTheSecretsServiceIsMissing(t *testing.T) {
	missing := errors.New("The name org.freedesktop.secrets was not provided by any .service files")
	keyring.MockInitWithError(missing)
	t.Cleanup(keyring.MockInit)

	dir := filepath.Join(t.TempDir(), "sessions")
	store, downgrade := storage.NewAuto("pay-admin-test", dir)

	require.NotNil(t, downgrade, "an unreachable keychain must be reported, not hidden")
	assert.Equal(t, dir, downgrade.Dir)
	assert.ErrorIs(t, downgrade, missing, "the cause must stay unwrappable for callers that want it")
	assert.Contains(t, downgrade.Error(), "no usable OS keychain")

	// The point of the fallback: the session actually survives.
	require.NoError(t, store.Save("pay-admin-test/staging", fileTokens()))
	got, err := store.Load("pay-admin-test/staging")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "access-value", got.AccessToken)
}

// TestNewAuto_FallsBackOnAnyUnreadableKeychain covers the conservative rule: a
// keychain we cannot read is one we must not rely on for writes either, even
// when the error is not a recognisable D-Bus one.
func TestNewAuto_FallsBackOnAnyUnreadableKeychain(t *testing.T) {
	keyring.MockInitWithError(errors.New("some other keyring failure"))
	t.Cleanup(keyring.MockInit)

	_, downgrade := storage.NewAuto("pay-admin-test", t.TempDir())

	require.NotNil(t, downgrade)
}

// TestNewAuto_ProbeDoesNotWrite pins that availability is decided by a read.
// A probe that wrote would leave a stray entry in every operator's keychain.
func TestNewAuto_ProbeDoesNotWrite(t *testing.T) {
	keyring.MockInit()
	t.Cleanup(keyring.MockInit)

	_, downgrade := storage.NewAuto("pay-admin-test", t.TempDir())
	require.Nil(t, downgrade)

	// MockInit starts empty and only Set adds entries, so anything readable
	// here would have been written by the probe.
	_, err := keyring.Get("pay-admin-test", "__availability_probe__")
	assert.ErrorIs(t, err, keyring.ErrNotFound, "the probe must not leave an entry behind")
}
