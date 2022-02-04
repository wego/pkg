package ecies

import (
	"bytes"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/wego/pkg/encryption/aes"
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

// EncryptStringToBase64 encrypts plaintext to ciphertext in base64 form using receiver public key
func EncryptStringToBase64(plaintext string, pub *PublicKey, ecdh ECDH, kdf KDF) (string, error) {
	bytes, err := Encrypt([]byte(plaintext), pub, ecdh, kdf)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}

// EncryptStringToHex encrypts plaintext to ciphertext in hex form using receiver public key
func EncryptStringToHex(plaintext string, pub *PublicKey, ecdh ECDH, kdf KDF) (string, error) {
	bytes, err := Encrypt([]byte(plaintext), pub, ecdh, kdf)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// DecryptBase64String decrypts ciphertext in base64 form to plaintext by receiver private key
func DecryptBase64String(ciphertext string, priv *PrivateKey, ecdh ECDH, kdf KDF) (string, error) {
	// check if the ciphertext is long enough
	keyBase64Size := base64.StdEncoding.EncodedLen(publicKeySize(keySize(priv.Pub.curve)))
	if len(ciphertext) <= keyBase64Size {
		return "", fmt.Errorf("ciphertext is too short")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", aes.ErrInvalidBase64String
	}

	bytes, err := Decrypt(data, priv, ecdh, kdf)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// DecryptHexString decrypts ciphertext in hex form to plaintext by receiver private key
func DecryptHexString(ciphertext string, priv *PrivateKey, ecdh ECDH, kdf KDF) (string, error) {
	// check if the ciphertext is long enough
	keyHexSize := hex.EncodedLen(publicKeySize(keySize(priv.Pub.curve)))
	if len(ciphertext) <= keyHexSize {
		return "", fmt.Errorf("ciphertext is too short")
	}

	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", aes.ErrInvalidHexString
	}

	bytes, err := Decrypt(data, priv, ecdh, kdf)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Encrypt encrypts data using receiver public key
func Encrypt(data []byte, pub *PublicKey, ecdh ECDH, kdf KDF) ([]byte, error) {
	if ecdh == nil {
		ecdh = defaultEncryptECDH
	}

	if kdf == nil {
		kdf = defaultKDF
	}

	// generate an ephemeral key pair
	priv, err := GenerateKey(pub.curve)
	if err != nil {
		return nil, err
	}

	// compute a shared secret then derive the encryption key
	masterSecret := ecdh(priv, pub)
	key, err := kdf(masterSecret)
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
func Decrypt(data []byte, priv *PrivateKey, ecdh ECDH, kdf KDF) ([]byte, error) {
	if ecdh == nil {
		ecdh = defaultDecryptECDH
	}

	if kdf == nil {
		kdf = defaultKDF
	}

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

	// compute a shared secret then derive the decryption key
	masterSecret := ecdh(priv, pub)
	key, err := kdf(masterSecret)
	if err != nil {
		return nil, err
	}

	return aes.Decrypt(data[pubKeySize:], key)
}
