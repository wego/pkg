package storage

import (
	"errors"
	"fmt"

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
	if err := probeKeyring(service); err != nil {
		return NewFile(dir), &Downgrade{Dir: dir, Cause: err}
	}
	return NewKeyring(service), nil
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
