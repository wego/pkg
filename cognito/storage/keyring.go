package storage

import (
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
// transaction. Instead each token set is written into one of two SLOTS, and a
// separate pointer entry names the slot that counts. Save fills the inactive
// slot and then moves the pointer, which is one small write and the only write
// that changes what Load sees. A failure anywhere before it leaves the pointer
// and the live slot untouched, so a torn write costs the new session, never
// the old one.
//
// Two slots rather than a counter keeps the entry count fixed and means a
// re-login never writes over the entries the current session is read from.
//
// Entries are accounted as "<namespace>/current" for the pointer and
// "<namespace>/<slot>/<field>" for the token fields.
const (
	fieldAccess  = "access"
	fieldID      = "id"
	fieldRefresh = "refresh"
	fieldMeta    = "meta"
	// fieldCurrent is the pointer entry naming the live slot. Committing a
	// token set is exactly one write of this entry.
	fieldCurrent = "current"
)

// The two slots a token set alternates between.
const (
	slotA = "a"
	slotB = "b"
)

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
		slot, err := k.liveSlot(namespace)
		if err != nil || slot == "" {
			// No pointer means not signed in, which is (nil, nil).
			return nil, err
		}

		tokens, err := k.readSlot(namespace, slot)
		if err == nil {
			return tokens, nil
		}

		// The read failed. If the pointer has moved since it was chosen, this
		// slot was superseded and cleared under us, so try the new one. If it
		// has not, the namespace really is inconsistent and the original error
		// is the honest one to report.
		moved, checkErr := k.slotMoved(namespace, slot)
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

// liveSlot returns the slot the pointer names, or "" when nothing is stored.
func (k *keyringStore) liveSlot(namespace string) (string, error) {
	slot, err := k.backend.Get(k.service, pointerAccount(namespace))
	if errors.Is(err, keyring.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read the keychain: %w (is the system keychain unlocked?)", err)
	}

	if slot != slotA && slot != slotB {
		// Nothing here writes any other value, so this entry was tampered with
		// or written by a version that stored something else. Guessing which
		// slot was meant would be worse than refusing.
		return "", fmt.Errorf("keychain names an unknown token slot %q for %q", slot, namespace)
	}

	return slot, nil
}

// slotMoved reports whether the pointer now names a slot other than the one a
// read was using.
func (k *keyringStore) slotMoved(namespace, slot string) (bool, error) {
	current, err := k.liveSlot(namespace)
	if err != nil {
		return false, err
	}

	return current != slot, nil
}

// readSlot assembles the token set held in one slot.
func (k *keyringStore) readSlot(namespace, slot string) (*cognito.TokenSet, error) {
	access, err := k.read(namespace, slot, fieldAccess)
	if err != nil {
		return nil, err
	}
	idToken, err := k.read(namespace, slot, fieldID)
	if err != nil {
		return nil, err
	}
	refresh, err := k.read(namespace, slot, fieldRefresh)
	if err != nil {
		return nil, err
	}
	rawMeta, err := k.read(namespace, slot, fieldMeta)
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
// It is atomic from Load's point of view: the fields go into the slot that is
// not live, and only the final pointer write makes them the session. If any
// write fails, the previous session is still whole and still what Load
// returns.
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

	live, err := k.currentSlot(namespace)
	if err != nil {
		return err
	}

	next := otherSlot(live)

	values := map[string]string{
		fieldAccess:  tokens.AccessToken,
		fieldID:      tokens.IDToken,
		fieldRefresh: tokens.RefreshToken,
		fieldMeta:    string(meta),
	}

	// Ordering is irrelevant now: nothing written here is reachable until the
	// pointer moves. tokenFields is used rather than ranging the map so the
	// write sequence is deterministic, which keeps failures reproducible.
	for _, field := range tokenFields {
		if err := k.backend.Set(k.service, fieldAccount(namespace, next, field), values[field]); err != nil {
			return fmt.Errorf("write %s to the keychain: %w (is the system keychain unlocked?)", field, err)
		}
	}

	// The commit.
	if err := k.backend.Set(k.service, pointerAccount(namespace), next); err != nil {
		return fmt.Errorf("commit the session to the keychain: %w (is the system keychain unlocked?)", err)
	}

	// The old slot is now unreachable. Clearing it keeps a superseded token set
	// out of the keychain, but the session is already committed, so a failure
	// here is not the caller's problem and must not fail the sign-in.
	k.clearSlot(namespace, live)

	return nil
}

// currentSlot reports the live slot, or "" when nothing is stored. An
// unrecognised value is treated as "nothing live": Save's job is to establish a
// good session, and it can do that without deciding what the bad value meant.
// Load is where a corrupt pointer is reported.
func (k *keyringStore) currentSlot(namespace string) (string, error) {
	slot, err := k.backend.Get(k.service, pointerAccount(namespace))
	if errors.Is(err, keyring.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read the current token slot: %w (is the system keychain unlocked?)", err)
	}

	if slot != slotA && slot != slotB {
		return "", nil
	}

	return slot, nil
}

// otherSlot returns the slot to write next. An empty live slot means nothing is
// stored, so either is free and slotA keeps a first login predictable.
func otherSlot(live string) string {
	if live == slotA {
		return slotB
	}

	return slotA
}

// clearSlot removes one slot's field entries, best effort. Callers use it for a
// slot nothing points at any more.
func (k *keyringStore) clearSlot(namespace, slot string) {
	if slot == "" {
		return
	}

	for _, field := range tokenFields {
		_ = k.backend.Delete(k.service, fieldAccount(namespace, slot, field))
	}
}

// Delete removes every entry under namespace. Entries already gone are fine:
// the desired end state is "signed out", so signing out twice must succeed.
//
// The pointer goes FIRST, for the same reason Save moves it last: once it is
// gone the operator is signed out, even if a later delete fails and leaves
// orphaned field entries behind.
func (k *keyringStore) Delete(namespace string) error {
	if err := k.validate(namespace); err != nil {
		return err
	}

	err := k.backend.Delete(k.service, pointerAccount(namespace))
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("delete the current token slot from the keychain: %w (is the system keychain unlocked?)", err)
	}

	// Both slots, not just the live one: a torn Save can leave fields in the
	// inactive slot, and signing out should not leave a token set behind.
	for _, slot := range []string{slotA, slotB} {
		for _, field := range tokenFields {
			err := k.backend.Delete(k.service, fieldAccount(namespace, slot, field))
			if err != nil && !errors.Is(err, keyring.ErrNotFound) {
				return fmt.Errorf("delete %s from the keychain: %w (is the system keychain unlocked?)", field, err)
			}
		}
	}

	return nil
}

// read fetches one field of one slot, treating a missing entry as an error:
// callers only reach it after the pointer has confirmed a committed session,
// so anything absent is an inconsistency rather than a logged-out state.
func (k *keyringStore) read(namespace, slot, field string) (string, error) {
	value, err := k.backend.Get(k.service, fieldAccount(namespace, slot, field))
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

// fieldAccount is the keychain account name for one field of one slot.
func fieldAccount(namespace, slot, field string) string {
	return namespace + "/" + slot + "/" + field
}
