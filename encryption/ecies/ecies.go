package ecies

import (
	"bytes"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	"github.com/wego/pkg/encryption/aes"
	"golang.org/x/crypto/hkdf"
)

//GenerateKey generates a new elliptic curve key pair
func GenerateKey(curve elliptic.Curve) (*PrivateKey, error) {
	priv, x, y, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{
		d: new(big.Int).SetBytes(priv),
		Pub: &PublicKey{
			curve: curve,
			x:     x,
			y:     y,
		},
	}, nil
}

// EncryptString encrypts plaintext to ciphertext in hex form using receiver public key
func EncryptString(plaintext string, pub *PublicKey) (string, error) {
	bytes, err := Encrypt([]byte(plaintext), pub)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// DecryptString decrypts ciphertext in hex form to plaintext by receiver private key
func DecryptString(ciphertext string, priv *PrivateKey) (string, error) {
	// check if the ciphertext is long enough
	keyHexSize := hex.EncodedLen(publicKeySize(keySize(priv.Pub.curve)))
	if len(ciphertext) <= keyHexSize {
		return "", fmt.Errorf("ciphertext is too short")
	}

	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", aes.ErrInvalidHexString
	}

	bytes, err := Decrypt(data, priv)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Encrypt encrypts data using receiver public key
func Encrypt(data []byte, pub *PublicKey) ([]byte, error) {
	// generate an ephemeral key pair
	priv, err := GenerateKey(pub.curve)
	if err != nil {
		return nil, err
	}

	// compute a shared secret
	key, err := computeEncryptKey(priv, pub)
	if err != nil {
		return nil, err
	}

	// encrypt
	encrypted, err := aes.Encrypt(data, key)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	out.Write(priv.Pub.Bytes())
	out.Write(encrypted)
	return out.Bytes(), nil
}

// Decrypt decrypts ciphertext by receiver private key
func Decrypt(data []byte, priv *PrivateKey) ([]byte, error) {
	// check if the ciphertext is long enough
	pubKeySize := publicKeySize(keySize(priv.Pub.curve))
	if len(data) <= pubKeySize {
		return nil, fmt.Errorf("data is too short")
	}

	// parse the public key
	pub, err := PublicKeyFromBytes(data[:pubKeySize], priv.Pub.curve)
	if err != nil {
		return nil, err
	}

	// get the shared secret
	key, err := computeDecryptKey(priv, pub)
	if err != nil {
		return nil, err
	}

	return aes.Decrypt(data[pubKeySize:], key)
}

func computeEncryptKey(sender *PrivateKey, receiver *PublicKey) ([]byte, error) {
	x, y := receiver.curve.ScalarMult(receiver.x, receiver.y, sender.d.Bytes())

	var key bytes.Buffer
	key.Write(x.Bytes())
	key.Write(receiver.Bytes())
	key.Write(y.Bytes())

	return kdf(key.Bytes())
}

func computeDecryptKey(receiver *PrivateKey, sender *PublicKey) ([]byte, error) {
	x, y := sender.curve.ScalarMult(sender.x, sender.y, receiver.d.Bytes())

	var key bytes.Buffer
	key.Write(x.Bytes())
	key.Write(receiver.Pub.Bytes())
	key.Write(y.Bytes())

	return kdf(key.Bytes())
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

func kdf(secret []byte) ([]byte, error) {
	key := make([]byte, aes.KeyLength)
	kdf := hkdf.New(sha256.New, secret, nil, nil)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return nil, fmt.Errorf("cannot read secret from HKDF reader: %w", err)
	}

	return key, nil
}
