package cognito_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wegostrings "github.com/wego/pkg/strings"

	"github.com/wego/pkg/cognito"
)

const (
	testAuthCode     = "the-authorization-code"
	testAccessValue  = "access-value"
	testIDTokenLabel = "id_token"
	testRefreshValue = "refresh-value"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		name            string
		givenConfig     func(*cognito.Config)
		givenBrowser    func(*fakeBrowser)
		givenHandler    func(*testing.T) http.HandlerFunc
		givenCancel     bool
		wantEmail       string
		wantErrContains string
	}{
		{
			name:            "blank client id tells the operator it is unprovisioned",
			givenConfig:     func(c *cognito.Config) { c.ClientID = "" },
			wantErrContains: "not provisioned",
		},
		{
			name:            "placeholder client id tells the operator it is unprovisioned",
			givenConfig:     func(c *cognito.Config) { c.ClientID = "REPLACE_WITH_COGNITO_CLIENT_ID" },
			wantErrContains: "not provisioned",
		},
		{
			name:            "missing authorize url is rejected",
			givenConfig:     func(c *cognito.Config) { c.AuthorizeURL = "" },
			wantErrContains: "authorize url",
		},
		{
			name:            "missing token url is rejected",
			givenConfig:     func(c *cognito.Config) { c.TokenURL = "" },
			wantErrContains: "token url",
		},
		{
			name:            "missing redirect uri is rejected",
			givenConfig:     func(c *cognito.Config) { c.RedirectURI = "" },
			wantErrContains: "redirect uri",
		},
		{
			name:            "missing callback address is rejected",
			givenConfig:     func(c *cognito.Config) { c.CallbackAddr = "" },
			wantErrContains: "callback address",
		},
		{
			name:            "browser launch failure is surfaced",
			givenBrowser:    func(f *fakeBrowser) { f.openErr = errors.New("no browser here") },
			wantErrContains: "open browser",
		},
		{
			name: "tampered state is rejected",
			givenBrowser: func(f *fakeBrowser) {
				f.tamper = func(q url.Values) { q.Set("state", "tampered-state") }
			},
			wantErrContains: "state mismatch",
		},
		{
			name: "missing state is rejected",
			givenBrowser: func(f *fakeBrowser) {
				f.tamper = func(q url.Values) { q.Del("state") }
			},
			wantErrContains: "state mismatch",
		},
		{
			name: "authorize server error is surfaced",
			givenBrowser: func(f *fakeBrowser) {
				f.tamper = func(q url.Values) {
					q.Del("code")
					q.Set("error", "access_denied")
					q.Set("error_description", "operator declined")
				}
			},
			wantErrContains: "access_denied",
		},
		{
			name: "missing code is rejected",
			givenBrowser: func(f *fakeBrowser) {
				f.tamper = func(q url.Values) { q.Del("code") }
			},
			wantErrContains: "no authorization code",
		},
		{
			name: "token endpoint rejection is surfaced",
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				return jsonHandler(http.StatusBadRequest, map[string]any{
					"error":             "invalid_grant",
					"error_description": "code already used",
				})
			},
			wantErrContains: "invalid_grant",
		},
		{
			name: "unparseable token response is rejected",
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				return rawHandler(http.StatusOK, "<html>not json</html>")
			},
			wantErrContains: "decode token response",
		},
		{
			name: "token response without an access token is rejected",
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				body := goodTokenBody(t, testOperatorEmail)
				delete(body, "access_token")
				return jsonHandler(http.StatusOK, body)
			},
			wantErrContains: "missing access_token",
		},
		{
			name: "token response without an id token is rejected",
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				body := goodTokenBody(t, testOperatorEmail)
				delete(body, testIDTokenLabel)
				return jsonHandler(http.StatusOK, body)
			},
			wantErrContains: "missing id_token",
		},
		{
			name: "token response without a refresh token is rejected",
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				body := goodTokenBody(t, testOperatorEmail)
				delete(body, "refresh_token")
				return jsonHandler(http.StatusOK, body)
			},
			wantErrContains: "missing refresh_token",
		},
		{
			name: "id token without an email claim is rejected",
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				body := goodTokenBody(t, testOperatorEmail)
				body[testIDTokenLabel] = mustIDToken(t, map[string]any{"sub": "abc"})
				return jsonHandler(http.StatusOK, body)
			},
			wantErrContains: "email",
		},
		{
			name: "email outside the allowed domain is rejected",
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				return jsonHandler(http.StatusOK, goodTokenBody(t, "someone@gmail.com"))
			},
			wantErrContains: "@wego.com",
		},
		{
			name:            "cancelled context stops waiting",
			givenBrowser:    func(f *fakeBrowser) { f.suppress = true },
			givenCancel:     true,
			wantErrContains: "context canceled",
		},
		{
			name:      "successful login returns a token set",
			wantEmail: testOperatorEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler http.HandlerFunc
			if tt.givenHandler != nil {
				handler = tt.givenHandler(t)
			}
			ts := newTokenServer(t, handler)

			browser := newFakeBrowser()
			if tt.givenBrowser != nil {
				tt.givenBrowser(browser)
			}

			cfg := baseConfig(t, ts.URL, mustFreeAddr(t))
			cfg.OpenBrowser = browser.open
			if tt.givenConfig != nil {
				tt.givenConfig(&cfg)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if tt.givenCancel {
				cancel()
			}

			got, err := cognito.Login(ctx, cfg)

			if wegostrings.IsNotEmpty(tt.wantErrContains) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)

			email, err := got.Email()
			require.NoError(t, err)
			assert.Equal(t, tt.wantEmail, email)
		})
	}
}

