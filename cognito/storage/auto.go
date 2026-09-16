package storage

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/wego/pkg/cognito"

	"github.com/zalando/go-keyring"
)

// probeAccount is the entry NewAuto reads to decide whether the OS keychain is
// usable. Nothing ever writes it, so a working keychain answers ErrNotFound.
const probeAccount = "__availability_probe__"

// Downgrade explains why NewAuto chose the file store over the keychain.
//
// It is returned rather than logged because a library has no business writing
// to a caller's stderr, and the caller is the only one that knows whether this
// deserves a warning, a one-line note, or silence.
type Downgrade struct {
	// Dir is where sessions are now written.
	Dir string
	// Cause is the keychain error that forced the fallback.
	Cause error
}

func (d *Downgrade) Error() string {
	return fmt.Sprintf("no usable OS keychain (%v); sessions are stored in %s", d.Cause, d.Dir)
}

func (d *Downgrade) Unwrap() error { return d.Cause }

// NewAuto returns the OS keychain store where one is usable, and a file store
// under dir where it is not, along with a non-nil *Downgrade saying why.
//
// The fallback exists because the keychain is simply absent on a headless
// Linux host: there is no org.freedesktop.secrets D-Bus service to talk to, so
// every write fails and the sign-in an operator just completed is thrown away
// at the last step. That is precisely the host where the paste-back flow is
// most needed.
//
// It is a real downgrade and the caller should say so out loud - see
// fileStore's doc comment. A caller that would rather fail than write a
// long-lived refresh token to disk should use NewKeyring directly.
func NewAuto(service, dir string) (Store, *Downgrade) {
	file := NewFile(dir)

	if err := probeKeyring(service); err != nil {
		// Only downgrade where the file store's protection actually holds. Its
		// guarantee is 0600/0700, which is a Unix guarantee; Go maps those bits
		// onto Windows ACLs only approximately, so falling back there would
		// write a long-lived refresh token to a file whose protection we have
		// not verified and have documented as Unix-only. Windows has a working
		// credential manager for go-keyring, so a probe failure there is a real
		// problem to surface rather than to route around: return the keychain
		// store and let its own error speak.
		if !fileStoreProtects() {
			return NewKeyring(service), nil
		}

		// Keychain unusable: read and write the file, but still try to clear
		// the keychain on Delete for the mirror-image case - a session written
		// on a host that had a keychain, being signed out from one that does
		// not.
		return &autoStore{primary: file, secondary: NewKeyring(service)}, &Downgrade{Dir: dir, Cause: err}
	}

	return &autoStore{primary: NewKeyring(service), secondary: file}, nil
}

// autoStore reads and writes one backend but deletes from both.
//
// Backend selection is per process and depends on whether the keychain happens
// to be reachable right now, so a session can outlive the condition that chose
// where it went. Without this, signing out was only as durable as the current
// selection: save while the keychain is down, sign out after it returns, lose
// the keychain again, and Load hands back the session the operator believed
// they had deleted - with a live refresh token in it. Deleting from both makes
// sign-out mean signed out, whichever backend is selected at the time.
//
// Only Delete fans out. Load and Save deliberately stay on the selected
// backend: reading both would make precedence ambiguous when the two disagree,
// and writing both would put the credential in the weaker store even on hosts
// that have a keychain.
type autoStore struct {
	primary   Store
	secondary Store
}

func (a *autoStore) Load(namespace string) (*cognito.TokenSet, error) {
	return a.primary.Load(namespace)
}

func (a *autoStore) Save(namespace string, tokens *cognito.TokenSet) error {
	return a.primary.Save(namespace, tokens)
}

// Delete clears the session from both backends.
//
// The primary's error is what the caller hears, since that is the backend they
// are working with. A secondary failure is reported only when the primary
// succeeded, because otherwise the primary error is the more useful one and
// the caller is going to retry anyway. Either way both are attempted: a
// half-done sign-out is the failure this method exists to prevent.
func (a *autoStore) Delete(namespace string) error {
	primaryErr := a.primary.Delete(namespace)
	secondaryErr := a.secondary.Delete(namespace)

	if primaryErr != nil {
		return primaryErr
	}
	if secondaryErr != nil {
		return fmt.Errorf("the session was cleared, but the inactive store could not be cleared too: %w", secondaryErr)
	}
	return nil
}

// probeKeyring reports whether the OS keychain can be reached.
//
// A read, not a write: reaching the backend is the whole question, and a probe
// that wrote would leave a stray entry behind and need cleaning up. ErrNotFound
// is the success case - it means the backend answered, and answered that this
// entry does not exist, which is exactly what we expect since nothing writes it.
//
// go-keyring exposes no sentinel for "service unavailable" - a missing D-Bus
// secrets service surfaces as an opaque error string - so anything that is not
// ErrNotFound is treated as unusable. That is the conservative direction: a
// keychain we cannot read is one we should not rely on either.
func probeKeyring(service string) error {
	_, err := keyring.Get(service, probeAccount)
	switch {
	case err == nil, errors.Is(err, keyring.ErrNotFound):
		return nil
	default:
		return err
	}
}

// fileStoreProtects reports whether this platform enforces the file store's
// documented owner-only guarantee.
//
// An allow-list rather than a deny-list: a platform nobody has checked should
// not silently inherit permission to hold a long-lived credential in a file.
func fileStoreProtects() bool {
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd", "openbsd", "netbsd", "dragonfly":
		return true
	default:
		return false
	}
}
