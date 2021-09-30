package encoding

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/wego/pkg/errors"
)

// Uint2String converts a number to a string using characters from chars.
//
// Validation:
//  - `length` must > 0
//  - length of `chars` must >= 10
//  - `paddingChars` must not be empty
//  - `chars` & `paddingChars` must not overlap
// If length of the result < required length, the result will be left padded with random characters from `paddingChars`.
func Uint2String(number uint64, chars, paddingChars string, length int) (s string, err error) {
	if length < 1 {
		return s, errors.New("invalid input: length <= 0")
	}
	err = verifyInputChars(chars, paddingChars)
	if err != nil {
		return
	}

	base := uint64(len(chars))
	for number > 0 {
		i := number % base
		s = string(chars[i]) + s
		number = number / base
	}

	resultLength := len(s)
	if resultLength > length {
		return "", fmt.Errorf("overlength: result [%s] is longer than required length [%d]", s, length)
	}

	// left padding if not meet the required length
	for ; resultLength < length; resultLength++ {
		s = string(paddingChars[rand.Intn(len(paddingChars))]) + s
	}

	return s, nil
}

// String2Uint is the revert of Uint2String
//  It returns negative result in 2 cases:
//  - s contains character not exist in chars
//  - the result is bigger than max int64 (integer overflow happen)
func String2Uint(s string, chars, paddingChars string) (n uint64, err error) {
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
		indexValue := uint64(strings.IndexByte(chars, s[i]))
		exponent := len(s) - 1 - i
		n += indexValue * uint64(math.Pow(float64(base), float64(exponent)))
	}
	return
}

// IDToRef convert defined ID to reference
func IDToRef(id uint, prefix string, length int, refChars, refPaddingChars string) (string, error) {
	ref, err := Uint2String(
		uint64(id), refChars, refPaddingChars, length-len(prefix))
	if err != nil {
		return "", errors.New(fmt.Sprintf("error generating ref for id [%d] with prefix [%s] and length [%d]", id, prefix, length), err)
	}
	return prefix + ref, nil
}

// RefToID convert defined reference to id
func RefToID(ref, prefix string, refChars, refPaddingChars string) (uint64, error) {
	ref = ref[len(prefix):]

	n, err := String2Uint(ref, refChars, refPaddingChars)
	if err != nil {
		return uint64(0), errors.New(fmt.Sprintf("can not convert ref [%s] with prefix [%s] back to id", ref, prefix), err)
	}
	return n, nil
}

func verifyInputChars(chars, paddingChars string) error {
	if len(chars) < 10 {
		return fmt.Errorf("invalid input: chars [%v] has length < 10", chars)
	}

	if paddingChars == "" {
		return errors.New("invalid input: empty paddingChars")
	}

	for _, c := range chars {
		if strings.Contains(paddingChars, string(c)) {
			return fmt.Errorf("invalid input: character [%s] appears in both chars & paddingChars", string(c))
		}
	}
	return nil
}
