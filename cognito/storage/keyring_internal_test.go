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
			// The pointer says a session was committed, so anything the slot is
			// missing is an inconsistency rather than a logged-out state.
			name: "a missing id_token is an inconsistency, not a logged-out state",
			givenSetup: func(f *fakeKeyring) {
				f.putCommitted(map[string]string{fieldAccess: "access-value"})
			},
			wantErrContains: "id",
		},
		{
			name: "a missing refresh_token is an inconsistency",
			givenSetup: func(f *fakeKeyring) {
				f.putCommitted(map[string]string{
					fieldAccess: "access-value",
					fieldID:     "id-value",
				})
			},
			wantErrContains: "refresh",
		},
		{
			name: "missing metadata is an inconsistency",
			givenSetup: func(f *fakeKeyring) {
				f.putCommitted(map[string]string{
					fieldAccess:  "access-value",
					fieldID:      "id-value",
					fieldRefresh: "refresh-value",
				})
			},
			wantErrContains: "meta",
		},
		{
			name: "corrupt metadata is reported",
			givenSetup: func(f *fakeKeyring) {
				f.putCommitted(map[string]string{
					fieldAccess:  "access-value",
					fieldID:      "id-value",
					fieldRefresh: "refresh-value",
					fieldMeta:    "{not json",
				})
			},
			wantErrContains: "metadata",
		},
		{
			// Fields present but never committed: the sign-in was torn before
			// the pointer moved, so the operator is simply not signed in.
			name: "an uncommitted slot reads as not logged in",
			givenSetup: func(f *fakeKeyring) {
				f.putField(slotA, fieldAccess, "access-value")
				f.putField(slotA, fieldID, "id-value")
			},
			wantNil: true,
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

// TestKeyringStore_SaveCommitsWithOnePointerWrite pins both halves of the
// storage contract.
//
// Fields stay in SEPARATE entries because zalando/go-keyring shells out to
// /usr/bin/security on macOS and refuses any command over 4096 bytes
// (keyring_darwin.go). After its base64 expansion that leaves roughly 3 KB of
// secret per entry, which a single combined token set would exceed.
//
// Atomicity therefore cannot come from writing one entry. It comes from
// writing the fields into the inactive slot and then moving the pointer, which
// is ONE small Set and the only write that changes what Load sees.
func TestKeyringStore_SaveCommitsWithOnePointerWrite(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, sampleTokens()))

	require.NotEmpty(t, backend.writes)
	assert.Equal(t, testNamespace+"/"+fieldCurrent, backend.writes[len(backend.writes)-1],
		"the pointer must be the last write, so nothing before it is observable")
	assert.Equal(t, 1, countWrites(backend.writes, testNamespace+"/"+fieldCurrent),
		"the commit must be a single pointer write")

	for _, field := range []string{fieldAccess, fieldID, fieldRefresh, fieldMeta} {
		assert.Contains(t, backend.writes, testNamespace+"/"+slotA+"/"+field,
			"each field keeps its own entry, to stay under the 4 KiB cap")
	}
}

// TestKeyringStore_SaveAlternatesSlots covers why there are two slots: a
// re-login must not overwrite the entries the current session is still being
// read from, or a torn write would corrupt the live session rather than an
// unused copy.
func TestKeyringStore_SaveAlternatesSlots(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, sampleTokens()))
	assert.Equal(t, slotA, backend.value(fieldCurrent))

	second := sampleTokens()
	second.AccessToken = "second-access"
	require.NoError(t, store.Save(testNamespace, second))
	assert.Equal(t, slotB, backend.value(fieldCurrent))

	third := sampleTokens()
	third.AccessToken = "third-access"
	require.NoError(t, store.Save(testNamespace, third))
	assert.Equal(t, slotA, backend.value(fieldCurrent), "slots alternate rather than growing")

	got, err := store.Load(testNamespace)
	require.NoError(t, err)
	assert.Equal(t, "third-access", got.AccessToken)
}

// TestKeyringStore_TornResaveLeavesThePreviousSession is the regression test
// for the reported defect.
//
// With one entry per field and no commit pointer, a re-login that failed
// partway left the NEW refresh token beside the OLD access token, id token and
// expiry. Load gated on the access token, which was present, so it handed that
// mixture back as a live session: a stale expiry paired with a fresh refresh
// token, or after an account switch one identity's id token paired with
// another's refresh token.
func TestKeyringStore_TornResaveLeavesThePreviousSession(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, sampleTokens()))

	// Let two writes of the re-login land, then fail.
	backend.okWrites = 0
	backend.failSetAfter = 2

	newer := &cognito.TokenSet{
		AccessToken:  "new-access",
		IDToken:      "new-id",
		RefreshToken: "new-refresh",
		ExpiresAt:    testExpiry.Add(time.Hour),
	}
	require.Error(t, store.Save(testNamespace, newer))

	backend.failSetAfter = -1

	got, err := store.Load(testNamespace)
	require.NoError(t, err, "the previous session must still be readable")
	require.NotNil(t, got)
	assert.Equal(t, sampleTokens(), got,
		"a torn re-login must leave the previous session exactly, never a mixture of the two")
}

