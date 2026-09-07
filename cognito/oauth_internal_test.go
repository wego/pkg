package cognito

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
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

	// An injected client is honoured for its transport and timeout but is
	// COPIED, never handed back: httpClient has to own CheckRedirect, and it
	// must not mutate a client the caller may use elsewhere. See
	// TestPostToken_RefusesRedirects for why the override exists.
	transport := &http.Transport{}
	custom := &http.Client{Timeout: time.Second, Transport: transport}
	got := Config{HTTPClient: custom}.httpClient()

	assert.NotSame(t, custom, got, "the caller's client must not be handed back")
	assert.Equal(t, time.Second, got.Timeout, "the caller's timeout must survive")
	assert.Same(t, transport, got.Transport, "the caller's transport must survive")
	assert.NotNil(t, got.CheckRedirect, "the returned client must refuse redirects")
	assert.Nil(t, custom.CheckRedirect, "the caller's client must not be mutated")

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

// TestPostToken_RefusesRedirects proves a token request is never re-sent to a
// host other than the one configured.
//
// The form carries the authorization code, the PKCE verifier and the client id
// on sign-in, and the refresh token on renewal. Go's default client follows up
// to ten redirects and re-sends the body verbatim on a 307 or 308, so
// following one would hand a complete credential set to whatever host the
// response named. A redirect is also an injection route in the other
// direction: the body that comes back would be parsed as a token set, letting
// an unintended host choose the session the CLI then uses.
func TestPostToken_RefusesRedirects(t *testing.T) {
	tests := []struct {
		name        string
		givenStatus int
		givenClient *http.Client
	}{
		{
			name:        "the default client refuses a 307, which would re-send the body",
			givenStatus: http.StatusTemporaryRedirect,
		},
		{
			name:        "the default client refuses a 302",
			givenStatus: http.StatusFound,
		},
		{
			// A caller supplying its own client must not be able to reinstate
			// following, deliberately or by copying a client from elsewhere.
			name:        "an injected permissive client cannot opt back into following",
			givenStatus: http.StatusTemporaryRedirect,
			givenClient: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return nil }},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var elsewhereHits atomic.Int32

			// Answers with a usable token set, so following the redirect would
			// look like a successful sign-in rather than an error.
			elsewhere := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				elsewhereHits.Add(1)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"access_token": "attacker-access",
					"id_token":     "attacker-id",
					"expires_in":   3600,
				})
			}))
			t.Cleanup(elsewhere.Close)

			origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, elsewhere.URL, tt.givenStatus)
			}))
			t.Cleanup(origin.Close)

			cfg := Config{TokenURL: origin.URL, HTTPClient: tt.givenClient}
			form := url.Values{"code": {"secret-code"}, "code_verifier": {"secret-verifier"}}

			tokens, err := cfg.postToken(context.Background(), form, "")

			require.Error(t, err, "a redirected token endpoint must not produce a session")
			assert.Nil(t, tokens)
			assert.Zero(t, elsewhereHits.Load(), "credentials must never reach the redirect target")
		})
	}
}

