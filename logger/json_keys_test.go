package logger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/logger"
)

const filtered = "[Filtered by Wego]"

func Test_RedactJSON_NonStringValues(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    string
		keys     [][]string
		expected string
	}{
		{
			name:     "nested number",
			input:    `{"customer":{"latitude":12.971598,"name":"bob"}}`,
			keys:     [][]string{{"customer", "latitude"}},
			expected: `{"customer":{"latitude":"` + filtered + `","name":"bob"}}`,
		},
		{
			name:     "nested negative number",
			input:    `{"customer":{"longitude":-77.0364}}`,
			keys:     [][]string{{"customer", "longitude"}},
			expected: `{"customer":{"longitude":"` + filtered + `"}}`,
		},
		{
			name:     "nested boolean",
			input:    `{"customer":{"resident":true}}`,
			keys:     [][]string{{"customer", "resident"}},
			expected: `{"customer":{"resident":"` + filtered + `"}}`,
		},
		{
			name:     "nested null",
			input:    `{"customer":{"pan":null}}`,
			keys:     [][]string{{"customer", "pan"}},
			expected: `{"customer":{"pan":"` + filtered + `"}}`,
		},
		{
			name:     "nested object",
			input:    `{"customer":{"address":{"line1":"1 Main St"}}}`,
			keys:     [][]string{{"customer", "address"}},
			expected: `{"customer":{"address":"` + filtered + `"}}`,
		},
		{
			name:     "nested array",
			input:    `{"customer":{"tokens":["a","b"]}}`,
			keys:     [][]string{{"customer", "tokens"}},
			expected: `{"customer":{"tokens":"` + filtered + `"}}`,
		},
		{
			name:     "number inside array items",
			input:    `{"devices":[{"latitude":1.5},{"latitude":2.5}]}`,
			keys:     [][]string{{"devices", "[]", "latitude"}},
			expected: `{"devices":[{"latitude":"` + filtered + `"},{"latitude":"` + filtered + `"}]}`,
		},
		{
			name:     "array items themselves",
			input:    `{"tokens":["a",2,null]}`,
			keys:     [][]string{{"tokens", "[]"}},
			expected: `{"tokens":["` + filtered + `","` + filtered + `","` + filtered + `"]}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, logger.RedactJSON(tc.input, "", tc.keys))
		})
	}
}

func Test_RedactJSON_CaseInsensitiveKeys(t *testing.T) {
	input := `{"Customer":{"LATITUDE":12.971598,"Longitude":77.594566}}`
	keys := [][]string{{"customer", "latitude"}, {"customer", "longitude"}}

	t.Run("off by default, and adds no key of its own", func(t *testing.T) {
		assert.Equal(t, input, logger.RedactJSON(input, "", keys))
	})

	t.Run("on, keeping each spelling", func(t *testing.T) {
		assert.Equal(t,
			`{"Customer":{"LATITUDE":"`+filtered+`","Longitude":"`+filtered+`"}}`,
			logger.RedactJSON(input, "", keys, logger.CaseInsensitiveKeys()),
		)
	})

	t.Run("distinct casings of the same key are all redacted", func(t *testing.T) {
		assert.Equal(t,
			`{"customer":{"latitude":"`+filtered+`","Latitude":"`+filtered+`"}}`,
			logger.RedactJSON(
				`{"customer":{"latitude":1.5,"Latitude":2.5}}`, "",
				[][]string{{"customer", "latitude"}},
				logger.CaseInsensitiveKeys(),
			),
		)
	})
}

func Test_RedactJSON_DuplicatedMembers(t *testing.T) {
	t.Run("duplicated leaf collapses into one redacted member", func(t *testing.T) {
		assert.Equal(t,
			`{"cvv":"`+filtered+`"}`,
			logger.RedactJSON(`{"cvv":"123","cvv":"456"}`, "", [][]string{{"cvv"}}),
		)
	})

	t.Run("every duplicated parent is walked into", func(t *testing.T) {
		assert.Equal(t,
			`{"customer":{"latitude":"`+filtered+`"},"customer":{"latitude":"`+filtered+`"}}`,
			logger.RedactJSON(
				`{"customer":{"latitude":1.5},"customer":{"latitude":2.5}}`, "",
				[][]string{{"customer", "latitude"}},
			),
		)
	})

	t.Run("a key present once keeps its position", func(t *testing.T) {
		assert.Equal(t,
			`{"a":1,"cvv":"`+filtered+`","z":2}`,
			logger.RedactJSON(`{"a":1,"cvv":"123","z":2}`, "", [][]string{{"cvv"}}),
		)
	})
}

func Test_RedactJSON_ReplacementIsEscaped(t *testing.T) {
	assert.Equal(t,
		`{"cvv":"say \"no\"\\"}`,
		logger.RedactJSON(`{"cvv":"123"}`, `say "no"\`, [][]string{{"cvv"}}),
	)
}

func Test_MaskJSON_NonStringValues(t *testing.T) {
	show := func(keys ...string) logger.MaskData {
		return logger.MaskData{JSONKeys: keys, FirstCharsToShow: 1, LastCharsToShow: 1}
	}

	for _, tc := range []struct {
		name     string
		input    string
		toMask   []logger.MaskData
		expected string
	}{
		{
			name:     "nested number masked through its literal",
			input:    `{"customer":{"latitude":12.971598}}`,
			toMask:   []logger.MaskData{show("customer", "latitude")},
			expected: `{"customer":{"latitude":"1*8"}}`,
		},
		{
			name:     "nested boolean",
			input:    `{"customer":{"resident":true}}`,
			toMask:   []logger.MaskData{show("customer", "resident")},
			expected: `{"customer":{"resident":"t*e"}}`,
		},
		{
			name:     "null is left alone",
			input:    `{"customer":{"pan":null}}`,
			toMask:   []logger.MaskData{show("customer", "pan")},
			expected: `{"customer":{"pan":null}}`,
		},
		{
			name:     "object is left alone",
			input:    `{"customer":{"address":{"line1":"1 Main St"}}}`,
			toMask:   []logger.MaskData{show("customer", "address")},
			expected: `{"customer":{"address":{"line1":"1 Main St"}}}`,
		},
		{
			name:     "array is left alone",
			input:    `{"customer":{"tokens":["abcde"]}}`,
			toMask:   []logger.MaskData{show("customer", "tokens")},
			expected: `{"customer":{"tokens":["abcde"]}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, logger.MaskJSON(tc.input, "", tc.toMask))
		})
	}
}

