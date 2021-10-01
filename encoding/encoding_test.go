package encoding_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/encoding"
)

// using hexa characters for convert number, xyz for padding
const (
	chars        = "0123456789abcdef"
	paddingChars = "xyz"
	prefix       = "p"
)

func Test_Uint2String_InvalidChars(t *testing.T) {
	assert := assert.New(t)

	hexa10, err := encoding.Uint2String(10, "abc", paddingChars, 1)
	assert.Error(err)
	assert.Equal("invalid input: chars [abc] has length < 10", err.Error())
	assert.Equal("", hexa10)
}

func Test_Uint2String_InvalidPaddingChars(t *testing.T) {
	assert := assert.New(t)

	hexa10, err := encoding.Uint2String(10, chars, "", 1)
	assert.Error(err)
	assert.Equal("invalid input: empty paddingChars", err.Error())
	assert.Equal("", hexa10)
}

func Test_Uint2String_CharsOverlapPaddingChars(t *testing.T) {
	assert := assert.New(t)

	hexa10, err := encoding.Uint2String(10, chars, "pxmap", 1)
	assert.Error(err)
	assert.Equal("invalid input: character [a] appears in both chars & paddingChars", err.Error())
	assert.Equal("", hexa10)
}

func Test_Uint2String_InvalidLength(t *testing.T) {
	assert := assert.New(t)

	hexa10, err := encoding.Uint2String(10, chars, paddingChars, 0)
	assert.Error(err)
	assert.Equal("invalid input: length <= 0", err.Error())
	assert.Equal("", hexa10)
}

func Test_Uint2String_Overlength(t *testing.T) {
	assert := assert.New(t)

	hexa9b, err := encoding.Uint2String(9876543210, chars, paddingChars, 5)
	assert.Error(err)
	assert.Equal("overlength: result [24cb016ea] is longer than required length [5]", err.Error())
	assert.Equal("", hexa9b)
}

func Test_Uint2String_String2Int_NoPadding(t *testing.T) {
	assert := assert.New(t)

	hexa15, err := encoding.Uint2String(15, chars, paddingChars, 1)
	assert.NoError(err)
	assert.Equal("f", hexa15)

	input, err := encoding.String2Uint(hexa15, chars, paddingChars)
	assert.NoError(err)
	assert.EqualValues(15, input)

	hexa2020, err := encoding.Uint2String(2020, chars, paddingChars, 3)
	assert.NoError(err)
	assert.Equal("7e4", hexa2020)

	input, err = encoding.String2Uint(hexa2020, chars, paddingChars)
	assert.NoError(err)
	assert.EqualValues(2020, input)
}

func Test_Uint2String_String2Int_WithPadding(t *testing.T) {
	assert := assert.New(t)

	hexa2020, err := encoding.Uint2String(2020, chars, paddingChars, 7)
	assert.NoError(err)
	// hexa2020 will be ????7e4, with ???? contains characters from paddingChars
	padding := hexa2020[:7-3]
	for i := range padding {
		assert.Contains(paddingChars, string(padding[i]))
	}
	assert.Equal("7e4", hexa2020[7-3:])

	input, err := encoding.String2Uint(hexa2020, chars, paddingChars)
	assert.NoError(err)
	assert.EqualValues(2020, input)

	hexa9b, err := encoding.Uint2String(9876543210, chars, paddingChars, 10)
	assert.NoError(err)
	// hexa9b will be ?24cb016ea, with ? contains 1 character from paddingChars
	padding = hexa9b[:10-9]
	for i := range padding {
		assert.Contains(paddingChars, string(padding[i]))
	}
	assert.Equal("24cb016ea", hexa9b[10-9:])

	input, err = encoding.String2Uint(hexa9b, chars, paddingChars)
	assert.NoError(err)
	assert.EqualValues(9876543210, input)
}

func Test_IDToRef_Ok(t *testing.T) {
	assert := assert.New(t)

	ref, err := encoding.IDToRef(9876543210, prefix, 10, chars, paddingChars)
	assert.NoError(err)
	assert.Equal(prefix+"24cb016ea", ref)
}

func Test_IDToRef_Error(t *testing.T) {
	assert := assert.New(t)

	ref, err := encoding.IDToRef(9876543210, prefix, 3, chars, paddingChars)
	assert.Error(err)
	assert.Equal(
		"error generating ref for id [9876543210] with prefix [p] and length [3]: overlength: result [24cb016ea] is longer than required length [2]",
		err.Error())
	assert.Empty(ref)
}

func Test_RefToID_Ok(t *testing.T) {
	assert := assert.New(t)

	id, err := encoding.RefToID(prefix+"24cb016ea", prefix, chars, paddingChars)
	assert.NoError(err)
	assert.EqualValues(9876543210, id)
}

func Test_RefToID_Error(t *testing.T) {
	assert := assert.New(t)

	id, err := encoding.RefToID(prefix+"24cb016ea", prefix, chars, paddingChars+"a")
	assert.Error(err)
	assert.Equal("can not convert ref [24cb016ea] with prefix [p] back to id: invalid input: character [a] appears in both chars & paddingChars",
		err.Error())
	assert.Zero(id)
}