// TestKeyringStore_LoadRejectsAnUnknownSlot covers a pointer naming a slot that
// is not one of the two. That cannot arise from this code, so it means the
// entry was tampered with or written by another version, and guessing which
// slot was meant would be worse than refusing.
func TestKeyringStore_LoadRejectsAnUnknownSlot(t *testing.T) {
	backend := newFakeKeyring()
	backend.put(fieldCurrent, "somewhere-else")
	store := &keyringStore{service: testService, backend: backend}

	_, err := store.Load(testNamespace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slot")
}

// countWrites reports how many times account appears in a write log.
func countWrites(writes []string, account string) int {
	n := 0
	for _, w := range writes {
		if w == account {
			n++
		}
	}

	return n
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
// namespace: sign-out must succeed rather than stranding the operator, and it
// must clear the slot a torn Save left behind as well as the live one.
func TestKeyringStore_DeleteToleratesMissingEntries(t *testing.T) {
	backend := newFakeKeyring()
	backend.putField(slotA, fieldAccess, "access-value")
	backend.putField(slotB, fieldRefresh, "orphaned-refresh")
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Delete(testNamespace))
	assert.Empty(t, backend.entries, "no token material may survive a sign-out, in either slot")
}

// TestKeyringStore_FailedCommitKeepsThePreviousSession covers the last write.
// Every field of the new session is already in the keychain; only the pointer
// move failed. Load must still return the old session, because the new one was
// never committed.
func TestKeyringStore_FailedCommitKeepsThePreviousSession(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, sampleTokens()))

	// Let all four field writes land, then fail the pointer write.
	backend.okWrites = 0
	backend.failSetAfter = len(tokenFields)

	newer := sampleTokens()
	newer.AccessToken = "new-access"

	err := store.Save(testNamespace, newer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "commit", "the failure must name the commit, not a field")

	backend.failSetAfter = -1

	got, err := store.Load(testNamespace)
	require.NoError(t, err)
	assert.Equal(t, sampleTokens(), got, "an uncommitted session must not become the live one")
}

// TestKeyringStore_SaveReportsAnUnreadablePointer covers a keychain that
// cannot be read at all. Save must not proceed on a guess about which slot is
// live: writing to the wrong one would overwrite the session it is meant to
// protect.
func TestKeyringStore_SaveReportsAnUnreadablePointer(t *testing.T) {
	backend := newFakeKeyring()
	backend.getErr = errors.New("keychain is locked")
	store := &keyringStore{service: testService, backend: backend}

	err := store.Save(testNamespace, sampleTokens())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slot")
	assert.Empty(t, backend.writes, "nothing may be written when the live slot is unknown")
}

// TestKeyringStore_SaveOverAnUnknownSlotStillSignsIn covers a pointer holding
// a value this code never writes. Load refuses it, but Save's job is to
// establish a good session and it can do that without deciding what the bad
// value meant.
func TestKeyringStore_SaveOverAnUnknownSlotStillSignsIn(t *testing.T) {
	backend := newFakeKeyring()
	backend.put(fieldCurrent, "somewhere-else")
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, sampleTokens()))
	assert.Equal(t, slotA, backend.value(fieldCurrent))

	got, err := store.Load(testNamespace)
	require.NoError(t, err)
	assert.Equal(t, sampleTokens(), got)
}

// TestKeyringStore_DeleteReportsAPointerFailure covers sign-out when the
// pointer cannot be removed. That is the entry that decides whether the
// operator is signed in, so the failure has to be reported rather than
// swallowed after clearing the fields.
func TestKeyringStore_DeleteReportsAPointerFailure(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}
	require.NoError(t, store.Save(testNamespace, sampleTokens()))

	backend.deleteErr = errors.New("keychain is locked")

	err := store.Delete(testNamespace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slot")
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
	mu      sync.Mutex
	entries map[string]string
	writes  []string
	setErr  error
	// failSetAfter fails every Set from the (failSetAfter+1)th of this counter
	// onward, so a test can tear a write sequence in the middle. Negative
	// disables it. Reset okWrites to re-arm.
	failSetAfter int
	okWrites     int
	getErr       error
	deleteErr    error
	// onGet fires once, on the next Get, before the read happens.
	onGet func(account string)
}

// takeOnGet returns the hook, if any. It does NOT clear it: the hook decides
// which account it wants to fire on, so consuming it on the first unrelated
// read would silently disarm the test.
func (f *fakeKeyring) takeOnGet() func(string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.onGet
}