func TestLogin_BindsTheCallbackPortBeforeOpeningTheBrowser(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	addr := ln.Addr().String()

	ts := newTokenServer(t, nil)
	browser := newFakeBrowser()
	cfg := baseConfig(t, ts.URL, addr)
	cfg.OpenBrowser = browser.open

	got, err := cognito.Login(context.Background(), cfg)

	require.Error(t, err, "an occupied callback port must fail fast, not fall back to another port")
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), addr, "the error must name the address the operator has to free")
	assert.Empty(t, browser.capturedURL(), "the browser must not be launched when the callback port is unavailable")
}

func TestLogin_SendsPKCEAndState(t *testing.T) {
	ts := newTokenServer(t, nil)
	browser := newFakeBrowser()
	cfg := baseConfig(t, ts.URL, mustFreeAddr(t))
	cfg.OpenBrowser = browser.open

	got, err := cognito.Login(context.Background(), cfg)
	require.NoError(t, err)

	authorizeURL, err := url.Parse(browser.capturedURL())
	require.NoError(t, err)
	aq := authorizeURL.Query()

	assert.Equal(t, "https://cognito.test/oauth2/authorize", authorizeURL.Scheme+"://"+authorizeURL.Host+authorizeURL.Path)
	assert.Equal(t, "code", aq.Get("response_type"))
	assert.Equal(t, cfg.ClientID, aq.Get("client_id"))
	assert.Equal(t, cfg.RedirectURI, aq.Get("redirect_uri"))
	assert.Equal(t, cfg.Scopes, aq.Get("scope"))
	assert.Equal(t, "S256", aq.Get("code_challenge_method"), "plain PKCE must never be used")
	assert.NotEmpty(t, aq.Get("state"), "state is CSRF protection and must always be sent")

	form := ts.lastForm(t)
	assert.Equal(t, "authorization_code", form.Get("grant_type"))
	assert.Equal(t, cfg.ClientID, form.Get("client_id"))
	assert.Equal(t, testAuthCode, form.Get("code"))
	assert.Equal(t, cfg.RedirectURI, form.Get("redirect_uri"))

	verifier := form.Get("code_verifier")
	require.NotEmpty(t, verifier, "the exchange must carry the PKCE verifier")
	assert.GreaterOrEqual(t, len(verifier), 43)
	assert.LessOrEqual(t, len(verifier), 128)

	sum := sha256.Sum256([]byte(verifier))
	assert.Equal(t, base64.RawURLEncoding.EncodeToString(sum[:]), aq.Get("code_challenge"),
		"code_challenge must be the S256 hash of the verifier actually redeemed")

	assert.Equal(t, fixedNow.Add(time.Hour), got.ExpiresAt, "ExpiresAt must be derived from the injected clock")
	assert.Equal(t, testAccessValue, got.AccessToken)
	assert.Equal(t, testRefreshValue, got.RefreshToken)
}

