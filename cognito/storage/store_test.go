package storage_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wego/pkg/cognito"
	"github.com/wego/pkg/cognito/storage"
)

func TestNewMemory_RoundTripsThroughThePublicAPI(t *testing.T) {
	store := storage.NewMemory()
	want := &cognito.TokenSet{
		AccessToken:  "access-value",
		IDToken:      "id-value",
		RefreshToken: "refresh-value",
		ExpiresAt:    time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	}

	before, err := store.Load("pay-admin/staging")
	require.NoError(t, err, "an empty store must report not-logged-in, not an error")
	require.Nil(t, before)

	require.NoError(t, store.Save("pay-admin/staging", want))

	got, err := store.Load("pay-admin/staging")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, *want, *got)

	require.NoError(t, store.Delete("pay-admin/staging"))

	after, err := store.Load("pay-admin/staging")
	require.NoError(t, err)
	assert.Nil(t, after)
}

// TestNewKeyring_ValidatesBeforeTouchingTheKeychain exercises the exported
// constructor without ever reaching the OS keychain: every case here is
// rejected by validation first, so no test can pop a keychain prompt.
func TestNewKeyring_ValidatesBeforeTouchingTheKeychain(t *testing.T) {
	tests := []struct {
		name           string
		givenService   string
		givenNamespace string
	}{
		{
			name:           "a blank service is rejected",
			givenService:   "",
			givenNamespace: "pay-admin/staging",
		},
		{
			name:           "a blank namespace is rejected",
			givenService:   "pay-admin",
			givenNamespace: "",
		},
		{
			name:           "both blank is rejected",
			givenService:   "",
			givenNamespace: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewKeyring(tt.givenService)
			require.NotNil(t, store)

			got, err := store.Load(tt.givenNamespace)
			require.Error(t, err)
			assert.Nil(t, got)

			require.Error(t, store.Save(tt.givenNamespace, &cognito.TokenSet{AccessToken: "access-value"}))
			require.Error(t, store.Delete(tt.givenNamespace))
		})
	}
}
