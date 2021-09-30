package aes_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/encryption/aes"
)

func Test_Encrypt_Decrypt_Ok(t *testing.T) {
	assert := assert.New(t)

	key := "1234567890qwertyuiop0123456789as"
	plaintext := "Wego is awesome!"

	ciphertext, err := aes.Encrypt(plaintext, key)
	assert.NoError(err)

	decrypted, err := aes.Decrypt(ciphertext, key)
	assert.NoError(err)

	assert.Equal(plaintext, decrypted)
}

func Test_Encrypt_KeyIsTooShort(t *testing.T) {
	assert := assert.New(t)

	key := "1234567890qwertyuiop0123456789a"
	plaintext := "Wego is awesome!"

	ciphertext, err := aes.Encrypt(plaintext, key)
	assert.Error(err)
	assert.Empty(ciphertext)
}

func Test_Decrypt_MalformedCiphertext_InvalidHexCode(t *testing.T) {
	assert := assert.New(t)

	key := "1234567890qwertyuiop0123456789as"
	ciphertext := "Wego is awesome!"

	plaintext, err := aes.Decrypt(ciphertext, key)
	assert.Error(err)
	assert.Empty(plaintext)
}

func Test_Decrypt_KeyIsTooShort(t *testing.T) {
	assert := assert.New(t)

	key := "1234567890qwertyuiop0123456789a"
	ciphertext := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 90})

	plaintext, err := aes.Decrypt(ciphertext, key)
	assert.Error(err)
	assert.Empty(plaintext)
}

func Test_Decrypt_MalformedCiphertext_InvalidNonceSize(t *testing.T) {
	assert := assert.New(t)

	key := "1234567890qwertyuiop0123456789as"
	ciphertext := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 90})

	plaintext, err := aes.Decrypt(ciphertext, key)
	assert.Error(err)
	assert.Empty(plaintext)
}