func TestLogin_ErrorsDoNotLeakVerifierOrState(t *testing.T) {
	ts := newTokenServer(t, jsonHandler(http.StatusBadRequest, map[string]any{"error": "invalid_grant"}))
	browser := newFakeBrowser()
	cfg := baseConfig(t, ts.URL, mustFreeAddr(t))
	cfg.OpenBrowser = browser.open

	_, err := cognito.Login(context.Background(), cfg)
	require.Error(t, err)

	authorizeURL, parseErr := url.Parse(browser.capturedURL())
	require.NoError(t, parseErr)
	state := authorizeURL.Query().Get("state")
	require.NotEmpty(t, state)
	verifier := ts.lastForm(t).Get("code_verifier")
	require.NotEmpty(t, verifier)

	assert.NotContains(t, err.Error(), state, "state must never appear in an error message")
	assert.NotContains(t, err.Error(), verifier, "the PKCE verifier must never appear in an error message")
}

func TestLogin_FallsBackToTheSystemClockWhenNowIsNil(t *testing.T) {
	ts := newTokenServer(t, nil)
	browser := newFakeBrowser()
	cfg := baseConfig(t, ts.URL, mustFreeAddr(t))
	cfg.OpenBrowser = browser.open
	cfg.Now = nil

	before := time.Now()
	got, err := cognito.Login(context.Background(), cfg)
	require.NoError(t, err)

	assert.WithinDuration(t, before.Add(time.Hour), got.ExpiresAt, time.Minute)
}

func TestLogin_AllowsAnyEmailWhenNoDomainIsConfigured(t *testing.T) {
	ts := newTokenServer(t, jsonHandler(http.StatusOK, goodTokenBody(t, "contractor@example.test")))
	browser := newFakeBrowser()
	cfg := baseConfig(t, ts.URL, mustFreeAddr(t))
	cfg.OpenBrowser = browser.open
	cfg.AllowedDomain = ""

	got, err := cognito.Login(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, got)
}

