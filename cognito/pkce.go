package cognito

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	wegostrings "github.com/wego/pkg/strings"
)

const (
	// verifierBytes yields a 128-character unpadded base64url verifier, the
	// longest RFC 7636 permits for code_verifier.
	verifierBytes = 96

	// stateBytes yields a 32-character unpadded base64url state value.
	stateBytes = 24
)

// generateVerifier draws a fresh PKCE code_verifier from crypto/rand.
func generateVerifier() (string, error) {
	buf := make([]byte, verifierBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate pkce verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// generateChallenge derives the S256 code_challenge for a verifier. Both the
// verifier and the challenge are unpadded base64url, per RFC 7636.
func generateChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// generateState draws a fresh OAuth state value from crypto/rand.
func generateState() (string, error) {
	buf := make([]byte, stateBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// stateMatches compares the state a callback presented against the one we
// minted, in constant time. A blank value on either side never matches: a
// callback that omits state is precisely the CSRF case the parameter exists to
// catch, so it must be rejected rather than waved through.
func stateMatches(want, got string) bool {
	if wegostrings.IsBlank(want) || wegostrings.IsBlank(got) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(want), []byte(got)) == 1
}
