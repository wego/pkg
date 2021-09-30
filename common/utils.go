package common

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

// Encrypt encrypts data using 256-bit AES-GCM, key must have length 32 or more
func Encrypt(plaintext string, key string) (ciphertext string, err error) {
	keyBytes, err := getAESKey(key)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(cryptoRand.Reader, nonce)
	if err != nil {
		return
	}

	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

// Decrypt decrypts data using 256-bit AES-GCM, key must have length 32 or more
func Decrypt(ciphertext string, key string) (plaintext string, err error) {
	cipherBytes, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("Malformed ciphertext: " + ciphertext)
	}

	keyBytes, err := getAESKey(key)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	if len(cipherBytes) < gcm.NonceSize() {
		return "", fmt.Errorf("malformed ciphertext")
	}

	plainBytes, err := gcm.Open(nil,
		cipherBytes[:gcm.NonceSize()],
		cipherBytes[gcm.NonceSize():],
		nil,
	)
	plaintext = string(plainBytes)
	return
}

func getAESKey(secret string) (key *[AESKeyLength]byte, err error) {
	if len(secret) < AESKeyLength {
		return nil, fmt.Errorf("secret key length is too short, require [%d] or more", AESKeyLength)
	}

	key = &[AESKeyLength]byte{}
	copy(key[:], secret[0:AESKeyLength])
	return
}

// BoolRef returns a reference to a bool value
func BoolRef(v bool) *bool {
	return &v
}

// StrRef returns a reference to a string value
func StrRef(v string) *string {
	return &v
}

// Int32Ref returns a reference to a int32 value
func Int32Ref(v int32) *int32 {
	return &v
}

// Int64Ref returns a reference to a int64 value
func Int64Ref(v int64) *int64 {
	return &v
}

// UintRef returns a reference to a uint value
func UintRef(v uint) *uint {
	return &v
}

// Uint32Ref returns a reference to a uint32 value
func Uint32Ref(v uint32) *uint32 {
	return &v
}

// TimeRef return a reference to time value
func TimeRef(v time.Time) *time.Time {
	if v.IsZero() {
		return nil
	}
	return &v
}
