package cognito

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	wegostrings "github.com/wego/pkg/strings"
)

// expiryLeeway is shaved off the real expiry so a token cannot lapse midway
// through a request we have already started.
const expiryLeeway = 60 * time.Second

// TokenSet is the set of Cognito tokens a signed-in operator holds.
//
// Holding tokens is the whole point of the type, so gosec's secret-field
// warning is expected here rather than a finding: these values are never
// logged, and the only place they are serialized to is the operator's own
// keychain (see github.com/wego/pkg/cognito/storage).
type TokenSet struct {
	AccessToken  string    `json:"access_token"` //nolint:gosec // G117: this type exists to carry tokens; stored only in the local keychain, never logged.
	IDToken      string    `json:"id_token"`
	RefreshToken string    `json:"refresh_token"` //nolint:gosec // G117: as above - the refresh token is the credential this type is for.
	ExpiresAt    time.Time `json:"expires_at"`
}

// IsExpired reports whether the access token is spent as of now, treating
// anything inside expiryLeeway of the deadline as already gone. A nil set, or
// one with no recorded expiry, counts as expired so callers refresh rather
// than send a token that may already be dead.
func (t *TokenSet) IsExpired(now time.Time) bool {
	if t == nil || t.ExpiresAt.IsZero() {
		return true
	}
	return !now.Before(t.ExpiresAt.Add(-expiryLeeway))
}

// Email returns the email claim from the id_token.
//
// It reads the JWT payload WITHOUT verifying the signature, which is correct
// here: the token came from our own keychain, and every server that accepts it
// verifies the signature itself. Treat this as decoding for display and
// client-side UX, NOT as validation - a caller must never make an
// authorization decision on the strength of these claims.
func (t *TokenSet) Email() (string, error) {
	if t == nil {
		return "", errors.New("no tokens: sign in first")
	}

	claims, err := decodeIDTokenClaims(t.IDToken)
	if err != nil {
		return "", err
	}

	email, ok := claims["email"].(string)
	if !ok || wegostrings.IsBlank(email) {
		return "", errors.New("id_token has no email claim")
	}
	return email, nil
}

// decodeIDTokenClaims splits a JWT and JSON-decodes its payload. It never
// panics on a malformed or truncated token: every failure is an error.
func decodeIDTokenClaims(idToken string) (map[string]any, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("id_token is not a jwt")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// Some encoders pad their segments; accept those too.
		payload, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, fmt.Errorf("decode id_token payload: %w", err)
		}
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parse id_token payload: %w", err)
	}
	return claims, nil
}