func Test_MaskJSON_CaseInsensitiveKeys(t *testing.T) {
	input := `{"Customer":{"PAN":"ABCDE1234F"}}`
	toMask := []logger.MaskData{{
		JSONKeys:         []string{"customer", "pan"},
		FirstCharsToShow: 2,
		LastCharsToShow:  1,
	}}

	assert.Equal(t, input, logger.MaskJSON(input, "", toMask))
	assert.Equal(t,
		`{"Customer":{"PAN":"AB*F"}}`,
		logger.MaskJSON(input, "", toMask, logger.CaseInsensitiveKeys()),
	)
}

func Test_MaskJSON_DuplicatedMembers(t *testing.T) {
	toMask := []logger.MaskData{{
		JSONKeys:         []string{"customerPAN"},
		FirstCharsToShow: 2,
		LastCharsToShow:  1,
	}}

	// encoding/json binds the last member, so that is the value that gets masked.
	assert.Equal(t,
		`{"customerPAN":"ZY*V"}`,
		logger.MaskJSON(`{"customerPAN":"ABCDE1234F","customerPAN":"ZYXWV"}`, "", toMask),
	)
}

func Test_MaskJSON_MaskedValueIsEscaped(t *testing.T) {
	assert.Equal(t,
		`{"name":"\"J*r\""}`,
		logger.MaskJSON(`{"name":"\"Joker\""}`, "", []logger.MaskData{{
			JSONKeys:         []string{"name"},
			FirstCharsToShow: 2,
			LastCharsToShow:  2,
		}}),
	)
}
