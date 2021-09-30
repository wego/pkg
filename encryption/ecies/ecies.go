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

// Encrypt encrypts plaintext using receiver public key
func Encrypt(plaintext string, pub *PublicKey) (string, error) {
	// generate an ephemeral key pair
	priv, err := GenerateKey(pub.curve)
	if err != nil {
		return "", err
	}

	// compute a shared secret
	key, err := computeEncryptKey(priv, pub)
	if err != nil {
		return "", err
	}

	// encrypt
	encrypted, err := aes.Encrypt(plaintext, key)
	if err != nil {
		return "", err
	}

	return priv.Pub.Hex() + encrypted, nil
}

// Decrypt decrypts ciphertext by receiver private key
func Decrypt(ciphertext string, priv *PrivateKey) (string, error) {
	// check if the ciphertext is long enough
	keyHexSize := hex.EncodedLen(publicKeySize(keySize(priv.Pub.curve)))
	if len(ciphertext) <= keyHexSize {
		return "", fmt.Errorf("ciphertext is too short")
	}

	// parse the public key
	pub, err := PublicKeyFromHex(ciphertext[:keyHexSize], priv.Pub.curve)
	if err != nil {
		return "", err
	}

	// get the shared secret
	key, err := computeDecryptKey(priv, pub)
	if err != nil {
		return "", err
	}

	return aes.Decrypt(ciphertext[keyHexSize:], key)
}

func computeEncryptKey(sender *PrivateKey, receiver *PublicKey) (secret string, err error) {
	x, y := receiver.curve.ScalarMult(receiver.x, receiver.y, sender.d.Bytes())

	var key bytes.Buffer
	key.Write(x.Bytes())
	key.Write(receiver.Bytes())
	key.Write(y.Bytes())

	secret, err = kdf(key.Bytes())
	return
}

func computeDecryptKey(receiver *PrivateKey, sender *PublicKey) (secret string, err error) {
	x, y := sender.curve.ScalarMult(sender.x, sender.y, receiver.d.Bytes())

	var key bytes.Buffer
	key.Write(x.Bytes())
	key.Write(receiver.Pub.Bytes())
	key.Write(y.Bytes())

	secret, err = kdf(key.Bytes())
	return
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

func kdf(secret []byte) (string, error) {
	key := make([]byte, aes.KeyLength)
	kdf := hkdf.New(sha256.New, secret, nil, nil)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return "", fmt.Errorf("cannot read secret from HKDF reader: %w", err)
	}

	return hex.EncodeToString(key), nil
}
