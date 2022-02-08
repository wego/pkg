package ecies

import (
	"bytes"
	"crypto/elliptic"
	"crypto/sha256"
	"io"

	"github.com/wego/pkg/encryption/aes"
	"github.com/wego/pkg/errors"
	"golang.org/x/crypto/hkdf"
)

// KDF accepts a master secret & derives a encryption key
type KDF func(masterSecret []byte) (key []byte, err error)

// ECDH calculates a master secret from a private key & a public key.
// Its output will be passed to a KDF for deriving the encryption key.
type ECDH func(priv *PrivateKey, pub *PublicKey) (masterSecret []byte)

func defaultEncryptECDH(sender *PrivateKey, receiver *PublicKey) []byte {
	x, y := receiver.Curve.ScalarMult(receiver.x, receiver.y, sender.d.Bytes())

	var key bytes.Buffer
	key.Write(x.Bytes())
	key.Write(receiver.Bytes())
	key.Write(y.Bytes())

	return key.Bytes()
}

func defaultDecryptECDH(receiver *PrivateKey, sender *PublicKey) []byte {
	x, y := sender.Curve.ScalarMult(sender.x, sender.y, receiver.d.Bytes())

	var key bytes.Buffer
	key.Write(x.Bytes())
	key.Write(receiver.Pub.Bytes())
	key.Write(y.Bytes())

	return key.Bytes()
}

func defaultKDF(secret []byte) ([]byte, error) {
	key := make([]byte, aes.KeyLength)
	kdf := hkdf.New(sha256.New, secret, nil, nil)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return nil, errors.New("cannot read secret from HKDF reader", err)
	}

	return key, nil
}

func keySize(curve elliptic.Curve) int {
	bitSize := curve.Params().BitSize

	size := bitSize / 8
	if bitSize%8 > 0 {
		size++
	}
	return size
}

func publicKeySize(keySize int) int {
	return keySize*2 + 1
}

func zeroPad(b []byte, length int) []byte {
	if len(b) < length {
		b = append(make([]byte, length-len(b)), b...)
	}

	return b
}
