package ecies_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/encryption/ecies"
	"testing"
)

const (
	privateKey = `
-----BEGIN EC PRIVATE KEY-----
MIHcAgEBBEIBJ05YGc5gmMUq1l2atl3HqQkyrGQbOH9v8nK/hhwM4VpbmRPUI/xF
zx3F5zQW0DZa+TTvXwzd5bJtUjQ75nzYcUmgBwYFK4EEACOhgYkDgYYABAE6zyKf
6PUGDIOHnW+TwzYxGBUq1TCjkxHr8Mda+5FLxONdCD/Gc4S9wbxwQecKSChaIgPQ
E1QHr/VAE5vBfX7nagDxGqL/0HMPUIxHbG4fYyVw6O5mCuA8JiMfcZOXvsZhqTNb
vMLQIXPinzVBR4u4lFQWRLttAvYC9JGO7Ar4smKXQQ==
-----END EC PRIVATE KEY-----
`
	publicKey = `
-----BEGIN PUBLIC KEY-----
MIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQBOs8in+j1BgyDh51vk8M2MRgVKtUw
o5MR6/DHWvuRS8TjXQg/xnOEvcG8cEHnCkgoWiID0BNUB6/1QBObwX1+52oA8Rqi
/9BzD1CMR2xuH2MlcOjuZgrgPCYjH3GTl77GYakzW7zC0CFz4p81QUeLuJRUFkS7
bQL2AvSRjuwK+LJil0E=
-----END PUBLIC KEY-----
`
)

// ecdh ...
func ecdh(priv *ecies.PrivateKey, pub *ecies.PublicKey) (secret []byte) {
	p := pub.ScalarMult(priv)

	// ECDH on browser only support up to 528 bits (66 bytes)
	// sometimes the x length is not enough, we need to left pad with 0
	const length = 66
	secret = p.X.Bytes()
	if len(secret) < length {
		return append(make([]byte, length-len(secret)), secret...)
	}
	return secret
}

func Test_Load_Unload_Keys_Ok(t *testing.T) {
	assert := assert.New(t)
	priv, err := ecies.PrivateKeyFromPEMString(privateKey)
	assert.NoError(err)
	assert.NotNil(priv)

	pub, err := ecies.PublicKeyFromPEMString(publicKey)
	assert.NoError(err)
	assert.NotNil(pub)
	assert.Equal(priv.Pub.Hex(), pub.Hex())

	toEncrypt := []byte("hello world")
	encrypted, err := ecies.Encrypt(toEncrypt, pub, ecdh, nil)
	assert.NoError(err)
	assert.NotNil(encrypted)

	decrypted, err := ecies.Decrypt(encrypted, priv, ecdh, nil)
	assert.NoError(err)
	assert.NotNil(decrypted)
	assert.Equal(toEncrypt, decrypted)
}
