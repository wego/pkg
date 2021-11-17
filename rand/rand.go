package rand

import (
	"fmt"
	"github.com/wego/pkg/errors"
	"math"
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

	numbers                      = "0123456789"
	letters                      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerLetters                 = "abcdefghijklmnopqrstuvwxyz"
	upperLetters                 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbersAndUpperLetters       = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberAndLowerLetters        = "0123456789abcdefghijklmnopqrstuvwxyz"
	numbersAndLetters            = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbersLength                = 10
	lettersLength                = 26
	lowerLettersLength           = 26
	upperLettersLength           = 26
	numbersAndLowerLettersLength = 36
	numbersAndUpperLettersLength = 36
	numbersAndLettersLen         = 62

	charIdxBits = 6 // 6 bits to represent a letter index, for the biggest case, numbers and letters is 62
	charIdxMask = 1<<charIdxBits - 1
	charIdxMax  = 63 / charIdxBits // # of letter indices fitting in 63 bits
	// this concurrent is chosen based on the benchmark result, if too small or too large, the performance will be bad
	// the test result on a 4~56 cores container, the best range is from about 30 - around 100, choose 64 for it's easier to calculate
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
		Letters | Lower | Upper:   letters,
		Upper:                     upperLetters,
		Lower:                     lowerLetters,
		Numbers | Letters:         numbersAndLetters,
		Numbers | Letters | Upper: numbersAndUpperLetters,
		Numbers | Upper:           numbersAndUpperLetters,
		Numbers | Letters | Lower: numberAndLowerLetters,
		Numbers | Lower:           numberAndLowerLetters,
	}

	optionMappingLen = map[int]int{
		Numbers:                   numbersLength,
		Letters:                   lettersLength,
		Letters | Lower | Upper:   lettersLength,
		Upper:                     upperLettersLength,
		Lower:                     lowerLettersLength,
		Numbers | Letters:         numbersAndLettersLen,
		Numbers | Letters | Upper: numbersAndUpperLettersLength,
		Numbers | Upper:           numbersAndUpperLettersLength,
		Numbers | Letters | Lower: numbersAndLowerLettersLength,
		Numbers | Lower:           numbersAndLowerLettersLength,
	}

	optionNamesMapping = map[int]string{
		Numbers:                   "Numbers",
		Letters:                   "Letters",
		Letters | Lower | Upper:   "Letters",
		Upper:                     "Upper",
		Lower:                     "Lower",
		Numbers | Letters:         "NumbersAndLetters",
		Numbers | Letters | Upper: "NumbersAndUpperLetters",
		Numbers | Upper:           "NumbersAndUpperLetters",
		Numbers | Letters | Lower: "NumberAndLowerLetters",
		Numbers | Lower:           "NumberAndLowerLetters",
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
// NOTE:
// this function will not check the options, length, so you should make sure the options and length are valid and
// correct before calling this function.
func StringWithOption(length, option int, prefix, suffix string) string {
	source, ok := optionMapping[option]
	if !ok {
		source = numbersAndLetters
	}
	prefixLength, suffixLength := len(prefix), len(suffix)
	totalLength := length + prefixLength + suffixLength
	buf := make([]byte, totalLength)

	if prefixLength > 0 {
		copy(buf, prefix)
	}
	if suffixLength > 0 {
		copy(buf[totalLength-len(suffix):], suffix)
	}

	// 63 random bits, enough for charIdxMax characters!
	for i, start, cache, remain := length-1, prefixLength, Int63(), charIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = Int63(), charIdxMax
		}
		if idx := int(cache & charIdxMask); idx < len(source) {
			buf[i+start] = source[idx]
			i--
		}
		cache >>= charIdxBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&buf))
}

// CheckOption checks if the given option is valid.
func CheckOption(option, length, numbers int) error {
	seedLength, ok := optionMappingLen[option]
	if !ok {
		return errors.New(errors.Unprocessable, fmt.Sprintf("invalid option: %d", option))
	}

	min := int(math.Ceil(math.Log(float64(numbers+1)) / math.Log(float64(seedLength))))
	if min == 0 {
		min = 1
	}
	if length < min {
		return errors.New(errors.Unprocessable,
			fmt.Sprintf("can not generate %v %v codes with length %v, minimal length should be %v",
				numbers, optionNamesMapping[option], length, min))

	}
	return nil
}
