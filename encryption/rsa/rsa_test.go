package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

// Helper function to generate RSA key pair for testing
func generateTestKeyPair(t testing.TB) (publicKeyPEM, privateKeyPEM string) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Encode private key to PEM format
	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDER,
	}
	privateKeyPEM = string(pem.EncodeToMemory(privateKeyBlock))

	// Encode public key to PEM format
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	}
	publicKeyPEM = string(pem.EncodeToMemory(publicKeyBlock))

	return publicKeyPEM, privateKeyPEM
}

func TestEncryptDecrypt(t *testing.T) {
	publicKeyPEM, privateKeyPEM := generateTestKeyPair(t)

	testCases := []struct {
		name string
		data []byte
	}{
		{"Short message", []byte("Hello, World!")},
		{"Empty message", []byte("")},
		{"Binary data", []byte{0x00, 0x01, 0x02, 0x03, 0x04}},
		{"Long message", []byte("This is a longer message that tests the RSA encryption functionality")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := Encrypt(tc.data, publicKeyPEM)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Decrypt
			decrypted, err := Decrypt(encrypted, privateKeyPEM)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Verify
			if string(decrypted) != string(tc.data) {
				t.Errorf("Decrypted data doesn't match original. Got: %s, Want: %s", decrypted, tc.data)
			}
		})
	}
}

func TestEncryptToHexAndDecryptHex(t *testing.T) {
	publicKeyPEM, privateKeyPEM := generateTestKeyPair(t)
	plaintext := "Test message for hex encoding"

	// Encrypt to hex
	hexCiphertext, err := EncryptStringToHex(plaintext, publicKeyPEM)
	if err != nil {
		t.Fatalf("EncryptStringToHex failed: %v", err)
	}

	// Decrypt from hex
	decrypted, err := DecryptHexString(hexCiphertext, privateKeyPEM)
	if err != nil {
		t.Fatalf("DecryptHexString failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted text doesn't match. Got: %s, Want: %s", decrypted, plaintext)
	}
}

func TestEncryptToBase64AndDecryptBase64(t *testing.T) {
	publicKeyPEM, privateKeyPEM := generateTestKeyPair(t)
	plaintext := "Test message for base64 encoding"

	// Encrypt to base64
	base64Ciphertext, err := EncryptStringToBase64(plaintext, publicKeyPEM)
	if err != nil {
		t.Fatalf("EncryptStringToBase64 failed: %v", err)
	}

	// Decrypt from base64
	decrypted, err := DecryptBase64String(base64Ciphertext, privateKeyPEM)
	if err != nil {
		t.Fatalf("DecryptBase64String failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted text doesn't match. Got: %s, Want: %s", decrypted, plaintext)
	}
}

func TestEncryptWithInvalidKey(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{"Empty key", ""},
		{"Invalid PEM", "not a valid PEM string"},
		{"Wrong key type", `-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHHIG...
-----END CERTIFICATE-----`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Encrypt([]byte("test"), tc.key)
			if err == nil {
				t.Error("Expected error for invalid key, but got none")
			}
		})
	}
}

func TestDecryptWithInvalidKey(t *testing.T) {
	publicKeyPEM, _ := generateTestKeyPair(t)

	// First encrypt some data
	encrypted, err := Encrypt([]byte("test"), publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	testCases := []struct {
		name string
		key  string
	}{
		{"Empty key", ""},
		{"Invalid PEM", "not a valid PEM string"},
		{"Wrong key type", publicKeyPEM}, // Using public key for decryption
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decrypt(encrypted, tc.key)
			if err == nil {
				t.Error("Expected error for invalid key, but got none")
			}
		})
	}
}

func TestDecryptWithWrongPrivateKey(t *testing.T) {
	publicKeyPEM1, _ := generateTestKeyPair(t)
	_, privateKeyPEM2 := generateTestKeyPair(t)

	// Encrypt with first key pair
	encrypted, err := Encrypt([]byte("test"), publicKeyPEM1)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Try to decrypt with different private key
	_, err = Decrypt(encrypted, privateKeyPEM2)
	if err == nil {
		t.Error("Expected error when decrypting with wrong private key")
	}
}

func TestMaxMessageSize(t *testing.T) {
	publicKeyPEM, _ := generateTestKeyPair(t)

	// RSA with 2048-bit key and PKCS1v15 padding can encrypt max (keySize/8 - 11) bytes
	// For 2048-bit key: 256 - 11 = 245 bytes max
	maxSize := 245
	largeMessage := make([]byte, maxSize+1)
	for i := range largeMessage {
		largeMessage[i] = byte(i % 256)
	}

	// This should fail due to message being too large
	_, err := Encrypt(largeMessage, publicKeyPEM)
	if err == nil {
		t.Error("Expected error for message too large for RSA key")
	}

	// This should succeed
	smallerMessage := largeMessage[:maxSize]
	_, err = Encrypt(smallerMessage, publicKeyPEM)
	if err != nil {
		t.Errorf("Failed to encrypt message within size limit: %v", err)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	publicKeyPEM, _ := generateTestKeyPair(b)
	data := []byte("Benchmark test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Encrypt(data, publicKeyPEM)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	publicKeyPEM, privateKeyPEM := generateTestKeyPair(b)
	data := []byte("Benchmark test message")

	encrypted, err := Encrypt(data, publicKeyPEM)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decrypt(encrypted, privateKeyPEM)
		if err != nil {
			b.Fatal(err)
		}
	}
}
