package ecies_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/encryption/ecies"
)

func Test_Load_Unload_PrivateKey_Ok(t *testing.T) {
	assert := assert.New(t)

	bytes := priv.Bytes()
	loadedPriv := ecies.PrivateKeyFromBytes(bytes, curve)

	assert.Equal(bytes, loadedPriv.Bytes())

	hex := priv.Hex()
	loadedPriv, err := ecies.PrivateKeyFromHex(hex, curve)
	assert.NoError(err)
	assert.Equal(hex, loadedPriv.Hex())
}

func Test_PrivateKeyFromHex_Error(t *testing.T) {
	assert := assert.New(t)

	loadedPriv, err := ecies.PrivateKeyFromHex("abc", curve)
	assert.Error(err)
	assert.Contains(err.Error(), "error decoding hexKey")
	assert.Nil(loadedPriv)
}