func TestRefresh(t *testing.T) {
	tests := []struct {
		name            string
		givenConfig     func(*cognito.Config)
		givenRefresh    string
		givenHandler    func(*testing.T) http.HandlerFunc
		wantRefresh     string
		wantErrContains string
	}{
		{
			name:            "blank refresh token is rejected",
			givenRefresh:    "",
			wantErrContains: "refresh token",
		},
		{
			name:            "blank client id tells the operator it is unprovisioned",
			givenRefresh:    testRefreshValue,
			givenConfig:     func(c *cognito.Config) { c.ClientID = "" },
			wantErrContains: "not provisioned",
		},
		{
			name:            "placeholder client id tells the operator it is unprovisioned",
			givenRefresh:    testRefreshValue,
			givenConfig:     func(c *cognito.Config) { c.ClientID = "REPLACE_WITH_COGNITO_CLIENT_ID" },
			wantErrContains: "not provisioned",
		},
		{
			name:            "missing token url is rejected",
			givenRefresh:    testRefreshValue,
			givenConfig:     func(c *cognito.Config) { c.TokenURL = "" },
			wantErrContains: "token url",
		},
		{
			name:         "server failure is surfaced",
			givenRefresh: testRefreshValue,
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				return rawHandler(http.StatusInternalServerError, "upstream exploded")
			},
			wantErrContains: "500",
		},
		{
			name:         "unparseable response is rejected",
			givenRefresh: testRefreshValue,
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				return rawHandler(http.StatusOK, "not json")
			},
			wantErrContains: "decode token response",
		},
		{
			name:         "response without an access token is rejected",
			givenRefresh: testRefreshValue,
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				body := goodTokenBody(t, testOperatorEmail)
				delete(body, "access_token")
				return jsonHandler(http.StatusOK, body)
			},
			wantErrContains: "missing access_token",
		},
		{
			name:         "a rotated refresh token is adopted",
			givenRefresh: "original-refresh",
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				body := goodTokenBody(t, testOperatorEmail)
				body["refresh_token"] = "rotated-refresh"
				return jsonHandler(http.StatusOK, body)
			},
			wantRefresh: "rotated-refresh",
		},
		{
			name:         "an omitted refresh token falls back to the original",
			givenRefresh: "original-refresh",
			givenHandler: func(t *testing.T) http.HandlerFunc {
				t.Helper()
				body := goodTokenBody(t, testOperatorEmail)
				delete(body, "refresh_token")
				return jsonHandler(http.StatusOK, body)
			},
			wantRefresh: "original-refresh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler http.HandlerFunc
			if tt.givenHandler != nil {
				handler = tt.givenHandler(t)
			}
			ts := newTokenServer(t, handler)

			cfg := baseConfig(t, ts.URL, "127.0.0.1:0")
			if tt.givenConfig != nil {
				tt.givenConfig(&cfg)
			}

			got, err := cognito.Refresh(context.Background(), cfg, tt.givenRefresh)

			if wegostrings.IsNotEmpty(tt.wantErrContains) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantRefresh, got.RefreshToken)
			assert.Equal(t, "refresh_token", ts.lastForm(t).Get("grant_type"))
		})
	}
}

// TestRefresh_PreservesTheOriginalRefreshToken pins the load-bearing Cognito
// invariant: a refresh response omits refresh_token, and dropping the original
// would log the operator out on the very next command.
func TestRefresh_PreservesTheOriginalRefreshToken(t *testing.T) {
	body := goodTokenBody(t, testOperatorEmail)
	delete(body, "refresh_token")
	ts := newTokenServer(t, jsonHandler(http.StatusOK, body))

	cfg := baseConfig(t, ts.URL, "127.0.0.1:0")

	got, err := cognito.Refresh(context.Background(), cfg, "the-long-lived-refresh-token")
	require.NoError(t, err)
	assert.Equal(t, "the-long-lived-refresh-token", got.RefreshToken)
	assert.Equal(t, "the-long-lived-refresh-token", ts.lastForm(t).Get("refresh_token"),
		"the original refresh token must be what we present to Cognito")
	assert.Equal(t, fixedNow.Add(time.Hour), got.ExpiresAt)
}

// TestRefresh_DoesNotApplyTheDomainGate documents that the allowed-domain check
// is a login-time UX affordance, not something to re-run on every refresh.
func TestRefresh_DoesNotApplyTheDomainGate(t *testing.T) {
	ts := newTokenServer(t, jsonHandler(http.StatusOK, goodTokenBody(t, "someone@gmail.com")))
	cfg := baseConfig(t, ts.URL, "127.0.0.1:0")

	got, err := cognito.Refresh(context.Background(), cfg, testRefreshValue)
	require.NoError(t, err)
	require.NotNil(t, got)
}

func TestRefresh_HonoursContextCancellation(t *testing.T) {
	ts := newTokenServer(t, nil)
	cfg := baseConfig(t, ts.URL, "127.0.0.1:0")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got, err := cognito.Refresh(ctx, cfg, testRefreshValue)
	require.Error(t, err)
	assert.Nil(t, got)
}

