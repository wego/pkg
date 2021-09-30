package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// KeyLength is the min length of the secret key
const KeyLength = 32

// Encrypt encrypts data using 256-bit AES-GCM, key must have length 32 or more
func Encrypt(plaintext string, key string) (ciphertext string, err error) {
	keyBytes, err := getKey(key)
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
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return
	}

	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

// Decrypt decrypts data using 256-bit AES-GCM, key must have length 32 or more
func Decrypt(ciphertext string, key string) (plaintext string, err error) {
	cipherBytes, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("ciphertext [%s] is not a valid hex string", ciphertext)
	}

	keyBytes, err := getKey(key)
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
		return "", fmt.Errorf("ciphertext [%s] is too short", ciphertext)
	}

	plainBytes, err := gcm.Open(nil,
		cipherBytes[:gcm.NonceSize()],
		cipherBytes[gcm.NonceSize():],
		nil,
	)
	if err != nil {
		return
	}

	return string(plainBytes), nil
}

func getKey(secret string) (key *[KeyLength]byte, err error) {
	if len(secret) < KeyLength {
		return nil, fmt.Errorf("secret key is too short, require [%d] or more", KeyLength)
	}

	// try to get the key if it's in hex code
	secretBytes, err := hex.DecodeString(secret)
	if err != nil {
		secretBytes = []byte(secret)
	}

	key = &[KeyLength]byte{}
	copy(key[:], secretBytes[:KeyLength])
	return key, nil
}
