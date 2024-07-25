package xor

// Encrypt encrypts the given data using the given key.
func Encrypt(data, key []byte) []byte {
	return xor(data, key)
}

// Decrypt decrypts the given data using the given key.
func Decrypt(data, key []byte) []byte {
	return xor(data, key)
}

func xor(a, b []byte) []byte {
	for i := range a {
		a[i] ^= b[i%len(b)]
	}
	return a
}
