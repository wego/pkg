package storage

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wegostrings "github.com/wego/pkg/strings"
	"github.com/zalando/go-keyring"

	"github.com/wego/pkg/cognito"
)

const (
	testService   = "pay-admin-test"
	testNamespace = "pay-admin/staging"
)

var testExpiry = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

// TestStoreContract holds every Store implementation to the same behaviour.
// The keyring case runs against a fake backend, so no test ever pops an OS
// keychain prompt.
func TestStoreContract(t *testing.T) {
	impls := []struct {
		name  string
		build func() Store
	}{
		{
			name:  "memory",
			build: NewMemory,
		},
		{
			name: "keyring",
			build: func() Store {
				return &keyringStore{service: testService, backend: newFakeKeyring()}
			},
		},
	}

	tests := []struct {
		name string
		run  func(t *testing.T, store Store)
	}{
		{
			name: "a blank namespace is rejected",
			run: func(t *testing.T, store Store) {
				_, err := store.Load("")
				require.Error(t, err)
				require.Error(t, store.Save("", sampleTokens()))
				require.Error(t, store.Delete(""))
			},
		},
		{
			name: "saving nil tokens is rejected",
			run: func(t *testing.T, store Store) {
				require.Error(t, store.Save(testNamespace, nil))
			},
		},
		{
			name: "loading an unknown namespace is not an error",
			run: func(t *testing.T, store Store) {
				got, err := store.Load(testNamespace)
				require.NoError(t, err, "not logged in is a normal state, not a failure")
				assert.Nil(t, got)
			},
		},
		{
			name: "deleting an unknown namespace is not an error",
			run: func(t *testing.T, store Store) {
				require.NoError(t, store.Delete(testNamespace))
			},
		},
		{
			name: "delete clears a saved token set",
			run: func(t *testing.T, store Store) {
				require.NoError(t, store.Save(testNamespace, sampleTokens()))
				require.NoError(t, store.Delete(testNamespace))

				got, err := store.Load(testNamespace)
				require.NoError(t, err)
				assert.Nil(t, got)
			},
		},
		{
			name: "namespaces do not collide",
			run: func(t *testing.T, store Store) {
				staging := sampleTokens()
				staging.AccessToken = "staging-access"
				production := sampleTokens()
				production.AccessToken = "production-access"

				require.NoError(t, store.Save("pay-admin/staging", staging))
				require.NoError(t, store.Save("pay-admin/production", production))

				got, err := store.Load("pay-admin/staging")
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "staging-access", got.AccessToken)

				require.NoError(t, store.Delete("pay-admin/staging"))

				survivor, err := store.Load("pay-admin/production")
				require.NoError(t, err)
				require.NotNil(t, survivor, "deleting one namespace must not touch another")
				assert.Equal(t, "production-access", survivor.AccessToken)
			},
		},
		{
			name: "save then load round-trips every field",
			run: func(t *testing.T, store Store) {
				want := sampleTokens()
				require.NoError(t, store.Save(testNamespace, want))

				got, err := store.Load(testNamespace)
				require.NoError(t, err)
				require.NotNil(t, got)

				assert.Equal(t, want.AccessToken, got.AccessToken)
				assert.Equal(t, want.IDToken, got.IDToken)
				assert.Equal(t, want.RefreshToken, got.RefreshToken)
				assert.Equal(t, want.ExpiresAt.UTC(), got.ExpiresAt.UTC())
			},
		},
		{
			name: "save overwrites an earlier token set",
			run: func(t *testing.T, store Store) {
				require.NoError(t, store.Save(testNamespace, sampleTokens()))

				replacement := sampleTokens()
				replacement.AccessToken = "second-access"
				require.NoError(t, store.Save(testNamespace, replacement))

				got, err := store.Load(testNamespace)
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "second-access", got.AccessToken)
			},
		},
		{
			name: "the loaded token set is a copy",
			run: func(t *testing.T, store Store) {
				require.NoError(t, store.Save(testNamespace, sampleTokens()))

				first, err := store.Load(testNamespace)
				require.NoError(t, err)
				require.NotNil(t, first)
				first.AccessToken = "mutated-by-caller"

				second, err := store.Load(testNamespace)
				require.NoError(t, err)
				require.NotNil(t, second)
				assert.Equal(t, "access-value", second.AccessToken, "a caller must not be able to mutate stored tokens")
			},
		},
	}

	for _, impl := range impls {
		for _, tt := range tests {
			t.Run(impl.name+"/"+tt.name, func(t *testing.T) {
				tt.run(t, impl.build())
			})
		}
	}
}

