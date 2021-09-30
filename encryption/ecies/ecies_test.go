package ecies_test

import (
	"crypto/elliptic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/encryption/ecies"
)

func Test_Encrypt_Decrypt_Ok(t *testing.T) {
	assert := assert.New(t)

	priv, err := ecies.GenerateKey(elliptic.P521())
	assert.NoError(err)
	msg := "Wego is awesome!"

	ciphertext, err := ecies.Encrypt(msg, priv.Pub)
	assert.NoError(err)
	assert.NotEmpty(ciphertext)

	plaintext, err := ecies.Decrypt(ciphertext, priv)
	assert.NoError(err)

	assert.Equal(msg, plaintext)
}
