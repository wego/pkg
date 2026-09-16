package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wego/pkg/cognito"
)

// recordingStore counts what it was asked to do, so a test can assert that
// Delete reached a backend rather than inferring it from state.
type recordingStore struct {
	deletes   int
	deleteErr error
	saves     int
	loads     int
}

func (r *recordingStore) Load(string) (*cognito.TokenSet, error) { r.loads++; return nil, nil }
func (r *recordingStore) Save(string, *cognito.TokenSet) error   { r.saves++; return nil }
func (r *recordingStore) Delete(string) error                    { r.deletes++; return r.deleteErr }

// TestAutoStore_DeleteReachesBothBackends asserts the fan-out directly.
//
// This is an internal test with injected fakes because the external one it
// replaces could not prove anything: go-keyring's MockInit and
// MockInitWithError each install a FRESH provider, so toggling the keychain
// error to simulate an availability change also discarded every stored entry.
// The final assertion then held whether or not Delete had fanned out - verified
// by removing the fan-out and watching it still pass. Counting the calls is
// immune to that.
func TestAutoStore_DeleteReachesBothBackends(t *testing.T) {
	primary, secondary := &recordingStore{}, &recordingStore{}
	store := &autoStore{primary: primary, secondary: secondary}

	require.NoError(t, store.Delete("pay-admin/staging"))

	assert.Equal(t, 1, primary.deletes, "the selected backend must be cleared")
	assert.Equal(t, 1, secondary.deletes, "and so must the one that is not selected")
}

// TestAutoStore_DeleteAttemptsBothEvenWhenTheFirstFails is the case the
// fan-out exists for: a half-done sign-out leaves a live refresh token behind
// in whichever store was skipped.
func TestAutoStore_DeleteAttemptsBothEvenWhenTheFirstFails(t *testing.T) {
	primary := &recordingStore{deleteErr: errors.New("keychain locked")}
	secondary := &recordingStore{}
	store := &autoStore{primary: primary, secondary: secondary}

	err := store.Delete("pay-admin/staging")

	require.Error(t, err, "the caller hears about the backend they are working with")
	assert.Contains(t, err.Error(), "keychain locked")
	assert.Equal(t, 1, secondary.deletes, "the other backend is still attempted")
}

// TestAutoStore_DeleteReportsTheSecondaryOnlyWhenThePrimarySucceeded keeps the
// more useful error when both fail: the caller is going to retry against the
// primary anyway.
func TestAutoStore_DeleteReportsTheSecondaryOnlyWhenThePrimarySucceeded(t *testing.T) {
	t.Run("secondary failure surfaces when the primary worked", func(t *testing.T) {
		store := &autoStore{
			primary:   &recordingStore{},
			secondary: &recordingStore{deleteErr: errors.New("no dbus")},
		}

		err := store.Delete("pay-admin/staging")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "the session was cleared")
		assert.Contains(t, err.Error(), "no dbus")
	})

	t.Run("primary failure wins when both fail", func(t *testing.T) {
		store := &autoStore{
			primary:   &recordingStore{deleteErr: errors.New("primary boom")},
			secondary: &recordingStore{deleteErr: errors.New("secondary boom")},
		}

		err := store.Delete("pay-admin/staging")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "primary boom")
		assert.NotContains(t, err.Error(), "secondary boom")
	})
}

// TestAutoStore_LoadAndSaveStayOnThePrimary pins that the fan-out is limited to
// Delete. Writing both would put a long-lived refresh token in the weaker store
// even on hosts that have a keychain.
func TestAutoStore_LoadAndSaveStayOnThePrimary(t *testing.T) {
	primary, secondary := &recordingStore{}, &recordingStore{}
	store := &autoStore{primary: primary, secondary: secondary}

	_, err := store.Load("pay-admin/staging")
	require.NoError(t, err)
	require.NoError(t, store.Save("pay-admin/staging", &cognito.TokenSet{}))

	assert.Equal(t, 1, primary.loads)
	assert.Equal(t, 1, primary.saves)
	assert.Zero(t, secondary.loads, "reading both would make precedence ambiguous")
	assert.Zero(t, secondary.saves, "writing both would put the credential in the weaker store")
}

// TestFileStoreProtects covers the platform allow-list: the file store's
// owner-only guarantee is a Unix one, so NewAuto must not downgrade to it
// anywhere that guarantee has not been verified.
func TestFileStoreProtects(t *testing.T) {
	// Whatever this suite runs on is a platform we support, so the live answer
	// must be true - if it is not, the fallback is silently disabled here.
	assert.True(t, fileStoreProtects(), "the development platforms must support the file store")
}
