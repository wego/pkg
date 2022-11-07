package ecies

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"

	"github.com/wego/pkg/errors"
)

// Point is a pont on the curve
type Point struct {
	X, Y *big.Int
}

// PublicKey ...
type PublicKey struct {
	curve elliptic.Curve
	*Point
}

const (
	ecPublicKeyBlockType = "EC PUBLIC KEY"
)

// PublicKeyFromBytes parses a public key from its uncompressed raw bytes
func PublicKeyFromBytes(b []byte, curve elliptic.Curve) (*PublicKey, error) {
	size := keySize(curve)
	if len(b) != publicKeySize(size) {
		return nil, fmt.Errorf("invalid key length")
	}

	return &PublicKey{
		curve: curve,
		Point: &Point{
			X: new(big.Int).SetBytes(b[1 : size+1]),
			Y: new(big.Int).SetBytes(b[size+1:])},
	}, nil
}

// PublicKeyFromBase64 parses a public key from its base64 form
func PublicKeyFromBase64(base64Key string, curve elliptic.Curve) (*PublicKey, error) {
	b, e := base64.StdEncoding.DecodeString(base64Key)
	if e != nil {
		return nil, errors.New("error decoding base64Key: %w", e)
	}

	return PublicKeyFromBytes(b, curve)
}

// PublicKeyFromHex parses a public key from its hex form
func PublicKeyFromHex(hexKey string, curve elliptic.Curve) (*PublicKey, error) {
	b, e := hex.DecodeString(hexKey)
	if e != nil {
		return nil, errors.New("error decoding hexKey: %w", e)
	}

	return PublicKeyFromBytes(b, curve)
}

// PublicKeyFromPEMFile parses a public key from a PEM file
func PublicKeyFromPEMFile(path string) (*PublicKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return PublicKeyFromPEMBytes(pemBytes)
}

// PublicKeyFromPEMString parses a public key from a PEM string
func PublicKeyFromPEMString(pemString string) (*PublicKey, error) {
	return PublicKeyFromPEMBytes([]byte(pemString))
}

// PublicKeyFromPEMBytes parses a public key from its PEM form
func PublicKeyFromPEMBytes(bytes []byte) (*PublicKey, error) {
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to parse DER encoded public key: %s", hex.EncodeToString(bytes)), err)
	}

	switch pub.(type) {
	case *ecdsa.PublicKey:
		return &PublicKey{
			curve: pub.(*ecdsa.PublicKey).Curve,
			Point: &Point{
				X: pub.(*ecdsa.PublicKey).X,
				Y: pub.(*ecdsa.PublicKey).Y,
			},
		}, nil
	default:
		return nil, errors.New(fmt.Sprintf("not a EC public key: %T", pub))
	}
}

// Bytes returns the public key to raw bytes in uncompressed format (Ox04|x|y)
// https://secg.org/sec1-v2.pdf#subsubsection.2.3.3
func (pub *PublicKey) Bytes() []byte {
	size := keySize(pub.curve)
	ret := make([]byte, 1+2*size)
	ret[0] = 0x04 // uncompressed point

	// the FillBytes function will pad the bytes with 0s
	pub.X.FillBytes(ret[1 : 1+size])
	pub.Y.FillBytes(ret[1+size : 1+2*size])
	return ret
}

// Base64 returns public key bytes in base64 form
func (pub *PublicKey) Base64() string {
	return base64.StdEncoding.EncodeToString(pub.Bytes())
}

// Hex returns public key bytes in hex form
func (pub *PublicKey) Hex() string {
	return hex.EncodeToString(pub.Bytes())
}

// PEM returns the public key in PEM format - string
func (pub *PublicKey) PEM() string {
	return string(pub.ToPEM())
}

// ToPEM returns the public key in PEM format - bytes
func (pub *PublicKey) ToPEM() []byte {
	x509Encoded, _ := x509.MarshalPKIXPublicKey(&ecdsa.PublicKey{
		Curve: pub.curve,
		X:     pub.X,
		Y:     pub.Y,
	})
	return pem.EncodeToMemory(&pem.Block{Type: ecPublicKeyBlockType, Bytes: x509Encoded})
}

// SaveToPEMFile saves the public key to a PEM file
func (pub *PublicKey) SaveToPEMFile(path string) error {
	return os.WriteFile(path, pub.ToPEM(), 0644)
}

// ScalarMult returns publicKey * privateKey
func (pub *PublicKey) ScalarMult(priv *PrivateKey) *Point {
	x, y := pub.curve.ScalarMult(pub.X, pub.Y, priv.k.Bytes())
	return &Point{x, y}
}
