package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	wegostrings "github.com/wego/pkg/strings"
	"github.com/zalando/go-keyring"

	"github.com/wego/pkg/cognito"
)

// Layout.
//
// A Cognito JWT - the access token especially - can exceed the 4 KiB
// command-line cap zalando/go-keyring runs into on macOS, where it shells out
// to /usr/bin/security (see keyring_darwin.go, which refuses any command over
// 4096 bytes). After the library's base64 expansion that leaves roughly 3 KB
// of secret per entry, and a whole token set does not reliably fit. So each
// field keeps its own entry.
//
// That rules out getting atomicity by writing one entry, and a keychain has no
// transaction. Instead each token set is written under a fresh GENERATION
// name, and a separate pointer entry names the generation that counts. Save
// writes its fields and then moves the pointer, which is one small write and
// the only write that changes what Load sees. A failure anywhere before it
// leaves the pointer and the live generation untouched, so a torn write costs
// the new session, never the old one.
//
// Generations are UNIQUE AND NEVER REUSED, and that is load-bearing twice
// over. An earlier version alternated between two fixed slots, which broke in
// two ways review found:
//
//   - Two concurrent Save calls both read the same live slot, both computed
//     the same "other" slot, and interleaved their field writes into it. Both
//     committed, leaving one slot holding one session's access token beside
//     another's id and refresh tokens. Fresh names make the writers disjoint,
//     so each generation is whole and the later commit simply wins.
//   - The pointer could cycle A->B->A, so a generation that HAD changed under
//     an in-flight Load looked unchanged and Load returned fields it never
//     selected. A name that never repeats makes any change detectable.
//
// A hybrid is not merely untidy: the id token carries the operator identity
// that admin writes are audited against, so pairing it with another session's
// access or refresh token can attribute a production change to the wrong
// person.
//
// Entries are accounted as "<namespace>/current" for the pointer and
// "<namespace>/<generation>/<field>" for the token fields.
const (
	fieldAccess  = "access"
	fieldID      = "id"
	fieldRefresh = "refresh"
	fieldMeta    = "meta"
	// fieldCurrent is the pointer entry naming the live generation. Committing
	// a token set is exactly one write of this entry.
	fieldCurrent = "current"
)

// generationBytes is the entropy in a generation name. A generation is never
// reused, so this only has to make an accidental collision between two
// concurrent writers impossible in practice.
const generationBytes = 12

// loadAttempts is how many times Load will re-read after a concurrent commit
// moves the pointer out from under it. Each retry needs another process to
// commit a whole session in the gap, so three is far past what a real overlap
// produces and still terminates.
const loadAttempts = 3

// tokenFields are the per-field entries that make up one stored token set.
var tokenFields = []string{fieldAccess, fieldID, fieldRefresh, fieldMeta}

// keyringBackend is the slice of zalando/go-keyring this package depends on.
// Naming it lets tests substitute a fake instead of prompting a real keychain.
type keyringBackend interface {
	Set(service, user, password string) error
	Get(service, user string) (string, error)
	Delete(service, user string) error
}

// systemKeyring is the real OS keychain.
type systemKeyring struct{}

func (systemKeyring) Set(service, user, password string) error {
	return keyring.Set(service, user, password)
}

func (systemKeyring) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

func (systemKeyring) Delete(service, user string) error {
	return keyring.Delete(service, user)
}

// keyringMeta is the non-secret metadata, kept in one small entry beside the
// tokens themselves.
type keyringMeta struct {
	ExpiresAt time.Time `json:"expires_at"`
}

// keyringPointer is the commit record: which generation is live, and the one
// it replaced. Previous is kept only so Delete can reap it, since a keychain
// cannot be enumerated.
type keyringPointer struct {
	Current  string `json:"current"`
	Previous string `json:"previous,omitempty"`
}

// keyringStore persists tokens in the OS keychain.
type keyringStore struct {
	service string
	backend keyringBackend
}

// NewKeyring returns a Store backed by the operating system keychain, with
// every entry filed under service.
func NewKeyring(service string) Store {
	return &keyringStore{service: service, backend: systemKeyring{}}
}

