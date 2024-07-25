package xor_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/encryption/xor"
	"testing"
)

func Test_EncryptDecrypt_OK(t *testing.T) {
	ass := assert.New(t)
	data := []byte("some data")
	key := []byte("some key")
	encrypted := xor.Encrypt(data, key)
	ass.Equal(string(data), string(xor.Decrypt(encrypted, key)))
}