// baseConfig is a working staging-shaped config aimed at a test token server.
func baseConfig(t *testing.T, tokenURL, callbackAddr string) cognito.Config {
	t.Helper()
	return cognito.Config{
		AuthorizeURL:  "https://cognito.test/oauth2/authorize",
		TokenURL:      tokenURL,
		ClientID:      "test-client-id",
		RedirectURI:   "http://" + callbackAddr + "/callback",
		Scopes:        "openid email profile",
		AllowedDomain: "@wego.com",
		CallbackAddr:  callbackAddr,
		Now:           func() time.Time { return fixedNow },
	}
}

// goodTokenBody is a well-formed Cognito token response.
func goodTokenBody(t *testing.T, email string) map[string]any {
	t.Helper()
	return map[string]any{
		"access_token":   testAccessValue,
		testIDTokenLabel: mustIDToken(t, map[string]any{"email": email}),
		"refresh_token":  testRefreshValue,
		"expires_in":     3600,
		"token_type":     "Bearer",
	}
}

func jsonHandler(status int, body any) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		enc, err := json.Marshal(body)
		if err != nil {
			http.Error(w, "marshal failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(enc)
	}
}

func rawHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

// tokenServer records every form it is posted so tests can assert on the
// exchange itself, not just its result.
type tokenServer struct {
	*httptest.Server

	mu    sync.Mutex
	forms []url.Values
}

func newTokenServer(t *testing.T, handler http.HandlerFunc) *tokenServer {
	t.Helper()

	ts := &tokenServer{}
	if handler == nil {
		handler = jsonHandler(http.StatusOK, goodTokenBody(t, testOperatorEmail))
	}

	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		ts.mu.Lock()
		ts.forms = append(ts.forms, r.PostForm)
		ts.mu.Unlock()
		handler(w, r)
	}))
	t.Cleanup(ts.Close)

	return ts
}

func (ts *tokenServer) lastForm(t *testing.T) url.Values {
	t.Helper()
	ts.mu.Lock()
	defer ts.mu.Unlock()
	require.NotEmpty(t, ts.forms, "the token endpoint was never called")
	return ts.forms[len(ts.forms)-1]
}

// fakeBrowser stands in for the operator's browser: it reads the authorize URL
// and drives the loopback callback the way a real redirect would.
type fakeBrowser struct {
	openErr  error
	suppress bool
	tamper   func(url.Values)

	mu           sync.Mutex
	authorizeURL string
}

func newFakeBrowser() *fakeBrowser {
	return &fakeBrowser{}
}

func (f *fakeBrowser) open(rawURL string) error {
	f.mu.Lock()
	f.authorizeURL = rawURL
	f.mu.Unlock()

	if f.openErr != nil {
		return f.openErr
	}
	if f.suppress {
		return nil
	}

	authorizeURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	callback, err := url.Parse(authorizeURL.Query().Get("redirect_uri"))
	if err != nil {
		return err
	}

	q := url.Values{}
	q.Set("code", testAuthCode)
	q.Set("state", authorizeURL.Query().Get("state"))
	if f.tamper != nil {
		f.tamper(q)
	}
	callback.RawQuery = q.Encode()

	go func() {
		req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, callback.String(), nil)
		if reqErr != nil {
			return
		}
		//nolint:gosec // G704: a loopback callback URL derived from the authorize URL this test built.
		resp, doErr := http.DefaultClient.Do(req)
		if doErr != nil {
			return
		}
		_ = resp.Body.Close()
	}()

	return nil
}

func (f *fakeBrowser) capturedURL() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.authorizeURL
}

// mustFreeAddr reserves and releases a loopback port, returning its address.
func mustFreeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())
	return addr
}