// Load returns the tokens stored under namespace, or (nil, nil) if the
// operator is not signed in.
//
// A token set is read across five keychain entries, so a concurrent commit can
// move the pointer and clear the slot midway through. That is not an error and
// must not surface as one: Load re-reads the pointer and starts again on the
// slot that is now live, which is the ordinary case of a second pay-admin
// process finishing a login or a refresh while this one reads.
func (k *keyringStore) Load(namespace string) (*cognito.TokenSet, error) {
	if err := k.validate(namespace); err != nil {
		return nil, err
	}

	for range loadAttempts {
		pointer, err := k.readPointer(namespace)
		if err != nil || wegostrings.IsBlank(pointer.Current) {
			// No pointer means not signed in, which is (nil, nil).
			return nil, err
		}

		tokens, err := k.readGeneration(namespace, pointer.Current)
		if err == nil {
			return tokens, nil
		}

		// The read failed. If the pointer has moved since it was chosen, this
		// generation was superseded and cleared under us, so try the new one.
		// If it has not, the namespace really is inconsistent and the original
		// error is the honest one to report.
		moved, checkErr := k.generationMoved(namespace, pointer.Current)
		if checkErr != nil {
			return nil, checkErr
		}

		if !moved {
			return nil, err
		}
	}

	// Every attempt lost the same race, which needs a sustained run of commits
	// against one namespace. Report it rather than looping.
	return nil, fmt.Errorf("the keychain session for %q was replaced %d times while being read", namespace, loadAttempts)
}

// readPointer returns the commit record, or a zero record when nothing is
// stored.
func (k *keyringStore) readPointer(namespace string) (keyringPointer, error) {
	raw, err := k.backend.Get(k.service, pointerAccount(namespace))
	if errors.Is(err, keyring.ErrNotFound) {
		return keyringPointer{}, nil
	}
	if err != nil {
		return keyringPointer{}, fmt.Errorf("read the keychain: %w (is the system keychain unlocked?)", err)
	}

	var pointer keyringPointer
	if err := json.Unmarshal([]byte(raw), &pointer); err != nil {
		// Nothing here writes anything but this record, so the entry was
		// tampered with or written by a version that stored something else.
		return keyringPointer{}, fmt.Errorf("parse the token pointer for %q: %w", namespace, err)
	}

	if wegostrings.IsBlank(pointer.Current) {
		return keyringPointer{}, fmt.Errorf("the token pointer for %q names no generation", namespace)
	}

	return pointer, nil
}

// generationMoved reports whether the pointer now names a generation other
// than the one a read was using.
//
// Because a generation name is never reused, this is exact: an unchanged value
// really means nothing committed in between. Under the old two-slot layout the
// pointer could cycle back to the value a reader started on, so a changed
// generation looked unchanged.
func (k *keyringStore) generationMoved(namespace, generation string) (bool, error) {
	pointer, err := k.readPointer(namespace)
	if err != nil {
		return false, err
	}

	return pointer.Current != generation, nil
}

// newGeneration mints a name no other writer will pick.
func newGeneration() (string, error) {
	buf := make([]byte, generationBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate a token generation name: %w", err)
	}

	return hex.EncodeToString(buf), nil
}

// readGeneration assembles the token set held under one generation.
func (k *keyringStore) readGeneration(namespace, generation string) (*cognito.TokenSet, error) {
	access, err := k.read(namespace, generation, fieldAccess)
	if err != nil {
		return nil, err
	}
	idToken, err := k.read(namespace, generation, fieldID)
	if err != nil {
		return nil, err
	}
	refresh, err := k.read(namespace, generation, fieldRefresh)
	if err != nil {
		return nil, err
	}
	rawMeta, err := k.read(namespace, generation, fieldMeta)
	if err != nil {
		return nil, err
	}

	var meta keyringMeta
	if err := json.Unmarshal([]byte(rawMeta), &meta); err != nil {
		return nil, fmt.Errorf("parse token metadata from the keychain: %w", err)
	}

	return &cognito.TokenSet{
		AccessToken:  access,
		IDToken:      idToken,
		RefreshToken: refresh,
		ExpiresAt:    meta.ExpiresAt,
	}, nil
}

