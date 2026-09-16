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

// TestNewAuto_DeleteClearsBothBackends is the regression for a session that
// outlived the condition which chose where it was stored.
//
// Sign in while the keychain is down, so the session goes to the file. Sign out
// after it comes back, when NewAuto selects the keychain. Then lose the
// keychain again. Before Delete cleared both backends the file survived, and
// Load handed back the session the operator believed they had deleted, live
// refresh token and all.
func TestNewAuto_DeleteClearsBothBackends(t *testing.T) {
	dir := t.TempDir()
	missing := errors.New("The name org.freedesktop.secrets was not provided by any .service files")

	keyring.MockInitWithError(missing)
	whileDown, downgrade := storage.NewAuto("pay-admin-test", dir)
	require.NotNil(t, downgrade, "the keychain is unavailable for this step")
	require.NoError(t, whileDown.Save("pay-admin-test/staging", fileTokens()))

	// The keychain returns, so this process selects it, and the operator signs out.
	keyring.MockInit()
	whileUp, noDowngrade := storage.NewAuto("pay-admin-test", dir)
	require.Nil(t, noDowngrade, "the keychain is available for this step")
	require.NoError(t, whileUp.Delete("pay-admin-test/staging"))

	// The keychain goes away again and the file store is selected once more.
	keyring.MockInitWithError(missing)
	t.Cleanup(keyring.MockInit)
	whileDownAgain, _ := storage.NewAuto("pay-admin-test", dir)

	got, err := whileDownAgain.Load("pay-admin-test/staging")

	require.NoError(t, err)
	assert.Nil(t, got, "a session deleted while signed out must not come back when the keychain disappears")
}

// TestNewAuto_DeleteClearsTheKeychainFromAFileHost is the mirror image: the
// session was written on a host with a keychain, and the sign-out happens while
// the keychain is unreachable. Both directions have to clear both backends, or
// sign-out only means "signed out on whichever backend happens to be selected".
func TestNewAuto_DeleteClearsTheKeychainFromAFileHost(t *testing.T) {
	dir := t.TempDir()

	keyring.MockInit()
	whileUp, noDowngrade := storage.NewAuto("pay-admin-test", dir)
	require.Nil(t, noDowngrade)
	require.NoError(t, whileUp.Save("pay-admin-test/staging", fileTokens()))

	// Sign out while the keychain is unreachable: the file store is selected,
	// but the session lives in the keychain.
	keyring.MockInitWithError(errors.New("dbus: couldn't determine address of session bus"))
	whileDown, downgrade := storage.NewAuto("pay-admin-test", dir)
	require.NotNil(t, downgrade)
	// The keychain delete cannot succeed here, so the caller is told - but the
	// attempt still has to be made rather than skipped.
	_ = whileDown.Delete("pay-admin-test/staging")

	keyring.MockInit()
	t.Cleanup(keyring.MockInit)
	whileUpAgain, _ := storage.NewAuto("pay-admin-test", dir)

	got, err := whileUpAgain.Load("pay-admin-test/staging")

	require.NoError(t, err)
	assert.Nil(t, got, "the keychain entry must be cleared even when sign-out ran on a file-backed host")
}

// TestNewAuto_SaveWritesOnlyTheSelectedBackend pins that the fan-out is limited
// to Delete. Writing both would put a long-lived refresh token in the weaker
// store on hosts that have a keychain, which is the opposite of the intent.
func TestNewAuto_SaveWritesOnlyTheSelectedBackend(t *testing.T) {
	keyring.MockInit()
	t.Cleanup(keyring.MockInit)

	dir := t.TempDir()
	store, downgrade := storage.NewAuto("pay-admin-test", dir)
	require.Nil(t, downgrade)

	require.NoError(t, store.Save("pay-admin-test/staging", fileTokens()))

	entries, err := os.ReadDir(dir)
	if !os.IsNotExist(err) {
		require.NoError(t, err)
		assert.Empty(t, entries, "a host with a keychain must not also write the session to disk")
	}
}
