package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
			name: "an uncommitted generation reads as not logged in",
			givenSetup: func(f *fakeKeyring) {
				f.putField(testGeneration, fieldAccess, "access-value")
				f.putField(testGeneration, fieldID, "id-value")
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

	generation := backend.liveGeneration(t)
	for _, field := range tokenFields {
		assert.Contains(t, backend.writes, testNamespace+"/"+generation+"/"+field,
			"each field keeps its own entry, to stay under the 4 KiB cap")
	}
}

// TestKeyringStore_SaveNeverReusesAGeneration covers the property both
// round-4 findings turn on. A name that can come round again lets two writers
// collide and lets a pointer cycle back to a value a reader started on.
func TestKeyringStore_SaveNeverReusesAGeneration(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	seen := map[string]bool{}

	for i := range 6 {
		tokens := sampleTokens()
		tokens.AccessToken = fmt.Sprintf("access-%d", i)
		require.NoError(t, store.Save(testNamespace, tokens))

		generation := backend.liveGeneration(t)
		assert.False(t, seen[generation], "generation %q was reused", generation)
		seen[generation] = true

		got, err := store.Load(testNamespace)
		require.NoError(t, err)
		assert.Equal(t, tokens.AccessToken, got.AccessToken, "the newest session must be the live one")
	}

	assert.Len(t, seen, 6)
}

// TestKeyringStore_SaveReapsTheGenerationBeforeLast covers cleanup. The
// generation directly replaced is kept, because a Load may still be reading
// it; the one before that cannot have a reader who started after its
// replacement committed, so it goes.
func TestKeyringStore_SaveReapsTheGenerationBeforeLast(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, session("first")))
	first := backend.liveGeneration(t)

	require.NoError(t, store.Save(testNamespace, session("second")))
	second := backend.liveGeneration(t)

	require.NoError(t, store.Save(testNamespace, session("third")))

	assert.Empty(t, backend.entries[testService+"|"+testNamespace+"/"+first+"/"+fieldAccess],
		"the generation before last must be cleared")
	assert.NotEmpty(t, backend.entries[testService+"|"+testNamespace+"/"+second+"/"+fieldAccess],
		"the generation just replaced is kept for an in-flight Load")
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

// TestKeyringStore_LoadRejectsAnUnreadablePointer covers a pointer entry this
// code did not write. Guessing what it meant would be worse than refusing.
func TestKeyringStore_LoadRejectsAnUnreadablePointer(t *testing.T) {
	tests := []struct {
		name       string
		givenValue string
		wantErr    string
	}{
		{name: "not json", givenValue: "somewhere-else", wantErr: "parse the token pointer"},
		{name: "json naming no generation", givenValue: `{"current":""}`, wantErr: "names no generation"},
		// A bare slot name is what a pre-release build of this package wrote,
		// before generations replaced two reusable slots. That shape never
		// shipped, so no released version can produce it and there is nothing
		// to migrate — but a keychain carrying one still has to fail readably
		// rather than with a raw json error.
		{name: "a bare slot name from a pre-release layout", givenValue: "a", wantErr: "parse the token pointer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := newFakeKeyring()
			backend.put(fieldCurrent, tt.givenValue)
			store := &keyringStore{service: testService, backend: backend}

			_, err := store.Load(testNamespace)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Contains(t, err.Error(), "sign out of this environment and sign in again",
				"an unreadable pointer must name the recovery, which is not guessable from a json error")
		})
	}
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
	backend.putField("gen-current", fieldAccess, "access-value")
	backend.putField("gen-previous", fieldRefresh, "superseded-refresh")
	backend.putPointer("gen-current", "gen-previous")
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Delete(testNamespace))
	assert.Empty(t, backend.entries,
		"no token material may survive a sign-out, in either generation the pointer knows")
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

// TestKeyringStore_SaveStillSignsInWhenThePointerCannotBeRead covers a
// keychain that cannot be read at all.
//
// Under the old two-slot layout Save HAD to read the pointer, because it wrote
// to whichever slot the live one was not using; an unreadable pointer
// therefore had to fail rather than risk overwriting the live session. A fresh
// generation collides with nothing, so the read is now only an optimisation
// for reaping, and a sign-in can succeed without it.
func TestKeyringStore_SaveStillSignsInWhenThePointerCannotBeRead(t *testing.T) {
	backend := newFakeKeyring()
	backend.getErr = errors.New("keychain is locked")
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, sampleTokens()),
		"an unreadable pointer must not block establishing a session")

	backend.getErr = nil

	got, err := store.Load(testNamespace)
	require.NoError(t, err)
	assert.Equal(t, sampleTokens(), got)
}