// Save writes tokens under namespace, replacing anything already there.
//
// It is atomic from Load's point of view and safe against a concurrent Save.
// The fields go under a fresh generation no other writer will pick, and only
// the final pointer write makes them the session. If any write fails, the
// previous session is still whole and still what Load returns; if another Save
// overlaps, the two write disjoint generations and the later commit wins with
// its own session intact.
func (k *keyringStore) Save(namespace string, tokens *cognito.TokenSet) error {
	if err := k.validate(namespace); err != nil {
		return err
	}
	if tokens == nil {
		return errNoTokens
	}

	meta, err := json.Marshal(keyringMeta{ExpiresAt: tokens.ExpiresAt})
	if err != nil {
		return fmt.Errorf("encode token metadata: %w", err)
	}

	// A pointer that cannot be read is not a reason to refuse a sign-in: the
	// new generation does not collide with anything either way. What is lost is
	// only the chance to reap the generation being replaced.
	previous, _ := k.readPointer(namespace)

	generation, err := newGeneration()
	if err != nil {
		return err
	}

	values := map[string]string{
		fieldAccess:  tokens.AccessToken,
		fieldID:      tokens.IDToken,
		fieldRefresh: tokens.RefreshToken,
		fieldMeta:    string(meta),
	}

	// Ordering is irrelevant: nothing written here is reachable until the
	// pointer moves, and no other writer shares this generation. tokenFields is
	// used rather than ranging the map so the sequence is deterministic, which
	// keeps failures reproducible.
	for _, field := range tokenFields {
		if err := k.backend.Set(k.service, fieldAccount(namespace, generation, field), values[field]); err != nil {
			return fmt.Errorf("write %s to the keychain: %w (is the system keychain unlocked?)", field, err)
		}
	}

	committed, err := json.Marshal(keyringPointer{Current: generation, Previous: previous.Current})
	if err != nil {
		return fmt.Errorf("encode the token pointer: %w", err)
	}

	// The commit.
	if err := k.backend.Set(k.service, pointerAccount(namespace), string(committed)); err != nil {
		return fmt.Errorf("commit the session to the keychain: %w (is the system keychain unlocked?)", err)
	}

	// The generation before the one just replaced is now unreachable by any
	// reader that started before this commit, so it is safe to clear. The one
	// directly replaced is deliberately KEPT until the next Save, because a
	// Load may still be part way through reading it; Load's retry handles the
	// rest. Failure here is not the caller's problem: the session is committed.
	k.clearGeneration(namespace, previous.Previous)

	return nil
}

// clearGeneration removes one generation's field entries, best effort. Callers
// use it for a generation nothing can still be reading.
func (k *keyringStore) clearGeneration(namespace, generation string) {
	if wegostrings.IsBlank(generation) {
		return
	}

	for _, field := range tokenFields {
		_ = k.backend.Delete(k.service, fieldAccount(namespace, generation, field))
	}
}

// Delete removes every entry under namespace. Entries already gone are fine:
// the desired end state is "signed out", so signing out twice must succeed.
//
// The pointer is read first so the generations it names can be cleared, then
// removed BEFORE them: once it is gone the operator is signed out even if a
// later delete fails.
//
// A keychain cannot be enumerated, so only the generations the pointer knows
// about can be reaped. A generation orphaned by a crash between its field
// writes and its commit, or by two Saves overlapping, is therefore not
// reachable here. It holds a superseded token set that no code path will ever
// return, and Cognito refresh tokens expire, so it decays rather than
// accumulating indefinitely. Eliminating that window needs process-shared
// locking, which is a larger change than this seam warrants.
func (k *keyringStore) Delete(namespace string) error {
	if err := k.validate(namespace); err != nil {
		return err
	}

	// A pointer that cannot be parsed must not block signing out; the entry is
	// removed below regardless.
	pointer, _ := k.readPointer(namespace)

	err := k.backend.Delete(k.service, pointerAccount(namespace))
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("delete the token pointer from the keychain: %w (is the system keychain unlocked?)", err)
	}

	for _, generation := range []string{pointer.Current, pointer.Previous} {
		k.clearGeneration(namespace, generation)
	}

	return nil
}

// read fetches one field of one slot, treating a missing entry as an error:
// callers only reach it after the pointer has confirmed a committed session,
// so anything absent is an inconsistency rather than a logged-out state.
func (k *keyringStore) read(namespace, generation, field string) (string, error) {
	value, err := k.backend.Get(k.service, fieldAccount(namespace, generation, field))
	if err != nil {
		return "", fmt.Errorf("read %s from the keychain: %w", field, err)
	}
	return value, nil
}

// validate checks the inputs every operation needs.
func (k *keyringStore) validate(namespace string) error {
	if wegostrings.IsBlank(k.service) {
		return errors.New("keychain service name must not be blank")
	}
	if wegostrings.IsBlank(namespace) {
		return errBlankNamespace
	}
	return nil
}

// pointerAccount is the keychain account name of a namespace's commit pointer.
func pointerAccount(namespace string) string {
	return namespace + "/" + fieldCurrent
}

// fieldAccount is the keychain account name for one field of one generation.
func fieldAccount(namespace, generation, field string) string {
	return namespace + "/" + generation + "/" + field
}