// TestLogin_NoBrowser covers the headless path: the caller suppresses the
// launch and surfaces the URL itself, and the sign-in still completes on the
// loopback listener. Suppressing the launch must not change anything else —
// the same PKCE and state parameters have to be sent, because a URL an
// operator pastes by hand is the same URL a browser would have been given.
func TestLogin_NoBrowser(t *testing.T) {
	ts := newTokenServer(t, nil)

	// The fake browser doubles as the prompt: it records the URL and drives the
	// callback, which is exactly what an operator pasting the URL would cause.
	prompt := newFakeBrowser()

	cfg := baseConfig(t, ts.URL, mustFreeAddr(t))
	cfg.NoBrowser = true
	cfg.PromptURL = prompt.open
	cfg.OpenBrowser = func(string) error {
		t.Error("NoBrowser must not launch a browser")

		return nil
	}

	got, err := cognito.Login(context.Background(), cfg)
	require.NoError(t, err)
	require.Equal(t, testAccessValue, got.AccessToken)

	authorizeURL, err := url.Parse(prompt.capturedURL())
	require.NoError(t, err, "PromptURL must receive a usable authorize url")

	aq := authorizeURL.Query()
	assert.Equal(t, cfg.ClientID, aq.Get("client_id"))
	assert.Equal(t, cfg.RedirectURI, aq.Get("redirect_uri"))
	assert.Equal(t, "S256", aq.Get("code_challenge_method"), "plain PKCE must never be used")
	assert.NotEmpty(t, aq.Get("code_challenge"))
	assert.NotEmpty(t, aq.Get("state"), "state is CSRF protection and must always be sent")
}

func TestLogin_NoBrowserRequiresPromptURL(t *testing.T) {
	cfg := baseConfig(t, "https://unused.test/oauth2/token", mustFreeAddr(t))
	cfg.NoBrowser = true
	cfg.PromptURL = nil

	_, err := cognito.Login(context.Background(), cfg)

	require.Error(t, err, "a login with no browser and no way to report the url cannot complete")
	assert.Contains(t, err.Error(), "PromptURL",
		"the error must name the field that is missing")
}

func TestLogin_NoBrowserPromptFailurePropagates(t *testing.T) {
	wantErr := errors.New("no tty to print to")

	cfg := baseConfig(t, "https://unused.test/oauth2/token", mustFreeAddr(t))
	cfg.NoBrowser = true
	cfg.PromptURL = func(string) error { return wantErr }

	_, err := cognito.Login(context.Background(), cfg)

	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr, "a prompt failure must reach the caller unwrapped in meaning")
	assert.NotContains(t, err.Error(), "open browser", "the browser path was not taken")
}

// TestLogin_RejectsAnUnusableExpiresIn pins that a token response with no
// usable expires_in fails loudly.
//
// Accepting it silently is worse than failing: ExpiresAt would land on exactly
// now(), IsExpired subtracts a leeway on top, and a login that had just
// succeeded would read as already expired — sending the operator back through
// sign-in on their very next command with no indication why.
func TestLogin_RejectsAnUnusableExpiresIn(t *testing.T) {
	tests := []struct {
		name            string
		givenExpiresIn  any
		wantErrContains string
	}{
		{
			name:            "expires_in omitted entirely",
			givenExpiresIn:  nil,
			wantErrContains: "expires_in",
		},
		{
			name:            "expires_in zero",
			givenExpiresIn:  0,
			wantErrContains: "expires_in",
		},
		{
			name:            "expires_in negative",
			givenExpiresIn:  -1,
			wantErrContains: "expires_in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := goodTokenBody(t, testOperatorEmail)
			delete(body, "expires_in")
			if tt.givenExpiresIn != nil {
				body["expires_in"] = tt.givenExpiresIn
			}

			ts := newTokenServer(t, jsonHandler(http.StatusOK, body))
			browser := newFakeBrowser()
			cfg := baseConfig(t, ts.URL, mustFreeAddr(t))
			cfg.OpenBrowser = browser.open

			_, err := cognito.Login(context.Background(), cfg)

			require.Error(t, err, "an unusable expires_in must not produce a session")
			assert.Contains(t, err.Error(), tt.wantErrContains)
		})
	}
}

