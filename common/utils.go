package common

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"time"
)

// Int2String converts a number to a string using characters from chars.
//
// Validation:
//  - `number` must >= 0
//  - `length` must > 0
//  - length of `chars` must >= 10
//  - `paddingChars` must not be empty
//  - `chars` & `paddingChars` must not overlap
// If length of the result < required length, the result will be left padded with random characters from `paddingChars`.
func Int2String(number int64, chars, paddingChars string, length int) (s string, err error) {
	if number < 0 {
		return s, fmt.Errorf("invalid input: number < 0")
	}
	if length < 1 {
		return s, fmt.Errorf("invalid input: length <= 0")
	}
	err = verifyInputChars(chars, paddingChars)
	if err != nil {
		return
	}

	base := int64(len(chars))
	for number > 0 {
		i := number % base
		s = string(chars[i]) + s
		number = number / base
	}

	resultLength := len(s)
	if resultLength > length {
		return "", fmt.Errorf("overlength: result %s is longer than required length %d", s, length)
	}

	// left padding if not meet the required length
	for ; resultLength < length; resultLength++ {
		s = string(paddingChars[rand.Intn(len(paddingChars))]) + s
	}

	return s, nil
}

// String2Int is the revert of Int2String
//  It returns negative result in 2 cases:
//  - s contains character not exist in chars
//  - the result is bigger than max int64 (integer overflow happen)
func String2Int(s string, chars, paddingChars string) (n int64, err error) {
	err = verifyInputChars(chars, paddingChars)
	if err != nil {
		return
	}

	// remove padding from the left side of string
	for _, c := range s {
		if !strings.Contains(paddingChars, string(c)) {
			break
		}
		s = s[1:]
	}

	// example for chars=0123456789ABCDEF, s=3A0DE (after remove padding)
	// base=16
	// n = 3*16^4 + 10*16^3 + 0*16^2 + 13*16^1 + 14*16^0 = 237790
	base := len(chars)
	for i := range s {
		indexValue := int64(strings.IndexByte(chars, s[i]))
		exponent := len(s) - 1 - i
		n += indexValue * int64(math.Pow(float64(base), float64(exponent)))
	}
	return
}

// Encrypt encrypts data using 256-bit AES-GCM, key must have length 32 or more
func Encrypt(plaintext string, key string) (ciphertext string, err error) {
	keyBytes, err := getAESKey(key)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(cryptoRand.Reader, nonce)
	if err != nil {
		return
	}

	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

// Decrypt decrypts data using 256-bit AES-GCM, key must have length 32 or more
func Decrypt(ciphertext string, key string) (plaintext string, err error) {
	cipherBytes, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("Malformed ciphertext: " + ciphertext)
	}

	keyBytes, err := getAESKey(key)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	if len(cipherBytes) < gcm.NonceSize() {
		return "", fmt.Errorf("malformed ciphertext")
	}

	plainBytes, err := gcm.Open(nil,
		cipherBytes[:gcm.NonceSize()],
		cipherBytes[gcm.NonceSize():],
		nil,
	)
	plaintext = string(plainBytes)
	return
}

func getAESKey(secret string) (key *[AESKeyLength]byte, err error) {
	if len(secret) < AESKeyLength {
		return nil, fmt.Errorf("secret key length is too short, require [%d] or more", AESKeyLength)
	}

	key = &[AESKeyLength]byte{}
	copy(key[:], secret[0:AESKeyLength])
	return
}

// BoolRef returns a reference to a bool value
func BoolRef(v bool) *bool {
	return &v
}

// StrRef returns a reference to a string value
func StrRef(v string) *string {
	return &v
}

// Int32Ref returns a reference to a int32 value
func Int32Ref(v int32) *int32 {
	return &v
}

// Int64Ref returns a reference to a int64 value
func Int64Ref(v int64) *int64 {
	return &v
}

// UintRef returns a reference to a uint value
func UintRef(v uint) *uint {
	return &v
}

// Uint32Ref returns a reference to a uint32 value
func Uint32Ref(v uint32) *uint32 {
	return &v
}

// TimeRef return a reference to time value
func TimeRef(v time.Time) *time.Time {
	if v.IsZero() {
		return nil
	}
	return &v
}

// IDToRef convert defined ID to reference
func IDToRef(id uint, prefix string, length int, refChars, refPaddingChars string) (string, error) {
	ref, err := Int2String(
		int64(id), refChars, refPaddingChars, length-len(prefix))
	if err != nil {
		return "", fmt.Errorf("can not generate ref for ID %d with prefix %s and length %d", id, prefix, length)
	}
	return prefix + ref, nil
}

// RefToID convert defined reference to id
func RefToID(ref, prefix string, refChars, refPaddingChars string) (int64, error) {
	ref = ref[len(prefix):]

	n, err := String2Int(ref, refChars, refPaddingChars)
	if err != nil {
		return int64(0), fmt.Errorf("can not convert ref %s with prefix %s back to ID", ref, prefix)
	}
	return n, nil
}

func verifyInputChars(chars, paddingChars string) error {
	if len(chars) < 10 {
		return fmt.Errorf("invalid input: chars %v has length < 10", chars)
	}

	if paddingChars == "" {
		return fmt.Errorf("invalid input: empty paddingChars")
	}

	for _, c := range chars {
		if strings.Contains(paddingChars, string(c)) {
			return fmt.Errorf("invalid input: character %s appears in both chars & paddingChars", string(c))
		}
	}
	return nil
}
