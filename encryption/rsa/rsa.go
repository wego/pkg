package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"

	"github.com/wego/pkg/encryption"
	"github.com/wego/pkg/errors"
)

var (
	// ErrInvalidPEM indicates the provided string is not a valid PEM format
	ErrInvalidPEM = errors.New("invalid PEM format")
	// ErrNotRSAKey indicates the provided key is not an RSA key
	ErrNotRSAKey = errors.New("provided key is not an RSA key")
	// ErrNotPKCS1v15 indicates encryption/decryption failed
	ErrNotPKCS1v15 = errors.New("RSA operation failed")
)

// EncryptToHex encrypts data to encrypted (hex form) using RSA PKCS1v15 with the provided public key in PEM format
func EncryptToHex(data []byte, publicKeyPEM string) (string, error) {
	bytes, err := Encrypt(data, publicKeyPEM)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// EncryptStringToHex encrypts plaintext to encrypted (hex form) using RSA PKCS1v15 with the provided public key in PEM format
func EncryptStringToHex(plaintext, publicKeyPEM string) (string, error) {
	return EncryptToHex([]byte(plaintext), publicKeyPEM)
}

// DecryptHex decrypts a hex form encrypted to raw data([]byte) using RSA PKCS1v15 with the provided private key in PEM format
func DecryptHex(encrypted, privateKeyPEM string) ([]byte, error) {
	data, err := hex.DecodeString(encrypted)
	if err != nil {
		return nil, errors.New(encryption.MsgInvalidHexString, err)
	}

	return Decrypt(data, privateKeyPEM)
}

// DecryptHexString decrypts a hex form encrypted to the plaintext using RSA PKCS1v15 with the provided private key in PEM format
func DecryptHexString(encrypted, privateKeyPEM string) (string, error) {
	bytes, err := DecryptHex(encrypted, privateKeyPEM)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// EncryptToBase64 encrypts data to encrypted (base64 form) using RSA PKCS1v15 with the provided public key in PEM format
func EncryptToBase64(data []byte, publicKeyPEM string) (string, error) {
	bytes, err := Encrypt(data, publicKeyPEM)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}

// EncryptStringToBase64 encrypts plaintext to encrypted (base64 form) using RSA PKCS1v15 with the provided public key in PEM format
func EncryptStringToBase64(plaintext, publicKeyPEM string) (string, error) {
	return EncryptToBase64([]byte(plaintext), publicKeyPEM)
}

// DecryptBase64 decrypts a base64 form encrypted to raw data([]byte) using RSA PKCS1v15 with the provided private key in PEM format
func DecryptBase64(encrypted, privateKeyPEM string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, errors.New(encryption.MsgInvalidBase64String, err)
	}

	return Decrypt(data, privateKeyPEM)
}

// DecryptBase64String decrypts a base64 form encrypted to the plaintext using RSA PKCS1v15 with the provided private key in PEM format
func DecryptBase64String(encrypted, privateKeyPEM string) (string, error) {
	bytes, err := DecryptBase64(encrypted, privateKeyPEM)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Encrypt encrypts data using RSA PKCS1v15 with the provided public key in PEM format
func Encrypt(data []byte, publicKeyPEM string) (encrypted []byte, err error) {
	// Parse PEM block
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, ErrInvalidPEM
	}

	// Parse the public key
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// Try parsing as PKCS1 format
		publicKeyInterface, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, ErrNotPKCS1v15
		}
	}

	// Type assert to rsa.PublicKey
	rsaPublicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, ErrNotRSAKey
	}

	// Encrypt the data using the RSA public key
	return encrypt(data, rsaPublicKey)
}

// encrypt encrypts data using RSA public key with PKCS1v15 padding
func encrypt(data []byte, publicKey *rsa.PublicKey) (encrypted []byte, err error) {
	// Encrypt the content using PKCS1v15 (equivalent to Java's RSA/ECB/PKCS1Padding)
	encrypted, err = rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		return nil, ErrNotPKCS1v15
	}
	return
}

// Decrypt decrypts data using RSA PKCS1v15 with the provided private key in PEM format
func Decrypt(encrypted []byte, privateKeyPEM string) (data []byte, err error) {
	// Parse PEM block
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, ErrInvalidPEM
	}

	// Parse the private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try parsing as PKCS8 format
		privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, ErrNotPKCS1v15
		}

		var ok bool
		privateKey, ok = privateKeyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, ErrNotRSAKey
		}
	}

	return decrypt(encrypted, privateKey)
}

// decrypt decrypts data using RSA private key with PKCS1v15 padding
func decrypt(encrypted []byte, privateKey *rsa.PrivateKey) (data []byte, err error) {
	// Decrypt the content using PKCS1v15 (equivalent to Java's RSA/ECB/PKCS1Padding)
	data, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, encrypted)
	if err != nil {
		return nil, ErrNotPKCS1v15
	}
	return
}