// TestLogin_PasteBack drives a whole sign-in through the paste-back path: no
// browser is launched, no callback port is bound, and the code arrives as the
// redirect URL an operator copied out of their address bar.
func TestLogin_PasteBack(t *testing.T) {
	ts := newTokenServer(t, nil)

	var shown string

	cfg := cognito.Config{
		AuthorizeURL:  "https://cognito.test/oauth2/authorize",
		TokenURL:      ts.URL,
		ClientID:      "test-client-id",
		RedirectURI:   "http://localhost:8100/callback",
		Scopes:        "openid email",
		AllowedDomain: "@wego.com",
		Now:           func() time.Time { return fixedNow },
		PromptURL:     func(u string) error { shown = u; return nil },
	}
	// CallbackAddr is deliberately left empty: nothing binds a port here, and
	// that is the whole reason this path exists.
	cfg.ReadRedirect = func() (string, error) {
		return "http://localhost:8100/callback?code=pasted-code&state=" + stateOf(t, shown), nil
	}

	tokens, err := cognito.Login(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, tokens)

	assert.Contains(t, shown, "code_challenge_method=S256", "the operator must be shown a PKCE authorize url")

	form := ts.lastForm(t)
	assert.Equal(t, "pasted-code", form.Get("code"), "the pasted code must be what is redeemed")
	assert.NotEmpty(t, form.Get("code_verifier"), "PKCE must still bind the exchange to this process")
	assert.Equal(t, "authorization_code", form.Get("grant_type"))
}

// TestLogin_PasteBackRejectsAForeignState covers the guard that replaces the
// listener's origin check. Nothing about a pasted URL proves where it came
// from except the state, so a code from another attempt must be refused.
func TestLogin_PasteBackRejectsAForeignState(t *testing.T) {
	ts := newTokenServer(t, nil)

	cfg := cognito.Config{
		AuthorizeURL:  "https://cognito.test/oauth2/authorize",
		TokenURL:      ts.URL,
		ClientID:      "test-client-id",
		RedirectURI:   "http://localhost:8100/callback",
		AllowedDomain: "@wego.com",
		Now:           func() time.Time { return fixedNow },
		PromptURL:     func(string) error { return nil },
		ReadRedirect: func() (string, error) {
			return "http://localhost:8100/callback?code=pasted-code&state=someone-elses-state", nil
		},
	}

	_, err := cognito.Login(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "state mismatch")
	assert.Empty(t, ts.forms, "a code with a foreign state must never reach the token endpoint")
}

// TestLogin_PasteBackReportsAReadFailure covers the operator abandoning the
// prompt, e.g. ctrl-D at the paste.
func TestLogin_PasteBackReportsAReadFailure(t *testing.T) {
	cfg := cognito.Config{
		AuthorizeURL:  "https://cognito.test/oauth2/authorize",
		TokenURL:      "https://unused.test/oauth2/token",
		ClientID:      "test-client-id",
		RedirectURI:   "http://localhost:8100/callback",
		AllowedDomain: "@wego.com",
		PromptURL:     func(string) error { return nil },
		ReadRedirect:  func() (string, error) { return "", errors.New("stdin closed") },
	}

	_, err := cognito.Login(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stdin closed")
}

// stateOf pulls the state parameter out of an authorize URL.
func stateOf(t *testing.T, authorizeURL string) string {
	t.Helper()

	parsed, err := url.Parse(authorizeURL)
	require.NoError(t, err)

	state := parsed.Query().Get("state")
	require.NotEmpty(t, state, "the authorize url must carry a state")

	return state
}
