package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_maskAuthorizationHeader(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "Bearer with sk_test_ prefix",
			value: "Bearer sk_test_1234567890",
			want:  "Bearer sk_test_12***890",
		},
		{
			name:  "Bearer with sk_ prefix",
			value: "Bearer sk_1234567890",
			want:  "Bearer sk_12***890",
		},
		{
			name:  "Bearer with pk_test_ prefix",
			value: "Bearer pk_test_1234567890",
			want:  "Bearer pk_test_12***890",
		},
		{
			name:  "Bearer with pk_ prefix",
			value: "Bearer pk_1234567890",
			want:  "Bearer pk_12***890",
		},
		{
			name:  "Bearer with no prefix",
			value: "Bearer 1234567890",
			want:  "Bearer 12***890",
		},
		{
			name:  "Basic auth",
			value: "Basic 123=4567890",
			want:  "Basic 12***890",
		},
		{
			name:  "No auth type",
			value: "1234567890",
			want:  "12***890",
		},
		{
			name:  "With empty space at start",
			value: " 1234567890",
			want:  " 12***890",
		},
	}

	assert := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := maskAuthorizationHeader(tc.value)

			assert.Equal(tc.want, got)
		})
	}
}
