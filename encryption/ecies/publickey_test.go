package ecies_test

import (
	"crypto/elliptic"
	"encoding/base64"
	"encoding/hex"
	"math/big"
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

	pem := priv.Pub.PEM()
	pub, err = ecies.PublicKeyFromPEMString(pem)
	assert.NoError(err)
	assert.Equal(pem, pub.PEM())
	assert.Equal(priv.Pub.Bytes(), pub.Bytes())
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

// invalidP521Bytes builds a well-formed uncompressed public key (0x04 || X || Y)
// whose coordinates (1, 1) are NOT on the P521 curve.
// This simulates corrupted or malicious encrypted payloads where the embedded
// ephemeral public key has valid length but invalid curve coordinates.
// Go 1.24+ panics in crypto/elliptic.ScalarMult for such points (PAY-2108).
func invalidP521Bytes() []byte {
	size := (elliptic.P521().Params().BitSize + 7) / 8 // 66 bytes
	b := make([]byte, 1+2*size)
	b[0] = 0x04 // uncompressed point prefix
	big.NewInt(1).FillBytes(b[1 : size+1])
	big.NewInt(1).FillBytes(b[size+1:])
	return b
}

// Test_PublicKeyFromBytes_InvalidCurvePoint verifies that off-curve points are
// rejected at parse time rather than causing a panic in downstream ScalarMult.
func Test_PublicKeyFromBytes_InvalidCurvePoint(t *testing.T) {
	assert := assert.New(t)

	pub, err := ecies.PublicKeyFromBytes(invalidP521Bytes(), elliptic.P521())
	assert.Error(err)
	assert.Contains(err.Error(), "point is not on the curve")
	assert.Nil(pub)
}

// Test_PublicKeyFromBase64_InvalidCurvePoint ensures the validation propagates
// through the base64 parsing path (PublicKeyFromBase64 -> PublicKeyFromBytes).
func Test_PublicKeyFromBase64_InvalidCurvePoint(t *testing.T) {
	assert := assert.New(t)

	b64 := base64.StdEncoding.EncodeToString(invalidP521Bytes())
	pub, err := ecies.PublicKeyFromBase64(b64, elliptic.P521())
	assert.Error(err)
	assert.Contains(err.Error(), "point is not on the curve")
	assert.Nil(pub)
}

// Test_PublicKeyFromHex_InvalidCurvePoint ensures the validation propagates
// through the hex parsing path (PublicKeyFromHex -> PublicKeyFromBytes).
func Test_PublicKeyFromHex_InvalidCurvePoint(t *testing.T) {
	assert := assert.New(t)

	h := hex.EncodeToString(invalidP521Bytes())
	pub, err := ecies.PublicKeyFromHex(h, elliptic.P521())
	assert.Error(err)
	assert.Contains(err.Error(), "point is not on the curve")
	assert.Nil(pub)
}