// TestCodeFromRedirect covers pulling the authorization code out of the URL an
// operator pasted back.
//
// This is the sign-in path for a machine with no browser AND no way to forward
// the callback port: Cognito redirects to a loopback URL nothing is listening
// on, the browser shows a connection error, and the address bar holds the code.
func TestCodeFromRedirect(t *testing.T) {
	tests := []struct {
		name       string
		givenInput string
		wantCode   string
		wantState  string
		wantErr    string
	}{
		{
			name:       "a pasted redirect yields its code and state",
			givenInput: "http://localhost:8100/callback?code=abc123&state=xyz789",
			wantCode:   "abc123",
			wantState:  "xyz789",
		},
		{
			name:       "surrounding whitespace from a copy is tolerated",
			givenInput: "  http://localhost:8100/callback?code=abc123&state=xyz789\n",
			wantCode:   "abc123",
			wantState:  "xyz789",
		},
		{
			name:       "https and an unexpected host are accepted; the state check is the guard",
			givenInput: "https://example.test/cb?state=xyz789&code=abc123",
			wantCode:   "abc123",
			wantState:  "xyz789",
		},
		{
			name:       "percent-encoded values are decoded",
			givenInput: "http://localhost:8100/callback?code=a%2Fb%2Bc&state=s%3D1",
			wantCode:   "a/b+c",
			wantState:  "s=1",
		},
		{name: "nothing pasted", givenInput: "   ", wantErr: "nothing was pasted"},
		{
			// The commonest mistake: copying the path but not the query.
			name:       "a url with no query says what is missing",
			givenInput: "http://localhost:8100/callback",
			wantErr:    "no query",
		},
		{
			name:       "a bare code is refused because it carries no state",
			givenInput: "abc123",
			wantErr:    "no query",
		},
		{
			name:       "a denied sign-in reports the provider's error",
			givenInput: "http://localhost:8100/callback?error=access_denied&error_description=User+denied",
			wantErr:    "access_denied",
		},
		{
			name:       "a denied sign-in includes the description when there is one",
			givenInput: "http://localhost:8100/callback?error=access_denied&error_description=User+denied",
			wantErr:    "User denied",
		},
		{
			name:       "a query with state but no code is refused",
			givenInput: "http://localhost:8100/callback?state=xyz789",
			wantErr:    "no authorization code",
		},
		{
			// Without state there is nothing to bind the code to this attempt.
			name:       "a query with code but no state is refused",
			givenInput: "http://localhost:8100/callback?code=abc123",
			wantErr:    "no state",
		},
		{
			name:       "an unparsable url is reported",
			givenInput: "http://[::1",
			wantErr:    "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, state, err := codeFromRedirect(tt.givenInput)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Empty(t, code)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, code)
			assert.Equal(t, tt.wantState, state)
		})
	}
}

// TestValidateForLogin_PasteBack covers the config rules the paste-back path
// adds: it needs somewhere to show the URL, and it does NOT need a callback
// port, since nothing binds one.
func TestValidateForLogin_PasteBack(t *testing.T) {
	base := func() Config {
		return Config{
			AuthorizeURL: "https://cognito.test/oauth2/authorize",
			TokenURL:     "https://cognito.test/oauth2/token",
			ClientID:     "client",
			RedirectURI:  "http://localhost:8100/callback",
		}
	}

	t.Run("paste-back without PromptURL is refused", func(t *testing.T) {
		cfg := base()
		cfg.ReadRedirect = func() (string, error) { return "", nil }

		require.ErrorContains(t, cfg.validateForLogin(), "PromptURL")
	})

	t.Run("paste-back needs no callback address", func(t *testing.T) {
		cfg := base()
		cfg.ReadRedirect = func() (string, error) { return "", nil }
		cfg.PromptURL = func(string) error { return nil }

		require.NoError(t, cfg.validateForLogin(), "nothing binds a port on this path")
	})

	t.Run("the listener path still needs a callback address", func(t *testing.T) {
		require.ErrorContains(t, base().validateForLogin(), "callback address")
	})
}

// TestPresentAuthorizeURL_PasteBackImpliesThePrompt pins the coupling: setting
// ReadRedirect must not leave a browser being launched, or the operator gets a
// browser AND a paste prompt for the same sign-in.
func TestPresentAuthorizeURL_PasteBackImpliesThePrompt(t *testing.T) {
	var prompted, opened string

	cfg := Config{
		PromptURL:    func(u string) error { prompted = u; return nil },
		OpenBrowser:  func(u string) error { opened = u; return nil },
		ReadRedirect: func() (string, error) { return "", nil },
	}

	require.NoError(t, cfg.presentAuthorizeURL("https://cognito.test/authorize?state=s"))

	assert.Equal(t, "https://cognito.test/authorize?state=s", prompted)
	assert.Empty(t, opened, "no browser may be launched on the paste-back path")
}