// TestKeyringStore_SaveOverAnUnreadablePointerStillSignsIn covers a pointer
// holding a value this code never wrote. Load refuses it, but Save's job is to
// establish a good session and it can do that without deciding what the bad
// value meant.
func TestKeyringStore_SaveOverAnUnreadablePointerStillSignsIn(t *testing.T) {
	backend := newFakeKeyring()
	backend.put(fieldCurrent, "somewhere-else")
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, sampleTokens()))

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
	assert.Contains(t, err.Error(), "token pointer")
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
	// onSet fires on every Set, before the write happens, so a test can drive
	// a second writer into the middle of one Save.
	onSet func(account string)
}

// clearOnSet disarms the write hook. A hook that drives a nested Save MUST
// call this before doing so: the nested writes re-enter the hook, and guarding
// with sync.Once instead deadlocks, because Once.Do cannot be re-entered.
func (f *fakeKeyring) clearOnSet() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.onSet = nil
}

// clearOnGet disarms the read hook, for the same reason as clearOnSet.
func (f *fakeKeyring) clearOnGet() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.onGet = nil
}

// takeOnSet returns the write hook, if any.
func (f *fakeKeyring) takeOnSet() func(string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.onSet
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
	// Outside the lock, like onGet, so a hook can run a whole Save.
	if hook := f.takeOnSet(); hook != nil {
		hook(user)
	}

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

// testGeneration is a fixed generation name for seeding, standing in for the
// random one Save would mint.
const testGeneration = "gen0"

// putCommitted seeds a committed session under testGeneration, field by field,
// so a test can then remove or corrupt exactly one part of it.
func (f *fakeKeyring) putCommitted(fields map[string]string) {
	for field, value := range fields {
		f.putField(testGeneration, field, value)
	}
	f.putPointer(testGeneration, "")
}

// putPointer seeds the commit record directly.
func (f *fakeKeyring) putPointer(current, previous string) {
	raw, err := json.Marshal(keyringPointer{Current: current, Previous: previous})
	if err != nil {
		panic(err)
	}
	f.put(fieldCurrent, string(raw))
}

// liveGeneration reads the committed generation name straight out of the fake.
func (f *fakeKeyring) liveGeneration(t *testing.T) string {
	t.Helper()

	raw := f.value(fieldCurrent)
	require.NotEmpty(t, raw, "no pointer entry was written")

	var pointer keyringPointer
	require.NoError(t, json.Unmarshal([]byte(raw), &pointer))

	return pointer.Current
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
	backend.onGet = func(account string) {
		if account == testNamespace+"/"+fieldCurrent {
			return // the pointer read itself; let it through
		}

		backend.clearOnGet()
		require.NoError(t, store.Save(testNamespace, replacement))
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

// sessionIsWhole asserts a loaded token set came entirely from one session.
// A set pairing one login's access token with another's id and refresh tokens
// is the specific corruption these tests exist to prevent: the id token
// supplies the operator identity that admin writes are audited against, so a
// mismatched pair can attribute a production change to the wrong person.
func sessionIsWhole(t *testing.T, got *cognito.TokenSet, sessions map[string]*cognito.TokenSet) {
	t.Helper()
	require.NotNil(t, got)

	for name, want := range sessions {
		if got.AccessToken != want.AccessToken {
			continue
		}

		assert.Equal(t, want.IDToken, got.IDToken, "id token must come from the same session as the access token (%s)", name)
		assert.Equal(t, want.RefreshToken, got.RefreshToken, "refresh token must come from the same session as the access token (%s)", name)
		assert.Equal(t, want.ExpiresAt, got.ExpiresAt, "expiry must come from the same session as the access token (%s)", name)

		return
	}

	t.Fatalf("loaded access token %q belongs to no known session", got.AccessToken)
}

func session(tag string) *cognito.TokenSet {
	return &cognito.TokenSet{
		AccessToken:  tag + "-access",
		IDToken:      tag + "-id",
		RefreshToken: tag + "-refresh",
		ExpiresAt:    testExpiry,
	}
}

// TestKeyringStore_ConcurrentSavesNeverCommitAHybrid is the first regression
// named in review round 4.
//
// Two pay-admin processes overlapping a login and a refresh both used to read
// the same live slot, both compute the same "other" slot, and interleave their
// field writes into it. Both then committed, leaving one slot holding one
// session's access token beside another's id and refresh tokens.
func TestKeyringStore_ConcurrentSavesNeverCommitAHybrid(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	seed, first, second := session("seed"), session("first"), session("second")
	require.NoError(t, store.Save(testNamespace, seed))

	// The second writer runs to completion inside the first writer's very first
	// field write, which is the schedule that produced the hybrid.
	// Fire once the first writer has already stored its access token but
	// before its id token, so the second writer's whole session lands in
	// between. Firing earlier lets the first writer simply overwrite
	// everything, which is consistent and proves nothing.
	backend.onSet = func(account string) {
		if !strings.HasSuffix(account, "/"+fieldID) {
			return
		}

		backend.clearOnSet()
		require.NoError(t, store.Save(testNamespace, second))
	}

	require.NoError(t, store.Save(testNamespace, first))

	backend.onSet = nil

	got, err := store.Load(testNamespace)
	require.NoError(t, err)
	sessionIsWhole(t, got, map[string]*cognito.TokenSet{"seed": seed, "first": first, "second": second})
}

// TestKeyringStore_LoadIsNotFooledByPointerReuse is the second regression named
// in review round 4, the A->B->A schedule.
//
// Load re-reads the pointer after a failed field read and retried only when it
// had changed. With two reusable slots the pointer could cycle back to the one
// Load started on, so a changed generation looked unchanged and Load returned
// fields belonging to a session it never selected.
func TestKeyringStore_LoadIsNotFooledByPointerReuse(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	seed, first, second := session("seed"), session("first"), session("second")
	require.NoError(t, store.Save(testNamespace, seed))

	// Two commits land after Load has chosen its generation and read part of
	// it. Under slot reuse the second commit lands back on the first slot.
	// Fire after Load has read the access token but before the rest, so the
	// commits land inside one read of one generation.
	backend.onGet = func(account string) {
		if !strings.HasSuffix(account, "/"+fieldID) {
			return
		}

		backend.clearOnGet()
		require.NoError(t, store.Save(testNamespace, first))
		require.NoError(t, store.Save(testNamespace, second))
	}

	got, err := store.Load(testNamespace)
	require.NoError(t, err)
	sessionIsWhole(t, got, map[string]*cognito.TokenSet{"seed": seed, "first": first, "second": second})
}

// TestKeyringStore_LoadReportsAPointerFailureDuringRetry covers the branch
// where the pointer read that decides retry-or-report itself fails. Reporting
// that beats retrying blindly or claiming the namespace is inconsistent.
func TestKeyringStore_LoadReportsAPointerFailureDuringRetry(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}
	require.NoError(t, store.Save(testNamespace, sampleTokens()))

	// Break the field read, then break the keychain outright so the retry
	// check cannot establish whether the generation moved.
	generation := backend.liveGeneration(t)
	delete(backend.entries, testService+"|"+testNamespace+"/"+generation+"/"+fieldID)

	backend.onGet = func(account string) {
		if !strings.HasSuffix(account, "/"+fieldID) {
			return
		}

		backend.clearOnGet()
		backend.mu.Lock()
		backend.getErr = errors.New("keychain is locked")
		backend.mu.Unlock()
	}

	_, err := store.Load(testNamespace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "keychain")
}

// TestKeyringStore_SaveReportsACommitEncodingFailure is not reachable through
// the public API -- keyringPointer always marshals -- so the pointer encoder is
// exercised directly to prove the record round-trips exactly.
func TestKeyringStore_PointerRoundTrips(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}

	require.NoError(t, store.Save(testNamespace, session("one")))
	first := backend.liveGeneration(t)

	require.NoError(t, store.Save(testNamespace, session("two")))

	pointer, err := store.readPointer(testNamespace)
	require.NoError(t, err)
	assert.Equal(t, backend.liveGeneration(t), pointer.Current)
	assert.Equal(t, first, pointer.Previous, "the record must remember what it replaced, so Delete can reap it")
}

// TestKeyringStore_SaveReportsAFieldWriteFailure covers the write loop's error
// path with the generation layout, and that a failure leaves the live session
// untouched.
func TestKeyringStore_SaveReportsAFieldWriteFailure(t *testing.T) {
	backend := newFakeKeyring()
	store := &keyringStore{service: testService, backend: backend}
	require.NoError(t, store.Save(testNamespace, session("live")))

	backend.okWrites = 0
	backend.failSetAfter = 0

	err := store.Save(testNamespace, session("doomed"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "keychain")

	backend.failSetAfter = -1

	got, err := store.Load(testNamespace)
	require.NoError(t, err)
	assert.Equal(t, session("live"), got, "a failed write must not disturb the live session")
}
