package ecies

import (
	"crypto/elliptic"
	"encoding/base64"
	"encoding/hex"
	"math/big"

	"github.com/wego/pkg/errors"
)

// PrivateKey ...
type PrivateKey struct {
	Pub *PublicKey
	d   *big.Int
}

// PrivateKeyFromBytes parses a private key from its raw bytes
func PrivateKeyFromBytes(b []byte, curve elliptic.Curve) *PrivateKey {
	x, y := curve.ScalarBaseMult(b)

	return &PrivateKey{
		d: new(big.Int).SetBytes(b),
		Pub: &PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
	}
}

// PrivateKeyFromBase64 parses a private key from its base64 form
func PrivateKeyFromBase64(base64Key string, curve elliptic.Curve) (*PrivateKey, error) {
	b, e := base64.StdEncoding.DecodeString(base64Key)
	if e != nil {
		return nil, errors.New("error decoding base64Key", e)
	}

	return PrivateKeyFromBytes(b, curve), nil
}

// PrivateKeyFromHex parses a private key from its hex form
func PrivateKeyFromHex(hexKey string, curve elliptic.Curve) (*PrivateKey, error) {
	b, e := hex.DecodeString(hexKey)
	if e != nil {
		return nil, errors.New("error decoding hexKey", e)
	}

	return PrivateKeyFromBytes(b, curve), nil
}

// Bytes returns private key raw bytes
func (priv *PrivateKey) Bytes() []byte {
	return priv.d.Bytes()
}

// Base64 returns private key bytes in base64 form
func (priv *PrivateKey) Base64() string {
	return base64.StdEncoding.EncodeToString(priv.Bytes())
}

// Hex returns private key bytes in hex form
func (priv *PrivateKey) Hex() string {
	return hex.EncodeToString(priv.Bytes())
}
