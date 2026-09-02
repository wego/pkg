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

// A Cognito JWT - the access token especially - can exceed the 4 KiB
// command-line cap zalando/go-keyring runs into on macOS, where it shells out
// to /usr/bin/security. So each field gets its own keychain entry, keeping
// every individual write well under that ceiling.
//
// Entries are accounted as "<namespace>/<field>".
const (
	fieldAccess  = "access"
	fieldID      = "id"
	fieldRefresh = "refresh"
	fieldMeta    = "meta"
)

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
func (k *keyringStore) Load(namespace string) (*cognito.TokenSet, error) {
	if err := k.validate(namespace); err != nil {
		return nil, err
	}

	// Gate on the access token: its absence means "not signed in", which is a
	// normal state rather than a failure. Once it is present, anything else
	// missing is a real inconsistency and must be reported.
	access, err := k.backend.Get(k.service, account(namespace, fieldAccess))
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read the keychain: %w (is the system keychain unlocked?)", err)
	}

	idToken, err := k.read(namespace, fieldID)
	if err != nil {
		return nil, err
	}
	refresh, err := k.read(namespace, fieldRefresh)
	if err != nil {
		return nil, err
	}
	rawMeta, err := k.read(namespace, fieldMeta)
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

	// The access token is written LAST because Load gates on it: a write cut
	// short partway through then reads as "not signed in", which is
	// recoverable, rather than as a half-populated token set, which is not.
	entries := []struct {
		field string
		value string
	}{
		{field: fieldRefresh, value: tokens.RefreshToken},
		{field: fieldID, value: tokens.IDToken},
		{field: fieldMeta, value: string(meta)},
		{field: fieldAccess, value: tokens.AccessToken},
	}

	for _, entry := range entries {
		if err := k.backend.Set(k.service, account(namespace, entry.field), entry.value); err != nil {
			return fmt.Errorf("write %s to the keychain: %w (is the system keychain unlocked?)", entry.field, err)
		}
	}

	return nil
}

// Delete removes every entry under namespace. Entries already gone are fine:
// the desired end state is "signed out", so signing out twice must succeed.
func (k *keyringStore) Delete(namespace string) error {
	if err := k.validate(namespace); err != nil {
		return err
	}

	for _, field := range []string{fieldAccess, fieldID, fieldRefresh, fieldMeta} {
		err := k.backend.Delete(k.service, account(namespace, field))
		if err != nil && !errors.Is(err, keyring.ErrNotFound) {
			return fmt.Errorf("delete %s from the keychain: %w (is the system keychain unlocked?)", field, err)
		}
	}

	return nil
}

// read fetches one entry, treating a missing one as an error: callers only
// reach it after the access-token gate has confirmed a login exists.
func (k *keyringStore) read(namespace, field string) (string, error) {
	value, err := k.backend.Get(k.service, account(namespace, field))
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

// account is the keychain account name for one field of one namespace.
func account(namespace, field string) string {
	return namespace + "/" + field
}
