package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/wego/pkg/encryption"
	"github.com/wego/pkg/errors"
)

// KeyLength is the min length of the secret key
const KeyLength = 32

var (
	// ErrShortKey key is too short
	ErrShortKey = fmt.Errorf("key is too short, %d bytes is required", KeyLength)
)

// EncryptToHex encrypts data to ciphertext (hex form) using 256-bit AES-GCM, key must have length 32 or more
func EncryptToHex(data []byte, key string) (string, error) {
	bytes, err := Encrypt(data, []byte(key))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// EncryptStringToHex encrypts plaintext to ciphertext (hex form) using 256-bit AES-GCM, key must have length 32 or more
func EncryptStringToHex(plaintext, key string) (string, error) {
	return EncryptToHex([]byte(plaintext), key)
}

// DecryptHex decrypts a hex form ciphertext to raw data([]byte) using 256-bit AES-GCM, key must have length 32 or more
func DecryptHex(ciphertext, key string) ([]byte, error) {
	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.New(encryption.MsgInvalidHexString, err)
	}

	return Decrypt(data, []byte(key))
}

// DecryptHexString decrypts a hex form ciphertext to the plaintext using 256-bit AES-GCM, key must have length 32 or more
func DecryptHexString(ciphertext, key string) (string, error) {
	bytes, err := DecryptHex(ciphertext, key)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// EncryptToBase64 encrypts data to ciphertext (base64 form) using 256-bit AES-GCM, key must have length 32 or more
func EncryptToBase64(data []byte, key string) (string, error) {
	bytes, err := Encrypt(data, []byte(key))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}

// EncryptStringToBase64 encrypts plaintext to ciphertext (base64 form) using 256-bit AES-GCM, key must have length 32 or more
func EncryptStringToBase64(plaintext, key string) (string, error) {
	return EncryptToBase64([]byte(plaintext), key)
}

// DecryptBase64 decrypts a base64 form ciphertext to raw data([]byte) using 256-bit AES-GCM, key must have length 32 or more
func DecryptBase64(ciphertext, key string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.New(encryption.MsgInvalidBase64String, err)
	}

	return Decrypt(data, []byte(key))
}

// DecryptBase64String decrypts a base64 form ciphertext to the plaintext using 256-bit AES-GCM, key must have length 32 or more
func DecryptBase64String(ciphertext, key string) (string, error) {
	bytes, err := DecryptBase64(ciphertext, key)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Encrypt encrypts data using 256-bit AES-GCM, key must have length 32 or more
func Encrypt(data, key []byte) ([]byte, error) {
	if len(key) < KeyLength {
		return nil, ErrShortKey
	}

	block, err := aes.NewCipher(key[:KeyLength])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// Decrypt decrypts data using 256-bit AES-GCM, key must have length 32 or more
func Decrypt(data, key []byte) ([]byte, error) {
	if len(key) < KeyLength {
		return nil, ErrShortKey
	}

	block, err := aes.NewCipher(key[:KeyLength])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, errors.New(encryption.MsgCiphertextTooShort)
	}

	return gcm.Open(nil,
		data[:gcm.NonceSize()],
		data[gcm.NonceSize():],
		nil,
	)
}
