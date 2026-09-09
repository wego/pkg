// Package storage persists the Cognito token sets obtained by
// github.com/wego/pkg/cognito.
//
// Tokens live under an opaque namespace, e.g. "my-cli/staging". The caller
// chooses the string and the store never interprets it; namespacing is what
// keeps a staging login from overwriting a production one.
//
// Use NewMemory in tests and anywhere a keychain is unavailable; it satisfies
// the same interface as NewKeyring, so a caller swaps one for the other without
// touching its own code.
package storage

import (
	"errors"
	"sync"

	wegostrings "github.com/wego/pkg/strings"

	"github.com/wego/pkg/cognito"
)

// Store persists a TokenSet under an opaque namespace.
//
// A missing entry is NOT an error: Load returns (nil, nil) when nothing is
// stored, because "not logged in" is a normal state. Callers can therefore
// tell it apart from a backend that is genuinely broken, which does return an
// error.
type Store interface {
	Load(namespace string) (*cognito.TokenSet, error)
	Save(namespace string, tokens *cognito.TokenSet) error
	Delete(namespace string) error
}

var (
	errBlankNamespace = errors.New("storage namespace must not be blank")
	errNoTokens       = errors.New("no tokens to save")
)

// memoryStore keeps tokens in process memory only.
type memoryStore struct {
	mu     sync.RWMutex
	tokens map[string]cognito.TokenSet
}

// NewMemory returns a Store that holds tokens for the lifetime of the process
// and nothing longer. It serves tests and callers with no usable keychain; a
// new process always starts signed out.
func NewMemory() Store {
	return &memoryStore{tokens: make(map[string]cognito.TokenSet)}
}

// Load returns the tokens stored under namespace, or (nil, nil) if there are
// none.
func (m *memoryStore) Load(namespace string) (*cognito.TokenSet, error) {
	if wegostrings.IsBlank(namespace) {
		return nil, errBlankNamespace
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	stored, ok := m.tokens[namespace]
	if !ok {
		return nil, nil
	}

	// stored is already a copy of the map value, so a caller mutating the
	// result cannot reach back into the store.
	return &stored, nil
}

// Save writes tokens under namespace, replacing anything already there.
func (m *memoryStore) Save(namespace string, tokens *cognito.TokenSet) error {
	if wegostrings.IsBlank(namespace) {
		return errBlankNamespace
	}
	if tokens == nil {
		return errNoTokens
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[namespace] = *tokens

	return nil
}

// Delete removes the tokens under namespace. Deleting nothing is not an error.
func (m *memoryStore) Delete(namespace string) error {
	if wegostrings.IsBlank(namespace) {
		return errBlankNamespace
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, namespace)

	return nil
}
