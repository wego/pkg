package ecies

import (
	"bytes"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
)

// PublicKey ...
type PublicKey struct {
	curve elliptic.Curve
	x, y  *big.Int
}

// PublicKeyFromHex parses a public key from its hex form
func PublicKeyFromHex(hexKey string, curve elliptic.Curve) (*PublicKey, error) {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding hexKey: %w", err)
	}

	return PublicKeyFromBytes(b, curve)
}

// PublicKeyFromBytes parses a public key from its uncompressed raw bytes
func PublicKeyFromBytes(b []byte, curve elliptic.Curve) (*PublicKey, error) {
	size := keySize(curve)
	if len(b) != publicKeySize(size) {
		return nil, fmt.Errorf("invalid key length")
	}

	return &PublicKey{
		curve: curve,
		x:     new(big.Int).SetBytes(b[1 : size+1]),
		y:     new(big.Int).SetBytes(b[size+1:]),
	}, nil
}

// Bytes returns the public key to raw bytes in uncompressed format (Ox04|x|y)
// https://secg.org/sec1-v2.pdf#subsubsection.2.3.3
func (pub *PublicKey) Bytes() []byte {
	size := keySize(pub.curve)

	x := zeroPad(pub.x.Bytes(), size)
	y := zeroPad(pub.y.Bytes(), size)

	return bytes.Join([][]byte{{0x04}, x, y}, nil)
}

// Hex returns public key bytes in hex form
func (pub *PublicKey) Hex() string {
	return hex.EncodeToString(pub.Bytes())
}
