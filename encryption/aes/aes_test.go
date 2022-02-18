package aes_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/encryption"
	"github.com/wego/pkg/encryption/aes"
)

var (
	key        = "1234567890qwertyuiop0123456789as"
	plaintext  = "Wego is awesome!"
	plainBytes = []byte(plaintext)
)

func Test_EncryptBase64_DecryptBase64_Ok(t *testing.T) {
	assert := assert.New(t)

	ciphertext, err := aes.EncryptStringToBase64(plaintext, key)
	assert.NoError(err)
	assert.NotEmpty(ciphertext)

	bytes, err := base64.StdEncoding.DecodeString(ciphertext)
	assert.NoError(err)
	assert.NotZero(len(bytes))

	decrypted, err := aes.DecryptBase64String(ciphertext, key)
	assert.NoError(err)
	assert.Equal(plaintext, decrypted)

	ciphertext, err = aes.EncryptToBase64(plainBytes, key)
	assert.NoError(err)
	assert.NotEmpty(ciphertext)

	bytes, err = base64.StdEncoding.DecodeString(ciphertext)
	assert.NoError(err)
	assert.NotZero(len(bytes))

	decryptedData, err := aes.DecryptBase64(ciphertext, key)
	assert.NoError(err)
	assert.Equal(plainBytes, decryptedData)
}

func Test_EncryptHexString_DecryptHexString_Ok(t *testing.T) {
	assert := assert.New(t)

	ciphertext, err := aes.EncryptStringToHex(plaintext, key)
	assert.NoError(err)
	assert.NotEmpty(ciphertext)

	bytes, err := hex.DecodeString(ciphertext)
	assert.NoError(err)
	assert.NotZero(len(bytes))

	decrypted, err := aes.DecryptHexString(ciphertext, key)
	assert.NoError(err)
	assert.Equal(plaintext, decrypted)

	ciphertext, err = aes.EncryptToHex(plainBytes, key)
	assert.NoError(err)
	assert.NotEmpty(ciphertext)

	bytes, err = hex.DecodeString(ciphertext)
	assert.NoError(err)
	assert.NotZero(len(bytes))

	decryptedData, err := aes.DecryptHex(ciphertext, key)
	assert.NoError(err)
	assert.Equal(plainBytes, decryptedData)
}

func Test_Encrypt_KeyIsTooShort(t *testing.T) {
	assert := assert.New(t)

	ciphertext, err := aes.EncryptStringToHex(plaintext, key[1:])
	assert.Error(err)
	assert.Equal(aes.ErrShortKey, err)
	assert.Empty(ciphertext)
}

func Test_DecryptBase64String_InvalidString(t *testing.T) {
	assert := assert.New(t)

	ciphertext := "Wego is awesome!"

	plaintext, err := aes.DecryptBase64String(ciphertext, key)
	assert.Error(err)
	assert.Contains(err.Error(), encryption.MsgInvalidBase64String)
	assert.Empty(plaintext)
}

func Test_DecryptBase64String_KeyIsTooShort(t *testing.T) {
	assert := assert.New(t)

	ciphertext := base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 90})

	plaintext, err := aes.DecryptBase64String(ciphertext, key[1:])
	assert.Error(err)
	assert.Equal(aes.ErrShortKey, err)
	assert.Empty(plaintext)
}

func Test_DecryptBase64String_MalformedCiphertext_InvalidNonceSize(t *testing.T) {
	assert := assert.New(t)

	ciphertext := base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 90})

	plaintext, err := aes.DecryptBase64String(ciphertext, key)
	assert.Error(err)
	assert.Equal(encryption.MsgCiphertextTooShort, err.Error())
	assert.Empty(plaintext)
}

func Test_DecryptHexString_InvalidString(t *testing.T) {
	assert := assert.New(t)

	ciphertext := "Wego is awesome!"

	plaintext, err := aes.DecryptHexString(ciphertext, key)
	assert.Error(err)
	assert.Contains(err.Error(), encryption.MsgInvalidHexString)
	assert.Empty(plaintext)
}

func Test_DecryptHexString_KeyIsTooShort(t *testing.T) {
	assert := assert.New(t)

	ciphertext := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 90})

	plaintext, err := aes.DecryptHexString(ciphertext, key[1:])
	assert.Error(err)
	assert.Equal(aes.ErrShortKey, err)
	assert.Empty(plaintext)
}

func Test_DecryptHexString_MalformedCiphertext_InvalidNonceSize(t *testing.T) {
	assert := assert.New(t)

	ciphertext := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 90})

	plaintext, err := aes.DecryptHexString(ciphertext, key)
	assert.Error(err)
	assert.Equal(encryption.MsgCiphertextTooShort, err.Error())
	assert.Empty(plaintext)
}
