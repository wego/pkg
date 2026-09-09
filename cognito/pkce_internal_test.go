package cognito

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateVerifier(t *testing.T) {
	verifier, err := generateVerifier()
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(verifier), 43, "RFC 7636 requires a code_verifier of at least 43 chars, got %q", verifier)
	assert.LessOrEqual(t, len(verifier), 128, "RFC 7636 caps code_verifier at 128 chars, got %d chars", len(verifier))
	assert.NotContains(t, verifier, "=", "code_verifier must be base64url WITHOUT padding")
	assert.NotContains(t, verifier, "+", "code_verifier must be base64url, not standard base64")
	assert.NotContains(t, verifier, "/", "code_verifier must be base64url, not standard base64")

	_, err = base64.RawURLEncoding.DecodeString(verifier)
	require.NoError(t, err, "code_verifier must decode as unpadded base64url")
}

func TestGenerateVerifier_IsUniquePerCall(t *testing.T) {
	seen := make(map[string]struct{}, 32)
	for range 32 {
		verifier, err := generateVerifier()
		require.NoError(t, err)
		_, dup := seen[verifier]
		require.False(t, dup, "generateVerifier returned a duplicate; it must be drawn from crypto/rand")
		seen[verifier] = struct{}{}
	}
}

func TestGenerateChallenge(t *testing.T) {
	tests := []struct {
		name          string
		givenVerifier string
		want          string
	}{
		{
			// RFC 7636 appendix B test vector - pins the S256 transformation.
			name:          "rfc 7636 appendix b vector",
			givenVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			want:          "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		},
		{
			name:          "empty verifier still hashes",
			givenVerifier: "",
			want:          base64.RawURLEncoding.EncodeToString(sha256Sum("")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateChallenge(tt.givenVerifier)
			assert.Equal(t, tt.want, got, "challenge for verifier %q", tt.givenVerifier)
			assert.NotContains(t, got, "=", "code_challenge must be unpadded base64url")
		})
	}
}

func TestGenerateState(t *testing.T) {
	state, err := generateState()
	require.NoError(t, err)

	require.NotEmpty(t, state)
	assert.NotContains(t, state, "=", "state must be unpadded base64url")
	_, err = base64.RawURLEncoding.DecodeString(state)
	require.NoError(t, err, "state must decode as unpadded base64url")

	other, err := generateState()
	require.NoError(t, err)
	assert.NotEqual(t, state, other, "state must be freshly drawn from crypto/rand on every call")
}

func TestStateMatches(t *testing.T) {
	tests := []struct {
		name      string
		givenWant string
		givenGot  string
		want      bool
	}{
		{name: "mismatch", givenWant: "abc", givenGot: "abd", want: false},
		{name: "missing callback state", givenWant: "abc", givenGot: "", want: false},
		{name: "both blank is still a mismatch", givenWant: "", givenGot: "", want: false},
		{name: "different length", givenWant: "abc", givenGot: "abcdef", want: false},
		{name: "match", givenWant: "abc", givenGot: "abc", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, stateMatches(tt.givenWant, tt.givenGot))
		})
	}
}

// sha256Sum computes the expectation independently of the code under test.
func sha256Sum(s string) []byte {
	sum := sha256.Sum256([]byte(s))
	return sum[:]
}
