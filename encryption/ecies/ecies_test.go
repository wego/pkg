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

func Test_DecryptBase64String_InvalidPublicKeyBytes(t *testing.T) {
	assert := assert.New(t)

	ciphertext := `Z2Znc3Rnc2dyZ3JzZ3ZydGdkZnZkc2RmZHNnZmdzdGdzZ3JncnNndnJ0Z2RmdmRzZGZkc2dmZ3N0Z3Nncmdyc2d2cnRn` +
		`ZGZ2ZHNkZmRzZ2Znc3Rnc2dyZ3JzZ3ZydGdkZnZkc2RmZHNnZmdzdGdzZ3JncnNndnJ0Z2RmdmRzZGZkc2dmZ3N0Z3Nncmdyc2d2cn` +
		`RnZGZ2ZHNkZmRzZ2Znc3Rnc2dyZ3JzZ3ZydGdkZnZkc2RmZHNnZmdzdGdzZ3JncnNndnJ0Z2RmdmRzZGZkc2dmZ3N0Z3Nncmdyc2d2` +
		`cnRnZGZ2ZHNkZmRzZ2Znc3Rnc2dyZ3JzZ3ZydGdkZnZkc2RmZHNnZmdzdGdzZ3JncnNndnJ0Z2RmdmRzZGZkc2dmZ3N0Z3Nncmdyc2` +
		`d2cnRnZGZ2ZHNkZmRzZ2Znc3Rnc2dyZ3JzZ3ZydGdkZnZkc2RmZHNnZmdzdGdzZ3JncnNndnJ0Z2RmdmRzZGZkc2dmZ3N0Z3Nncmdy` +
		`c2d2cnRnZGZ2ZHNkZmRzZ2Znc3Rnc2dyZ3JzZ3ZydGdkZnZkc2RmZHM=`
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

func Test_DecryptHexString_InvalidPublicKeyBytes(t *testing.T) {
	assert := assert.New(t)

	ciphertext := `013040f2b8bb9e32fec39eb3d79e7ffa06ebae89790bd099fb7004b85ee92f09e0a564081619478d15a3fbad8` +
		`cfc5b05f1c9fdd0ee9a974461214739a0b47268497a01189e1884f0e1249e3b4ee08396c47f81cf5b0d00447554cb291ebb` +
		`804ff632e682596953311b880f8337b099eca655f4cdbb1a413bd5182991fa771e62e5028c30ef369b739f9e084be78efbd2` +
		`075db61fd40118478281b6c874bcc3f450459804112b6e76a53405a260b836f79718856e3e0c7d58`
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
