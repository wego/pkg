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

// error
var (
	ErrShortKey = fmt.Errorf("key is too short, %d bytes is required", KeyLength)
)

// EncryptStringToHex encrypts plaintext to ciphertext (hex form) using 256-bit AES-GCM, key must have length 32 or more
func EncryptStringToHex(plaintext, key string) (string, error) {
	bytes, err := Encrypt([]byte(plaintext), []byte(key))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// DecryptHexString decrypts a hex form ciphertext to the plaintext using 256-bit AES-GCM, key must have length 32 or more
func DecryptHexString(ciphertext, key string) (string, error) {
	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", errors.New(encryption.MsgInvalidHexString, err)
	}

	bytes, err := Decrypt(data, []byte(key))
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// EncryptStringToBase64 encrypts plaintext to ciphertext (base64 form) using 256-bit AES-GCM, key must have length 32 or more
func EncryptStringToBase64(plaintext, key string) (string, error) {
	bytes, err := Encrypt([]byte(plaintext), []byte(key))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}

// DecryptBase64String decrypts a base64 form ciphertext to the plaintext using 256-bit AES-GCM, key must have length 32 or more
func DecryptBase64String(ciphertext, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", errors.New(encryption.MsgInvalidBase64String, err)
	}

	bytes, err := Decrypt(data, []byte(key))
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
