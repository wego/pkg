package ecies

import (
	"bytes"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"math/big"

	"github.com/wego/pkg/encryption"
	"github.com/wego/pkg/encryption/aes"
	"github.com/wego/pkg/errors"
)

// GenerateKey generates a new elliptic curve key pair
func GenerateKey(curve elliptic.Curve) (*PrivateKey, error) {
	priv, x, y, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{
		k: new(big.Int).SetBytes(priv),
		Pub: &PublicKey{
			curve: curve,
			Point: &Point{
				X: x,
				Y: y},
		},
	}, nil
}

// EncryptToBase64 encrypts data to ciphertext in base64 form using receiver public key
func EncryptToBase64(data []byte, pub *PublicKey, ecdh ECDH, kdf KDF) (string, error) {
	encryptedBytes, err := Encrypt(data, pub, ecdh, kdf)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

// EncryptStringToBase64 encrypts plaintext to ciphertext in base64 form using receiver public key
func EncryptStringToBase64(plaintext string, pub *PublicKey, ecdh ECDH, kdf KDF) (string, error) {
	return EncryptToBase64([]byte(plaintext), pub, ecdh, kdf)
}

// EncryptToHex encrypts data to ciphertext in hex form using receiver public key
func EncryptToHex(data []byte, pub *PublicKey, ecdh ECDH, kdf KDF) (string, error) {
	encryptedBytes, err := Encrypt(data, pub, ecdh, kdf)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(encryptedBytes), nil
}

// EncryptStringToHex encrypts plaintext to ciphertext in hex form using receiver public key
func EncryptStringToHex(plaintext string, pub *PublicKey, ecdh ECDH, kdf KDF) (string, error) {
	return EncryptToHex([]byte(plaintext), pub, ecdh, kdf)
}

// DecryptBase64 decrypts ciphertext in base64 form to raw data([]byte) by receiver private key
func DecryptBase64(ciphertext string, priv *PrivateKey, ecdh ECDH, kdf KDF) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.New(nil, encryption.MsgInvalidBase64String, err)
	}

	return Decrypt(data, priv, ecdh, kdf)
}

// DecryptBase64String decrypts ciphertext in base64 form to plaintext by receiver private key
func DecryptBase64String(ciphertext string, priv *PrivateKey, ecdh ECDH, kdf KDF) (string, error) {
	decryptedBytes, err := DecryptBase64(ciphertext, priv, ecdh, kdf)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
}

// DecryptHex decrypts ciphertext in hex form to raw data([]byte) by receiver private key
func DecryptHex(ciphertext string, priv *PrivateKey, ecdh ECDH, kdf KDF) ([]byte, error) {
	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.New(nil, encryption.MsgInvalidHexString, err)
	}

	return Decrypt(data, priv, ecdh, kdf)
}

// DecryptHexString decrypts ciphertext in hex form to plaintext by receiver private key
func DecryptHexString(ciphertext string, priv *PrivateKey, ecdh ECDH, kdf KDF) (string, error) {
	decryptedBytes, err := DecryptHex(ciphertext, priv, ecdh, kdf)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
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
	_, _ = out.Write(priv.Pub.Bytes())
	_, _ = out.Write(encrypted)
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
		return nil, errors.New(nil, encryption.MsgCiphertextTooShort)
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
