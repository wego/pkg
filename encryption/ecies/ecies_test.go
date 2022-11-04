package ecies_test

import (
	"crypto/elliptic"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/encryption"
	"github.com/wego/pkg/encryption/ecies"
)

var (
	curve      = elliptic.P521()
	priv, _    = ecies.GenerateKey(curve)
	plainText  = "Wego is awesome!"
	plainBytes = []byte(plainText)
)

func Test_EncryptBase64String_DecryptBase64String_Ok(t *testing.T) {
	assert := assert.New(t)

	ciphertext, err := ecies.EncryptStringToBase64(plainText, priv.Pub, nil, nil)
	assert.NoError(err)
	assert.NotEmpty(ciphertext)

	bytes, err := base64.StdEncoding.DecodeString(ciphertext)
	assert.NoError(err)
	assert.NotEmpty(bytes)

	decoded, err := ecies.DecryptBase64String(ciphertext, priv, nil, nil)
	assert.NoError(err)
	assert.Equal(plainText, decoded)

	ciphertext, err = ecies.EncryptToBase64(plainBytes, priv.Pub, nil, nil)
	assert.NoError(err)
	assert.NotEmpty(ciphertext)

	bytes, err = base64.StdEncoding.DecodeString(ciphertext)
	assert.NoError(err)
	assert.NotEmpty(bytes)

	decryptedData, err := ecies.DecryptBase64(ciphertext, priv, nil, nil)
	assert.NoError(err)
	assert.Equal(plainBytes, decryptedData)

}

func Test_EncryptHexString_DecryptHexString_Ok(t *testing.T) {
	assert := assert.New(t)

	ciphertext, err := ecies.EncryptStringToHex(plainText, priv.Pub, nil, nil)
	assert.NoError(err)
	assert.NotEmpty(ciphertext)

	bytes, err := hex.DecodeString(ciphertext)
	assert.NoError(err)
	assert.NotEmpty(bytes)

	decoded, err := ecies.DecryptHexString(ciphertext, priv, nil, nil)
	assert.NoError(err)
	assert.Equal(plainText, decoded)

	ciphertext, err = ecies.EncryptToHex(plainBytes, priv.Pub, nil, nil)
	assert.NoError(err)
	assert.NotEmpty(ciphertext)

	bytes, err = hex.DecodeString(ciphertext)
	assert.NoError(err)
	assert.NotEmpty(bytes)

	decryptedData, err := ecies.DecryptHex(ciphertext, priv, nil, nil)
	assert.NoError(err)
	assert.Equal(plainBytes, decryptedData)
}

func Test_DecryptBase64String_Ciphertext_TooShort(t *testing.T) {
	assert := assert.New(t)

	plaintext, err := ecies.DecryptBase64String("ZGNjZGFjdmRm", priv, nil, nil)
	assert.Error(err)
	assert.Equal(encryption.MsgCiphertextTooShort, err.Error())
	assert.Empty(plaintext)
}

func Test_DecryptBase64String_InvalidBase64String(t *testing.T) {
	assert := assert.New(t)

	ciphertext := `Lorem ipsum dolor sit amet`
	plaintext, err := ecies.DecryptBase64String(ciphertext, priv, nil, nil)
	assert.Error(err)
	assert.Contains(err.Error(), encryption.MsgInvalidBase64String)
	assert.Empty(plaintext)
}

func Test_DecryptBase64String_PublicKeyMismatch(t *testing.T) {
	assert := assert.New(t)

	ciphertext := `BAAxCirH5mdrsQk/viVLKABpJOUTFqIqvzpklhb5VME41lR1G3QwTx+X+VzsvQsWLUkUKQCkxOOU6+M3GBwOhwINS` +
		`QH8eOQbXOb0uAXcbmPCHaJ1kq5QH7pkHMMMWtUAqR7Ls/7PCy1KJ3vb+5P15MVm9WN3uB7T+NAdVNIbKyQJNUnLuCbc236I9LkJ` +
		`/SiZ3AkwQhJkSbc7w2mHP1sdN2gE7UyBsonL/+G5Hw==`
	plaintext, err := ecies.DecryptBase64String(ciphertext, priv, nil, nil)
	assert.Error(err)
	assert.Contains(err.Error(), "message authentication failed")
	assert.Empty(plaintext)
}

func Test_DecryptHexString_InvalidHexString(t *testing.T) {
	assert := assert.New(t)

	ciphertext := `Lorem ipsum dolor sit amet`
	plaintext, err := ecies.DecryptHexString(ciphertext, priv, nil, nil)
	assert.Error(err)
	assert.Contains(err.Error(), encryption.MsgInvalidHexString)
	assert.Empty(plaintext)
}

func Test_DecryptHexString_Ciphertext_TooShort(t *testing.T) {
	assert := assert.New(t)

	plaintext, err := ecies.DecryptHexString("013040f2b8", priv, nil, nil)
	assert.Error(err)
	assert.Equal(encryption.MsgCiphertextTooShort, err.Error())
	assert.Empty(plaintext)
}

func Test_DecryptHexString_PublicKeyMismatch(t *testing.T) {
	assert := assert.New(t)

	ciphertext := `04013f40882961559ee7d283221eff5f7639aa2c4ffe69215ddf6491fa3acf69897e97100a26713da7e5b1d61` +
		`e59825c1b80723f428c5d517efdcc00e64408b6f6d38e0078e34747b78c588d672b6b7368622be61406e2478992ebd62c21` +
		`63f16d0eccd0915829b05e702feb0a6933cb9b6a81f06b33fb8648b1e7269d91d170d93869b8f983f1f7be5536258d39dd8` +
		`ca64b9ced2bd316be805702d8728206f168ab7f8ecb14fcd29ee0c5e5`
	plaintext, err := ecies.DecryptHexString(ciphertext, priv, nil, nil)
	assert.Error(err)
	assert.Contains(err.Error(), "message authentication failed")
	assert.Empty(plaintext)
}

func Test_PublicKey_ScalarMult(t *testing.T) {
	assert := assert.New(t)

	newPriv, err := ecies.GenerateKey(curve)
	assert.NoError(err)

	// if we implement correctly then
	// (publicKeyA * privateKeyB) = (publicKeyB * privateKeyA)
	point1 := newPriv.Pub.ScalarMult(priv)
	point2 := priv.Pub.ScalarMult(newPriv)

	assert.Equal(point1.X, point2.X)
	assert.Equal(point1.Y, point2.Y)
}
