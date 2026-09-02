package cognito

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenEndpointError(t *testing.T) {
	tests := []struct {
		name            string
		givenStatus     int
		givenBody       string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:        "non-json body is withheld entirely",
			givenStatus: http.StatusInternalServerError,
			givenBody:   "access_token=super-secret-value",
			// An unparseable body could be anything, so we report only the
			// status: a token must never ride out inside an error string.
			wantContains:    []string{"500", "withheld"},
			wantNotContains: []string{"super-secret-value"},
		},
		{
			name:            "json without an error field is withheld",
			givenStatus:     http.StatusBadGateway,
			givenBody:       `{"id_token":"secret-jwt-value"}`,
			wantContains:    []string{"502", "withheld"},
			wantNotContains: []string{"secret-jwt-value"},
		},
		{
			name:         "oauth error is echoed",
			givenStatus:  http.StatusBadRequest,
			givenBody:    `{"error":"invalid_grant"}`,
			wantContains: []string{"400", "invalid_grant"},
		},
		{
			name:         "oauth error and description are echoed",
			givenStatus:  http.StatusBadRequest,
			givenBody:    `{"error":"invalid_grant","error_description":"code expired"}`,
			wantContains: []string{"400", "invalid_grant", "code expired"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tokenEndpointError(tt.givenStatus, []byte(tt.givenBody))

			require.Error(t, err)
			for _, want := range tt.wantContains {
				assert.Contains(t, err.Error(), want)
			}
			for _, notWant := range tt.wantNotContains {
				assert.NotContains(t, err.Error(), notWant)
			}
		})
	}
}

func TestConfig_BuildAuthorizeURL(t *testing.T) {
	tests := []struct {
		name                  string
		givenAuthorizeURL     string
		givenIdentityProvider string
		wantQueryHasProvider  bool
	}{
		{
			name:              "an existing query string is appended to",
			givenAuthorizeURL: "https://cognito.test/oauth2/authorize?foo=bar",
		},
		{
			name:              "no identity provider leaves the parameter out",
			givenAuthorizeURL: "https://cognito.test/oauth2/authorize",
		},
		{
			name:                  "identity provider is forwarded when set",
			givenAuthorizeURL:     "https://cognito.test/oauth2/authorize",
			givenIdentityProvider: "Google",
			wantQueryHasProvider:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				AuthorizeURL:     tt.givenAuthorizeURL,
				ClientID:         "client",
				RedirectURI:      "http://127.0.0.1:8110/callback",
				Scopes:           "openid email",
				IdentityProvider: tt.givenIdentityProvider,
			}

			got, err := url.Parse(cfg.buildAuthorizeURL("the-state", "the-challenge"))
			require.NoError(t, err)

			query := got.Query()
			assert.Equal(t, "the-state", query.Get("state"))
			assert.Equal(t, "the-challenge", query.Get("code_challenge"))
			assert.Equal(t, "S256", query.Get("code_challenge_method"))
			assert.Equal(t, tt.givenIdentityProvider, query.Get("identity_provider"))
			assert.Equal(t, tt.wantQueryHasProvider, query.Has("identity_provider"))

			if tt.givenAuthorizeURL == "https://cognito.test/oauth2/authorize?foo=bar" {
				assert.Equal(t, "bar", query.Get("foo"), "an existing query parameter must survive")
			}
		})
	}
}

func TestConfig_Defaults(t *testing.T) {
	var zero Config

	assert.Equal(t, defaultHTTPTimeout, zero.httpClient().Timeout, "a nil HTTPClient must get a timeout")
	assert.WithinDuration(t, time.Now(), zero.now(), time.Minute, "a nil Now must fall back to the system clock")

	custom := &http.Client{Timeout: time.Second}
	assert.Same(t, custom, Config{HTTPClient: custom}.httpClient())

	fixed := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, fixed, Config{Now: func() time.Time { return fixed }}.now())
}

func TestConfig_OpenBrowserAtUsesTheInjectedOpener(t *testing.T) {
	var got string
	cfg := Config{OpenBrowser: func(rawURL string) error {
		got = rawURL
		return nil
	}}

	require.NoError(t, cfg.openBrowserAt("https://example.test/authorize"))
	assert.Equal(t, "https://example.test/authorize", got)
}