func TestKeyringStore_LoadFailures(t *testing.T) {
	tests := []struct {
		name            string
		givenSetup      func(*fakeKeyring)
		wantNil         bool
		wantErrContains string
	}{
		{
			name: "a keychain failure on the gate entry is reported",
			givenSetup: func(f *fakeKeyring) {
				f.getErr = errors.New("keychain is locked")
			},
			wantErrContains: "keychain",
		},
		{
			name: "a missing id_token is an inconsistency, not a logged-out state",
			givenSetup: func(f *fakeKeyring) {
				f.put(fieldAccess, "access-value")
			},
			wantErrContains: "id",
		},
		{
			name: "a missing refresh_token is an inconsistency",
			givenSetup: func(f *fakeKeyring) {
				f.put(fieldAccess, "access-value")
				f.put(fieldID, "id-value")
			},
			wantErrContains: "refresh",
		},
		{
			name: "missing metadata is an inconsistency",
			givenSetup: func(f *fakeKeyring) {
				f.put(fieldAccess, "access-value")
				f.put(fieldID, "id-value")
				f.put(fieldRefresh, "refresh-value")
			},
			wantErrContains: "meta",
		},
		{
			name: "corrupt metadata is reported",
			givenSetup: func(f *fakeKeyring) {
				f.put(fieldAccess, "access-value")
				f.put(fieldID, "id-value")
				f.put(fieldRefresh, "refresh-value")
				f.put(fieldMeta, "{not json")
			},
			wantErrContains: "metadata",
		},
		{
			name:       "nothing stored reads as not logged in",
			givenSetup: func(_ *fakeKeyring) {},
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := newFakeKeyring()
			tt.givenSetup(backend)
			store := &keyringStore{service: testService, backend: backend}

			got, err := store.Load(testNamespace)

			if wegostrings.IsNotEmpty(tt.wantErrContains) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Nil(t, got)
		})
	}
}

// TestKeyringStore_SaveSplitsFieldsAcrossEntries pins the workaround for the
// 4 KiB argument cap that zalando/go-keyring hits on macOS: one entry per
// field, with the access token written last so a torn write reads as
// "not logged in" rather than as a half-populated token set.
func TestKeyringStore_SaveSplitsFieldsAcrossEntries(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, sampleTokens()))

	assert.Equal(t, []string{
		testNamespace + "/refresh",
		testNamespace + "/id",
		testNamespace + "/meta",
		testNamespace + "/access",
	}, backend.writes, "one namespaced entry per field, access token last")

	assert.Equal(t, "access-value", backend.value(fieldAccess))
	assert.Equal(t, "refresh-value", backend.value(fieldRefresh))
}

func TestKeyringStore_BackendFailuresAreReported(t *testing.T) {
	tests := []struct {
		name            string
		givenSetup      func(*fakeKeyring)
		givenAction     func(Store) error
		wantErrContains string
	}{
		{
			name: "a save failure is reported",
			givenSetup: func(f *fakeKeyring) {
				f.setErr = errors.New("keychain is locked")
			},
			givenAction:     func(s Store) error { return s.Save(testNamespace, sampleTokens()) },
			wantErrContains: "keychain",
		},
		{
			name: "a delete failure is reported",
			givenSetup: func(f *fakeKeyring) {
				f.deleteErr = errors.New("keychain is locked")
			},
			givenAction:     func(s Store) error { return s.Delete(testNamespace) },
			wantErrContains: "keychain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := newFakeKeyring()
			tt.givenSetup(backend)
			store := &keyringStore{service: testService, backend: backend}

			err := tt.givenAction(store)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrContains)
		})
	}
}

// TestKeyringStore_DeleteToleratesMissingEntries covers a half-populated
// namespace: sign-out must succeed rather than stranding the operator.
func TestKeyringStore_DeleteToleratesMissingEntries(t *testing.T) {
	backend := newFakeKeyring()
	backend.put(fieldAccess, "access-value")
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Delete(testNamespace))
	assert.Empty(t, backend.entries)
}

func TestKeyringStore_BlankServiceIsRejected(t *testing.T) {
	store := &keyringStore{service: "", backend: newFakeKeyring()}

	_, err := store.Load(testNamespace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service")
	require.Error(t, store.Save(testNamespace, sampleTokens()))
	require.Error(t, store.Delete(testNamespace))
}

func sampleTokens() *cognito.TokenSet {
	return &cognito.TokenSet{
		AccessToken:  "access-value",
		IDToken:      "id-value",
		RefreshToken: "refresh-value",
		ExpiresAt:    testExpiry,
	}
}

// fakeKeyring is an in-memory stand-in for the OS keychain.
type fakeKeyring struct {
	mu        sync.Mutex
	entries   map[string]string
	writes    []string
	setErr    error
	getErr    error
	deleteErr error
}

func newFakeKeyring() *fakeKeyring {
	return &fakeKeyring{entries: make(map[string]string)}
}

func (f *fakeKeyring) Set(service, user, password string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.setErr != nil {
		return f.setErr
	}
	f.entries[service+"|"+user] = password
	f.writes = append(f.writes, user)
	return nil
}

func (f *fakeKeyring) Get(service, user string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return "", f.getErr
	}
	value, ok := f.entries[service+"|"+user]
	if !ok {
		return "", keyring.ErrNotFound
	}
	return value, nil
}

func (f *fakeKeyring) Delete(service, user string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.deleteErr != nil {
		return f.deleteErr
	}
	key := service + "|" + user
	if _, ok := f.entries[key]; !ok {
		return keyring.ErrNotFound
	}
	delete(f.entries, key)
	return nil
}

// put seeds an entry under testNamespace, bypassing the store under test.
func (f *fakeKeyring) put(field, value string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries[testService+"|"+testNamespace+"/"+field] = value
}

// value reads an entry under testNamespace, bypassing the store under test.
func (f *fakeKeyring) value(field string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.entries[testService+"|"+testNamespace+"/"+field]
}
