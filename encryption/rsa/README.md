# RSA Encryption Package

This package provides RSA encryption and decryption functionality using PKCS1v15 padding, compatible with Java's RSA/ECB/PKCS1Padding.

## Features

- RSA encryption/decryption with PEM formatted keys
- Support for both PKCS1 and PKCS8 key formats
- Hex and Base64 encoding/decoding of encrypted data
- Comprehensive error handling
- Well-tested implementation

## Installation

```go
import "github.com/wego/pkg/encryption/rsa"
```

## Usage

### Basic Encryption and Decryption

```go
// RSA public key in PEM format
publicKeyPEM := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
-----END PUBLIC KEY-----`

// RSA private key in PEM format
privateKeyPEM := `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA...
-----END RSA PRIVATE KEY-----`

// Encrypt data
plaintext := []byte("Hello, World!")
encrypted, err := rsa.Encrypt(plaintext, publicKeyPEM)
if err != nil {
    log.Fatal(err)
}

// Decrypt data
decrypted, err := rsa.Decrypt(encrypted, privateKeyPEM)
if err != nil {
    log.Fatal(err)
}

fmt.Println(string(decrypted)) // Output: Hello, World!
```

### Using Hex Encoding

```go
// Encrypt to hex
hexCiphertext, err := rsa.EncryptStringToHex("Secret message", publicKeyPEM)
if err != nil {
    log.Fatal(err)
}

// Decrypt from hex
plaintext, err := rsa.DecryptHexString(hexCiphertext, privateKeyPEM)
if err != nil {
    log.Fatal(err)
}
```

### Using Base64 Encoding

```go
// Encrypt to base64
base64Ciphertext, err := rsa.EncryptStringToBase64("Secret message", publicKeyPEM)
if err != nil {
    log.Fatal(err)
}

// Decrypt from base64
plaintext, err := rsa.DecryptBase64String(base64Ciphertext, privateKeyPEM)
if err != nil {
    log.Fatal(err)
}
```

## Key Format Requirements

### Public Key
- Must be in PEM format
- Supports both PKCS1 (`RSA PUBLIC KEY`) and PKCS8 (`PUBLIC KEY`) formats

### Private Key
- Must be in PEM format
- Supports both PKCS1 (`RSA PRIVATE KEY`) and PKCS8 (`PRIVATE KEY`) formats

## Message Size Limitations

RSA encryption has inherent size limitations based on the key size and padding scheme:

- For a 2048-bit key with PKCS1v15 padding: maximum message size is 245 bytes (256 - 11)
- For a 4096-bit key with PKCS1v15 padding: maximum message size is 501 bytes (512 - 11)

For larger messages, consider using hybrid encryption (RSA + AES).

## Error Handling

The package defines the following errors:

- `ErrInvalidPEM`: The provided string is not a valid PEM format
- `ErrNotRSAKey`: The provided key is not an RSA key
- `ErrNotPKCS1v15`: RSA operation failed

## Performance Considerations

RSA operations are computationally expensive. For bulk data encryption, it's recommended to:

1. Generate a random AES key
2. Encrypt the data with AES
3. Encrypt the AES key with RSA
4. Send both the encrypted data and encrypted key

## Compatibility

This implementation uses PKCS1v15 padding, which is compatible with:
- Java's `RSA/ECB/PKCS1Padding`
- OpenSSL's default RSA padding
- Most standard RSA implementations

## Testing

Run the tests with:

```bash
go test ./rsa -v
```

Run benchmarks with:

```bash
go test ./rsa -bench=. -benchmem
```

## Example

See `example_test.go` for complete working examples. 