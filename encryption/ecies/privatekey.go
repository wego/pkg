package ecies

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/wego/pkg/errors"
	"math/big"
	"os"
)

// PrivateKey ...
type PrivateKey struct {
	Pub *PublicKey
	k   *big.Int
}

const (
	ecPrivateKeyBlockType = "EC PRIVATE KEY"
)

// PrivateKeyFromBytes parses a private key from its raw bytes
func PrivateKeyFromBytes(b []byte, curve elliptic.Curve) *PrivateKey {
	x, y := curve.ScalarBaseMult(b)

	return &PrivateKey{
		k: new(big.Int).SetBytes(b),
		Pub: &PublicKey{
			curve: curve,
			Point: &Point{
				X: x,
				Y: y,
			},
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

// PrivateKeyFromPEMBytes parses a private key from a PEM bytes
func PrivateKeyFromPEMBytes(bytes []byte) (*PrivateKey, error) {
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to parse DER encoded private key: %s", hex.EncodeToString(bytes)), err)
	}

	return &PrivateKey{
		Pub: &PublicKey{
			curve: priv.Curve,
			Point: &Point{
				X: priv.X,
				Y: priv.Y,
			},
		},
		k: priv.D,
	}, nil
}

// PrivateKeyFromPEMString parses a private key from a PEM string
func PrivateKeyFromPEMString(str string) (*PrivateKey, error) {
	return PrivateKeyFromPEMBytes([]byte(str))
}

// PrivateKeyFromPEMFile parses a private key from a PEM file
func PrivateKeyFromPEMFile(file string) (*PrivateKey, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read pem file: %s", file), err)
	}

	return PrivateKeyFromPEMBytes(bytes)
}

// Bytes returns private key raw bytes
func (priv *PrivateKey) Bytes() []byte {
	return priv.k.Bytes()
}

// Base64 returns private key bytes in base64 form
func (priv *PrivateKey) Base64() string {
	return base64.StdEncoding.EncodeToString(priv.Bytes())
}

// Hex returns private key bytes in hex form
func (priv *PrivateKey) Hex() string {
	return hex.EncodeToString(priv.Bytes())
}

// PEM returns private key in PEM form - string
func (priv *PrivateKey) PEM() string {
	return string(priv.ToPEM())
}

// ToPEM returns private key in PEM form - bytes
func (priv *PrivateKey) ToPEM() []byte {
	x509Encoded, _ := x509.MarshalECPrivateKey(&ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: priv.Pub.curve,
			X:     priv.Pub.X,
			Y:     priv.Pub.Y,
		},
		D: priv.k,
	})
	return pem.EncodeToMemory(&pem.Block{Type: ecPrivateKeyBlockType, Bytes: x509Encoded})
}

// SavePEMFile saves private key to a PEM file
func (priv *PrivateKey) SavePEMFile(file string) error {
	return os.WriteFile(file, priv.ToPEM(), 0644)
}
