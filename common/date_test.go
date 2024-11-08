package common_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/common"
)

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected common.Date
		hasError bool
	}{
		{
			name:     "Invalid date format",
			input:    []byte(`"2020-30-02"`),
			expected: common.Date{},
			hasError: true,
		},
		{
			name:     "Empty date",
			input:    []byte(`""`),
			expected: common.Date{},
			hasError: true,
		},
		{
			name:     "Normal date without quotes",
			input:    []byte(`1999-12-31`),
			expected: common.Date(time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC)),
			hasError: false,
		},
		{
			name:     "Normal date with quotes",
			input:    []byte(`"1999-12-31"`),
			expected: common.Date(time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC)),
			hasError: false,
		},
	}

	assert := assert.New(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result common.Date
			err := result.UnmarshalJSON(tc.input)

			if tc.hasError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, result)
		})
	}
}

func TestDate_MarshallJson(t *testing.T) {
	assert := assert.New(t)

	date := common.Date(time.Date(1999, 12, 31, 11, 11, 11, 0, time.Local))
	result, err := date.MarshalJSON()
	assert.NoError(err)
	assert.Equal([]byte(`"1999-12-31"`), result)
}

func TestDate_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected common.Date
		hasError bool
	}{
		{
			name:     "Invalid date format",
			input:    []byte(`"2020-30-02"`),
			expected: common.Date{},
			hasError: true,
		},
		{
			name:     "Empty date",
			input:    []byte(`""`),
			expected: common.Date{},
			hasError: true,
		},
		{
			name:     "Normal date without quotes",
			input:    []byte(`1999-12-31`),
			expected: common.Date(time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC)),
			hasError: false,
		},
		{
			name:     "Normal date with quotes",
			input:    []byte(`"1999-12-31"`),
			expected: common.Date(time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC)),
			hasError: false,
		},
	}

	assert := assert.New(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result common.Date
			err := result.UnmarshalText(tc.input)

			if tc.hasError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, result)
		})
	}
}

func TestDate_MarshallText(t *testing.T) {
	assert := assert.New(t)

	date := common.Date(time.Date(1999, 12, 31, 11, 11, 11, 0, time.Local))
	result, err := date.MarshalText()
	assert.NoError(err)
	assert.Equal([]byte(`"1999-12-31"`), result)
}

func TestDate_String(t *testing.T) {
	tests := []struct {
		name     string
		date     common.Date
		expected string
	}{
		{
			name:     "single digit day",
			date:     common.Date(time.Date(2023, 10, 5, 11, 11, 11, 0, time.Local)),
			expected: "2023-10-05",
		},
		{
			name:     "double digit month and day",
			date:     common.Date(time.Date(1999, 12, 31, 11, 11, 11, 0, time.Local)),
			expected: "1999-12-31",
		},
		{
			name:     "single digit month and day",
			date:     common.Date(time.Date(2000, 1, 1, 11, 11, 11, 0, time.Local)),
			expected: "2000-01-01",
		},
	}

	assert := assert.New(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.date.String()
			assert.Equal(tc.expected, result)
		})
	}
}

func TestDate_Before(t *testing.T) {
	assert := assert.New(t)

	now := time.Now()
	assert.True(common.Date(now).Before(common.Date(now.Add(time.Hour))))
	assert.False(common.Date(now.Add(time.Hour)).Before(common.Date(now)))
}

func TestDate_IsZero(t *testing.T) {
	assert := assert.New(t)

	assert.True(common.Date{}.IsZero())
	assert.False(common.Date(time.Now()).IsZero())
}

func TestDate_Equal(t *testing.T) {
	assert := assert.New(t)

	now := time.Now()
	assert.True(common.Date(now).Equal(now))
	assert.False(common.Date(now).Equal(now.Add(time.Hour)))
}
