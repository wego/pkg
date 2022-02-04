package ecies

import (
	"bytes"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
)

// PublicKey ...
type PublicKey struct {
	curve elliptic.Curve
	x, y  *big.Int
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

// PublicKeyFromBase64 parses a public key from its base64 form
func PublicKeyFromBase64(base64Key string, curve elliptic.Curve) (*PublicKey, error) {
	b, e := base64.StdEncoding.DecodeString(base64Key)
	if e != nil {
		return nil, fmt.Errorf("error decoding base64Key: %w", e)
	}

	return PublicKeyFromBytes(b, curve)
}

// PublicKeyFromHex parses a public key from its hex form
func PublicKeyFromHex(hexKey string, curve elliptic.Curve) (*PublicKey, error) {
	b, e := hex.DecodeString(hexKey)
	if e != nil {
		return nil, fmt.Errorf("error decoding hexKey: %w", e)
	}

	return PublicKeyFromBytes(b, curve)
}

// Bytes returns the public key to raw bytes in uncompressed format (Ox04|x|y)
// https://secg.org/sec1-v2.pdf#subsubsection.2.3.3
func (pub *PublicKey) Bytes() []byte {
	size := keySize(pub.curve)

	x := zeroPad(pub.x.Bytes(), size)
	y := zeroPad(pub.y.Bytes(), size)

	return bytes.Join([][]byte{{0x04}, x, y}, nil)
}

// Base64 returns public key bytes in base64 form
func (pub *PublicKey) Base64() string {
	return base64.StdEncoding.EncodeToString(pub.Bytes())
}

// Hex returns public key bytes in hex form
func (pub *PublicKey) Hex() string {
	return hex.EncodeToString(pub.Bytes())
}
