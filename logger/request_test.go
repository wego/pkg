package logger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/errors"
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			got := maskAuthorizationHeader(tc.value)

			assert.Equal(tc.want, got)
		})
	}
}

func TestRequest_fields(t *testing.T) {
	staticFieldsCount := 11
	requestedAt := time.Now()
	duration := 100 * time.Millisecond

	tests := []struct {
		name    string
		request *Request
		wantLen int // Expected number of fields in the output
		wantErr bool
	}{
		{
			name:    "Empty Request",
			request: &Request{},
			wantLen: staticFieldsCount, // Only the static fields are added
		},
		{
			name: "Full Request",
			request: &Request{
				Basics: map[string]interface{}{
					"nullField":   nil,
					"paymentRef":  "abc123",
					"emptyField1": "",
					"nullField2":  nil,
					"orderRef":    "zxc-456",
					"emptyField2": "",
				},
				Type:            "GET",
				Method:          "POST",
				URL:             "http://example.com",
				RequestHeaders:  Headers{},
				RequestBody:     "request-body",
				IP:              "127.0.0.1",
				StatusCode:      200,
				ResponseHeaders: Headers{},
				ResponseBody:    "response-body",
				RequestedAt:     requestedAt,
				Duration:        duration,
				Error:           errors.New("mock error"),
			},
			wantLen: 14, // All fields including Basics and Error
		},
		{
			name: "Request with Unmarshallable Basics",
			request: &Request{
				Basics: map[string]interface{}{
					"unmarshallable": make(chan int),
				},
			},
			wantLen: staticFieldsCount, // Unmarshallable field is skipped
			wantErr: true,              // Expecting error due to unmarshallable field
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			fields := tc.request.fields()

			assert.Equal(tc.wantLen, len(fields))

			if tc.wantLen > staticFieldsCount {
				// assert basics
				if fields[0].Key == "paymentRef" {
					assert.Equal(`"abc123"`, fields[0].String)
					assert.Equal("orderRef", fields[1].Key)
					assert.Equal(`"zxc-456"`, fields[1].String)
				} else {
					assert.Equal("orderRef", fields[0].Key)
					assert.Equal(`"zxc-456"`, fields[0].String)
					assert.Equal("paymentRef", fields[1].Key)
					assert.Equal(`"abc123"`, fields[1].String)
				}

				// assert static fields
				assert.Equal("GET", fields[2].String)
				assert.Equal("POST", fields[3].String)
				assert.Equal("http://example.com", fields[4].String)
				assert.Equal(Headers{}, fields[5].Interface)
				assert.Equal("request-body", fields[6].String)
				assert.Equal("127.0.0.1", fields[7].String)
				assert.Equal(int64(200), fields[8].Integer)
				assert.Equal(Headers{}, fields[9].Interface)
				assert.Equal("response-body", fields[10].String)
				assert.Equal(requestedAt.Format(time.RFC3339), fields[11].String)
				assert.Equal(duration.Milliseconds(), fields[12].Integer)
				assert.Equal("mock error", fields[13].String)
			}
		})
	}
}