func newFakeKeyring() *fakeKeyring {
	return &fakeKeyring{entries: make(map[string]string), failSetAfter: -1}
}

func (f *fakeKeyring) Set(service, user, password string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.setErr != nil {
		return f.setErr
	}
	if f.failSetAfter >= 0 && f.okWrites >= f.failSetAfter {
		return errors.New("keychain is locked")
	}
	f.okWrites++
	f.entries[service+"|"+user] = password
	f.writes = append(f.writes, user)
	return nil
}

func (f *fakeKeyring) Get(service, user string) (string, error) {
	// onGet runs OUTSIDE the lock and before the read, so a hook can drive a
	// whole Save through this same backend without deadlocking.
	if hook := f.takeOnGet(); hook != nil {
		hook(user)
	}

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

// put seeds a namespace-level entry (the pointer) under testNamespace,
// bypassing the store under test.
func (f *fakeKeyring) put(field, value string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries[testService+"|"+testNamespace+"/"+field] = value
}

// putField seeds one field of one slot, bypassing the store under test.
func (f *fakeKeyring) putField(slot, field, value string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries[testService+"|"+testNamespace+"/"+slot+"/"+field] = value
}

// putCommitted seeds a committed session in slotA, field by field, so a test
// can then remove or corrupt exactly one part of it.
func (f *fakeKeyring) putCommitted(fields map[string]string) {
	for field, value := range fields {
		f.putField(slotA, field, value)
	}
	f.put(fieldCurrent, slotA)
}

// value reads an entry under testNamespace, bypassing the store under test.
func (f *fakeKeyring) value(field string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.entries[testService+"|"+testNamespace+"/"+field]
}

// TestKeyringStore_LoadSurvivesAConcurrentCommit is the regression for a race
// the two-slot commit introduced.
//
// Load reads the pointer, then reads that slot's four fields one at a time.
// Between those reads another pay-admin process can finish a login or a
// refresh, commit the replacement slot and clear the one this Load already
// chose, so the read failed with "secret not found in keyring" even though a
// perfectly good session existed the whole time.
//
// Two overlapping processes is ordinary: a shell running a command while a
// refresh fires, or two terminals against the same namespace.
func TestKeyringStore_LoadSurvivesAConcurrentCommit(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, sampleTokens()))

	replacement := &cognito.TokenSet{
		AccessToken:  "refreshed-access",
		IDToken:      "refreshed-id",
		RefreshToken: "refreshed-refresh",
		ExpiresAt:    testExpiry.Add(time.Hour),
	}

	// Commit the replacement the moment this Load starts reading fields, which
	// is exactly the window the race lives in.
	var once sync.Once
	backend.onGet = func(account string) {
		if account == testNamespace+"/"+fieldCurrent {
			return // the pointer read itself; let it through
		}
		once.Do(func() {
			require.NoError(t, store.Save(testNamespace, replacement))
		})
	}

	got, err := store.Load(testNamespace)
	require.NoError(t, err, "a concurrent commit must not fail a Load that was already in flight")
	require.NotNil(t, got)

	// Either session is a correct answer; a mixture is not.
	assert.Contains(t, []string{sampleTokens().AccessToken, replacement.AccessToken}, got.AccessToken)
	if got.AccessToken == replacement.AccessToken {
		assert.Equal(t, replacement.RefreshToken, got.RefreshToken, "the two sessions must not mix")
		assert.Equal(t, replacement.IDToken, got.IDToken)
	} else {
		assert.Equal(t, sampleTokens().RefreshToken, got.RefreshToken, "the two sessions must not mix")
	}
}

// TestKeyringStore_LoadGivesUpOnEndlessCommits covers the retry bound. A
// namespace being rewritten faster than it can be read is not something a
// retry can fix, so Load must report it rather than spin.
func TestKeyringStore_LoadGivesUpOnEndlessCommits(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}
	require.NoError(t, store.Save(testNamespace, sampleTokens()))

	// Commit a fresh session on every field read, so the pointer has always
	// moved by the time the retry check looks.
	backend.onGet = func(account string) {
		if account == testNamespace+"/"+fieldCurrent {
			return
		}

		next := sampleTokens()
		next.AccessToken = "rewritten"
		_ = store.Save(testNamespace, next)
	}

	_, err := store.Load(testNamespace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "replaced")
}

// TestKeyringStore_LoadReportsARealInconsistency pins the other side of the
// retry: when the pointer has NOT moved, a missing field is a genuinely broken
// namespace and must be reported, not retried into a generic timeout.
func TestKeyringStore_LoadReportsARealInconsistency(t *testing.T) {
	backend := newFakeKeyring()
	backend.putCommitted(map[string]string{fieldAccess: "access-value"})
	store := &keyringStore{service: testService, backend: backend}

	_, err := store.Load(testNamespace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id", "the missing field must be named")
	assert.NotContains(t, err.Error(), "replaced")
}
