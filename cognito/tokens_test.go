package cognito_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wegostrings "github.com/wego/pkg/strings"

	"github.com/wego/pkg/cognito"
)

const testOperatorEmail = "ops@wego.com"

var fixedNow = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

func TestTokenSet_IsExpired(t *testing.T) {
	tests := []struct {
		name  string
		given *cognito.TokenSet
		want  bool
	}{
		{
			name:  "nil token set counts as expired",
			given: nil,
			want:  true,
		},
		{
			name:  "zero expiry counts as expired",
			given: &cognito.TokenSet{},
			want:  true,
		},
		{
			name:  "already elapsed",
			given: &cognito.TokenSet{ExpiresAt: fixedNow.Add(-time.Second)},
			want:  true,
		},
		{
			name:  "inside the safety skew",
			given: &cognito.TokenSet{ExpiresAt: fixedNow.Add(30 * time.Second)},
			want:  true,
		},
		{
			name:  "exactly at the safety skew",
			given: &cognito.TokenSet{ExpiresAt: fixedNow.Add(60 * time.Second)},
			want:  true,
		},
		{
			name:  "just beyond the safety skew",
			given: &cognito.TokenSet{ExpiresAt: fixedNow.Add(61 * time.Second)},
			want:  false,
		},
		{
			name:  "an hour of life left",
			given: &cognito.TokenSet{ExpiresAt: fixedNow.Add(time.Hour)},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.given.IsExpired(fixedNow), "evaluated at %v", fixedNow)
		})
	}
}

func TestTokenSet_Email(t *testing.T) {
	tests := []struct {
		name            string
		givenIDToken    string
		want            string
		wantErrContains string
	}{
		{
			name:            "blank id_token is rejected",
			givenIDToken:    "",
			wantErrContains: "not a jwt",
		},
		{
			name:            "two-segment token is not a jwt",
			givenIDToken:    "header.payload",
			wantErrContains: "not a jwt",
		},
		{
			name:            "truncated payload does not panic",
			givenIDToken:    "header.!!!not-base64!!!.signature",
			wantErrContains: "decode",
		},
		{
			name:            "payload is not json",
			givenIDToken:    "header." + rawB64("this is not json") + ".signature",
			wantErrContains: "parse",
		},
		{
			name:            "payload is a json array not an object",
			givenIDToken:    "header." + rawB64(`["nope"]`) + ".signature",
			wantErrContains: "parse",
		},
		{
			name:            "no email claim",
			givenIDToken:    "header." + rawB64(`{"sub":"abc"}`) + ".signature",
			wantErrContains: "no email claim",
		},
		{
			name:            "email claim is not a string",
			givenIDToken:    "header." + rawB64(`{"email":42}`) + ".signature",
			wantErrContains: "no email claim",
		},
		{
			name:            "email claim is blank",
			givenIDToken:    "header." + rawB64(`{"email":""}`) + ".signature",
			wantErrContains: "no email claim",
		},
		{
			name:         "padded base64 payload still decodes",
			givenIDToken: "header." + base64.URLEncoding.EncodeToString([]byte(`{"email":"`+testOperatorEmail+`"}`)) + ".signature",
			want:         testOperatorEmail,
		},
		{
			name:         "email is returned",
			givenIDToken: "header." + rawB64(`{"email":"`+testOperatorEmail+`"}`) + ".signature",
			want:         testOperatorEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &cognito.TokenSet{IDToken: tt.givenIDToken}

			var (
				got string
				err error
			)
			require.NotPanics(t, func() { got, err = ts.Email() }, "Email must never panic on a malformed id_token")

			if wegostrings.IsNotEmpty(tt.wantErrContains) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				assert.Empty(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTokenSet_EmailOnNilReceiver(t *testing.T) {
	var ts *cognito.TokenSet

	var err error
	require.NotPanics(t, func() { _, err = ts.Email() })
	require.Error(t, err)
}

// rawB64 is unpadded base64url, the encoding JWT segments use.
func rawB64(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

// mustIDToken builds an unsigned JWT carrying claims. The signature is
// deliberately junk: Email() must not verify it.
func mustIDToken(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	require.NoError(t, err)
	return rawB64(`{"alg":"RS256","typ":"JWT"}`) + "." +
		base64.RawURLEncoding.EncodeToString(payload) + ".not-a-real-signature"
}
