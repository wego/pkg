package rand

import (
	"fmt"
	"github.com/wego/pkg/errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// options to generate random string
const (
	Numbers = 1 << iota
	Letters = 1 << iota
	Upper   = 1 << iota
	Lower   = 1 << iota

	numbers                = "0123456789"
	letters                = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerLetters           = "abcdefghijklmnopqrstuvwxyz"
	upperLetters           = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbersAndUpperLetters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberAndLowerLetters  = "0123456789abcdefghijklmnopqrstuvwxyz"
	numbersAndLetters      = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbersAndLettersLen   = 62

	charIdxBits    = 6 // 6 bits to represent a letter index, for the biggest case, numbers and letters is 62
	charIdxMask    = 1<<charIdxBits - 1
	charIdxMax     = 63 / charIdxBits // # of letter indices fitting in 63 bits
	concurrent     = 64
	concurrentMask = concurrent - 1
)

var (
	mutexes        = make([]sync.Mutex, concurrent)
	rands          = make([]rand.Source, concurrent)
	index   uint32 = 0

	optionMapping = map[int]string{
		Numbers:                   numbers,
		Letters:                   letters,
		Upper:                     upperLetters,
		Lower:                     lowerLetters,
		Numbers | Letters:         numbersAndLetters,
		Numbers | Letters | Upper: numbersAndUpperLetters,
		Numbers | Upper:           numbersAndUpperLetters,
		Numbers | Letters | Lower: numberAndLowerLetters,
		Numbers | Lower:           numberAndLowerLetters,
	}
)

func init() {
	for i := 0; i < concurrent; i++ {
		rands[i] = rand.NewSource(time.Now().UnixNano() + rand.Int63())
	}
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func Int63() int64 {
	current := atomic.AddUint32(&index, 1) & concurrentMask
	mutexes[current].Lock()
	defer mutexes[current].Unlock()
	return rands[current].Int63()
}

// Uint64 returns a random uint64 value.
func Uint64() uint64 {
	r := Int63()
	return uint64(r)>>31 | uint64(r)<<32
}

// String same as StringWithOption(length, Numbers | Letters),
// use a separate function with constants to improve performance.
func String(length int) string {
	buf := make([]byte, length)
	// 63 random bits, enough for charIdxMax characters!
	for i, cache, remain := length-1, Int63(), charIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = Int63(), charIdxMax
		}
		if idx := int(cache & charIdxMask); idx < numbersAndLettersLen {
			buf[i] = numbersAndLetters[idx]
			i--
		}
		cache >>= charIdxBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&buf))
}

// StringWithOption returns a random string with the given length and options.
// The options can be combined using bitwise OR operation.
// The basic idea of this function is to generate a random number and map it to a character in the given options.
// This algorithm and idea is from https://stackoverflow.com/a/31832326 with an optimization for thead-safe by
// pre-allocating a bulk of mutexes and rand.Source.
func StringWithOption(length int, option int) (string, error) {
	source, ok := optionMapping[option]
	if !ok {
		return "", errors.New(errors.NotSupported, fmt.Sprintf("option %v is not supported", option))
	}

	buf := make([]byte, length)
	// 63 random bits, enough for charIdxMax characters!
	for i, cache, remain := length-1, Int63(), charIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = Int63(), charIdxMax
		}
		if idx := int(cache & charIdxMask); idx < len(source) {
			buf[i] = source[idx]
			i--
		}
		cache >>= charIdxBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&buf)), nil
}
