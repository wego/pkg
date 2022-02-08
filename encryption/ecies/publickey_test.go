package ecies_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/encryption/ecies"
)

func Test_Load_Unload_PublicKey_Ok(t *testing.T) {
	assert := assert.New(t)

	bytes := priv.Pub.Bytes()
	pub, err := ecies.PublicKeyFromBytes(bytes, curve)
	assert.NoError(err)
	assert.Equal(bytes, pub.Bytes())

	b64 := priv.Pub.Base64()
	pub, err = ecies.PublicKeyFromBase64(b64, curve)
	assert.NoError(err)
	assert.Equal(b64, pub.Base64())

	hex := priv.Pub.Hex()
	pub, err = ecies.PublicKeyFromHex(hex, curve)
	assert.NoError(err)
	assert.Equal(hex, pub.Hex())
}

func Test_PublicKeyFromBase64_Error(t *testing.T) {
	assert := assert.New(t)

	pub, err := ecies.PublicKeyFromBase64("abc", curve)
	assert.Error(err)
	assert.Contains(err.Error(), "error decoding base64Key")
	assert.Nil(pub)

	b64 := `dGFrZWl0ZWFzeQ==`
	pub, err = ecies.PublicKeyFromBase64(b64, curve)
	assert.Error(err)
	assert.Contains(err.Error(), "invalid key length")
	assert.Nil(pub)
}

func Test_PublicKeyFromHex_Error(t *testing.T) {
	assert := assert.New(t)

	pub, err := ecies.PublicKeyFromHex("abc", curve)
	assert.Error(err)
	assert.Contains(err.Error(), "error decoding hexKey")
	assert.Nil(pub)

	hex := `013040f2b8bb9e32fec39eb3d79e7ffa06ebae89790bd099fb7004b85ee92f09e0a564081619478d15a3fbad`
	pub, err = ecies.PublicKeyFromHex(hex, curve)
	assert.Error(err)
	assert.Contains(err.Error(), "invalid key length")
	assert.Nil(pub)
}
