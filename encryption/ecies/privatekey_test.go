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

	b64 := priv.Base64()
	loadedPriv, err := ecies.PrivateKeyFromBase64(b64, curve)
	assert.NoError(err)
	assert.Equal(b64, loadedPriv.Base64())

	hex := priv.Hex()
	loadedPriv, err = ecies.PrivateKeyFromHex(hex, curve)
	assert.NoError(err)
	assert.Equal(hex, loadedPriv.Hex())

	pem := priv.PEM()
	loadedPriv, err = ecies.PrivateKeyFromPEMString(pem)

	assert.NoError(err)
	assert.Equal(pem, loadedPriv.PEM())
	assert.Equal(priv.Bytes(), loadedPriv.Bytes())
}

func Test_PrivateKeyFromBase64_Error(t *testing.T) {
	assert := assert.New(t)

	loadedPriv, err := ecies.PrivateKeyFromBase64("abc", curve)
	assert.Error(err)
	assert.Contains(err.Error(), "error decoding base64Key")
	assert.Nil(loadedPriv)
}

func Test_PrivateKeyFromHex_Error(t *testing.T) {
	assert := assert.New(t)

	loadedPriv, err := ecies.PrivateKeyFromHex("abc", curve)
	assert.Error(err)
	assert.Contains(err.Error(), "error decoding hexKey")
	assert.Nil(loadedPriv)
}
