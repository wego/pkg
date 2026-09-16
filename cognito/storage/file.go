package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	wegostrings "github.com/wego/pkg/strings"

	"github.com/wego/pkg/cognito"
)

const (
	// storeDirMode keeps the directory owner-only. A group- or world-readable
	// directory would let another account enumerate which environments an
	// operator has sessions for even when it cannot read the files.
	storeDirMode fs.FileMode = 0o700
	// storeFileMode keeps each session owner-only. The refresh token inside is
	// long-lived, so this is the whole protection the file store offers.
	storeFileMode fs.FileMode = 0o600

	sessionSuffix = ".json"
)

// fileStore persists sessions as one JSON file per namespace.
//
// This is the fallback for a host with no OS keychain - a headless Linux box
// over SSH, where there is no org.freedesktop.secrets D-Bus service to talk to.
// It is a deliberate downgrade: the keychain encrypts at rest and gates access
// on the login session, whereas this is plaintext protected only by file mode,
// and anything running as the same user can read it. Prefer NewKeyring where a
// keychain exists; NewAuto picks for you and says when it had to fall back.
type fileStore struct {
	dir string
}

// NewFile returns a Store writing one owner-only JSON file per namespace under
// dir. The directory is created on first write, not here, so constructing a
// store this process never writes to touches no disk.
//
// The owner-only guarantee is a Unix one. Wego runs this on macOS workstations
// and Linux hosts, where 0600 and 0700 mean what they say. Go maps those bits
// onto Windows ACLs only approximately, so on Windows the file would be created
// but its protection would not be the one documented here - and NewAuto would
// not choose this store there anyway, since Windows has a working credential
// manager for go-keyring to use.
func NewFile(dir string) Store {
	return &fileStore{dir: dir}
}

// DefaultFileDir is where NewAuto puts sessions when it falls back. It follows
// the XDG state convention, which is the right home for data that persists
// between runs but is not configuration the operator edits.
func DefaultFileDir(service string) (string, error) {
	if state := os.Getenv("XDG_STATE_HOME"); wegostrings.IsNotBlank(state) {
		return filepath.Join(state, service), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate the home directory for the session store: %w", err)
	}
	return filepath.Join(home, ".local", "state", service), nil
}

// path maps a namespace to its file.
//
// A namespace is opaque and routinely contains a separator ("pay-admin/staging"),
// so it cannot be used as a filename directly: a caller could otherwise pick a
// namespace that escapes dir entirely. The name is a hash of the namespace with
// a readable prefix - the prefix so a human can tell the files apart, the hash
// so the mapping is total and injective for any input.
func (f *fileStore) path(namespace string) string {
	sum := sha256.Sum256([]byte(namespace))

	readable := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		default:
			return '-'
		}
	}, namespace)
	readable = strings.Trim(readable, "-")
	if len(readable) > 32 {
		readable = readable[:32]
	}
	if readable == "" {
		readable = "session"
	}

	return filepath.Join(f.dir, readable+"-"+hex.EncodeToString(sum[:8])+sessionSuffix)
}

// Load returns the session stored for namespace, or (nil, nil) when there is
// none - "not logged in" is a normal state, not an error.
func (f *fileStore) Load(namespace string) (*cognito.TokenSet, error) {
	if wegostrings.IsBlank(namespace) {
		return nil, errBlankNamespace
	}

	raw, err := os.ReadFile(f.path(namespace))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read the stored session: %w", err)
	}

	var tokens cognito.TokenSet
	if err := json.Unmarshal(raw, &tokens); err != nil {
		return nil, fmt.Errorf(
			"parse the stored session for %q: %w (the file is unreadable - sign out of this environment and sign in again to rewrite it)",
			namespace, err)
	}

	// Same completeness rule the keyring store applies: a session missing any
	// of the three tokens is an inconsistency, not a logged-out state, and
	// silently treating it as absent would hide a real problem.
	if wegostrings.IsBlank(tokens.AccessToken) || wegostrings.IsBlank(tokens.IDToken) || wegostrings.IsBlank(tokens.RefreshToken) {
		return nil, fmt.Errorf(
			"the stored session for %q is incomplete (sign out of this environment and sign in again to rewrite it)",
			namespace)
	}

	return &tokens, nil
}

// Save writes the session atomically.
//
// Write-then-rename rather than writing in place: a crash or a full disk
// half-way through an in-place write would leave a truncated file that Load
// then rejects, turning a transient failure into a session the operator has to
// repair by hand. rename(2) on the same filesystem is atomic, so a reader sees
// either the old session or the new one.
func (f *fileStore) Save(namespace string, tokens *cognito.TokenSet) error {
	if wegostrings.IsBlank(namespace) {
		return errBlankNamespace
	}
	if tokens == nil {
		return errNoTokens
	}

	if err := os.MkdirAll(f.dir, storeDirMode); err != nil {
		return fmt.Errorf("create the session directory %s: %w", f.dir, err)
	}
	// MkdirAll applies its mode only to directories it creates, so a directory
	// that already exists keeps whatever permissions it had - from an earlier
	// release, a permissive umask, or another tool. Tighten it every time
	// rather than trusting how it came to exist: the file mode stops others
	// reading a session, but a listable directory still discloses which
	// environments this operator holds sessions for.
	if err := os.Chmod(f.dir, storeDirMode); err != nil {
		return fmt.Errorf("restrict permissions on the session directory %s: %w", f.dir, err)
	}

	encoded, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("encode the session: %w", err)
	}

	final := f.path(namespace)

	// Created in the destination directory so the rename cannot cross a
	// filesystem boundary, and with the final mode from the start - a
	// world-readable temp file would expose the refresh token for the window
	// before a chmod, however short.
	temp, err := os.CreateTemp(f.dir, ".session-*")
	if err != nil {
		return fmt.Errorf("create a temporary session file in %s: %w", f.dir, err)
	}
	tempName := temp.Name()
	defer func() { _ = os.Remove(tempName) }() // no-op once the rename succeeds

	if err := temp.Chmod(storeFileMode); err != nil {
		_ = temp.Close()
		return fmt.Errorf("restrict permissions on the session file: %w", err)
	}
	if _, err := temp.Write(encoded); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write the session: %w", err)
	}
	// Flush to disk before the rename. Without this the rename can be durable
	// while the contents are not, which on a crash leaves an empty file where a
	// valid session used to be.
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("flush the session to disk: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close the session file: %w", err)
	}

	if err := os.Rename(tempName, final); err != nil {
		return fmt.Errorf("commit the session to %s: %w", final, err)
	}

	return nil
}

// Delete removes the stored session. A namespace with nothing stored is not an
// error, matching the keyring store.
func (f *fileStore) Delete(namespace string) error {
	if wegostrings.IsBlank(namespace) {
		return errBlankNamespace
	}

	if err := os.Remove(f.path(namespace)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("delete the stored session: %w", err)
	}
	return nil
}
