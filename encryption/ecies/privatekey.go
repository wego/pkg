package ecies

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
)

// PrivateKey ...
type PrivateKey struct {
	Pub *PublicKey
	d   *big.Int
}

// PrivateKeyFromHex parses a private key from its hex form
func PrivateKeyFromHex(hexKey string, curve elliptic.Curve) (*PrivateKey, error) {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding hexKey: %w", err)
	}

	return PrivateKeyFromBytes(b, curve), nil
}

// PrivateKeyFromBytes parses a private key from its raw bytes
func PrivateKeyFromBytes(b []byte, curve elliptic.Curve) *PrivateKey {
	x, y := curve.ScalarBaseMult(b)

	return &PrivateKey{
		d: new(big.Int).SetBytes(b),
		Pub: &PublicKey{
			curve: curve,
			x:     x,
			y:     y,
		},
	}
}

// Bytes returns private key raw bytes
func (priv *PrivateKey) Bytes() []byte {
	return priv.d.Bytes()
}

// Hex returns private key bytes in hex form
func (priv *PrivateKey) Hex() string {
	return hex.EncodeToString(priv.Bytes())
}
