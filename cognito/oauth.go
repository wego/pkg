// Package cognito implements the Cognito authorization-code-with-PKCE login
// flow a command-line tool uses to obtain a human operator's tokens.
//
// It is built for CLIs rather than services: the flow opens a browser, receives
// the authorization code on a loopback listener, and hands back the token set
// for the caller to cache. A service verifying an incoming token wants
// github.com/wego/pkg/http/jwt instead.
//
// Two deliberate constraints shape this package:
//
//   - It depends on nothing outside the standard library (bar Wego's string
//     helpers). Hand-rolling the OAuth exchange keeps every wire parameter
//     visible and auditable, which matters more here than the convenience an
//     OAuth library would buy.
//   - It holds no package-level mutable state. Every dependency - the clock,
//     the HTTP client, the browser opener - arrives through Config, so tests
//     and callers never race over shared globals.
package cognito

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	wegostrings "github.com/wego/pkg/strings"
)

const (
	// placeholderPrefix marks a Cognito value that ops has not filled in yet.
	placeholderPrefix = "REPLACE_WITH_"

	defaultHTTPTimeout    = 30 * time.Second
	maxTokenResponseBytes = 1 << 20
)

// Config carries everything the login flow needs. Nothing is read from the
// environment or from package state, so a caller holds the whole contract.
type Config struct {
	// AuthorizeURL is the Cognito hosted-UI authorize endpoint.
	AuthorizeURL string

	// TokenURL is the Cognito token endpoint.
	TokenURL string

	// ClientID is the Cognito app client id.
	ClientID string

	// RedirectURI is the callback URL registered on the app client. Its path
	// determines where the local callback server listens.
	RedirectURI string

	// Scopes is the space-separated scope list to request.
	Scopes string

	// AllowedDomain, when set, is the email suffix an operator must sign in
	// with, e.g. "@wego.com".
	AllowedDomain string

	// CallbackAddr is the host:port the local callback server binds, e.g.
	// "127.0.0.1:8110". It must agree with RedirectURI.
	CallbackAddr string

	// IdentityProvider, when set, is forwarded as identity_provider so Cognito
	// jumps straight to that IdP instead of showing its own chooser.
	IdentityProvider string

	// OpenBrowser launches the authorize URL. Nil uses the OS default opener.
	OpenBrowser func(string) error

	// HTTPClient calls the token endpoint. Nil uses a client with a timeout.
	HTTPClient *http.Client

	// Now supplies the current time, for computing token expiry. Nil uses
	// time.Now; tests inject a fixed clock.
	Now func() time.Time
}

// Login runs the browser-based authorization-code-with-PKCE flow and returns
// the operator's tokens.
func Login(ctx context.Context, cfg Config) (*TokenSet, error) {
	if err := cfg.validateForLogin(); err != nil {
		return nil, err
	}

	verifier, err := generateVerifier()
	if err != nil {
		return nil, err
	}
	state, err := generateState()
	if err != nil {
		return nil, err
	}

	// Bind the callback port BEFORE sending the operator to Cognito. If the
	// port is unavailable the redirect could never land, and there is no
	// fallback port to try, so failing here saves a pointless round trip.
	server, err := startCallbackServer(cfg.CallbackAddr, callbackPath(cfg.RedirectURI))
	if err != nil {
		return nil, err
	}
	defer server.shutdown()

	if err := cfg.openBrowserAt(cfg.buildAuthorizeURL(state, generateChallenge(verifier))); err != nil {
		return nil, fmt.Errorf("open browser for sign-in: %w", err)
	}

	code, callbackState, err := server.wait(ctx, defaultCallbackTimeout)
	if err != nil {
		return nil, err
	}

	// CSRF protection, not decoration: a callback whose state is not the value
	// we minted did not come from the authorize request we started, so the
	// code it carries is not ours to redeem.
	if !stateMatches(state, callbackState) {
		return nil, errors.New("oauth state mismatch: the sign-in callback did not come from this login attempt")
	}

	tokens, err := cfg.exchangeCode(ctx, code, verifier)
	if err != nil {
		return nil, err
	}

	if err := cfg.checkAllowedDomain(tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

// Refresh exchanges a refresh token for a fresh access and id token.
func Refresh(ctx context.Context, cfg Config, refreshToken string) (*TokenSet, error) {
	if wegostrings.IsBlank(refreshToken) {
		return nil, errors.New("no refresh token available: sign in again")
	}
	if err := cfg.validateClient(); err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", cfg.ClientID)
	form.Set("refresh_token", refreshToken)

	return cfg.postToken(ctx, form, refreshToken)
}

// exchangeCode redeems an authorization code together with its PKCE verifier.
func (c Config) exchangeCode(ctx context.Context, code, verifier string) (*TokenSet, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", c.ClientID)
	form.Set("code", code)
	form.Set("redirect_uri", c.RedirectURI)
	form.Set("code_verifier", verifier)

	return c.postToken(ctx, form, "")
}

// tokenResponse is the wire shape of a Cognito token response. The field names
// are fixed by the OAuth spec, so gosec's secret-field warning is expected.
type tokenResponse struct {
	AccessToken  string `json:"access_token"` //nolint:gosec // G117: OAuth wire field; decoded in-process and never logged.
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"` //nolint:gosec // G117: OAuth wire field, as above.
	ExpiresIn    int    `json:"expires_in"`
}

// postToken performs a token-endpoint call. fallbackRefresh is the refresh
// token to keep if the response omits one.
func (c Config) postToken(ctx context.Context, form url.Values, fallbackRefresh string) (*TokenSet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	//nolint:gosec // G704: TokenURL is operator-controlled CLI config (the Cognito domain), not request input.
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("call the token endpoint: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxTokenResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, tokenEndpointError(resp.StatusCode, body)
	}

	var parsed tokenResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	// Cognito omits refresh_token from a refresh response, so the original has
	// to be carried forward. This is load-bearing: dropping it would silently
	// sign the operator out on their next command. Do not add validation that
	// rejects this fallback.
	refresh := parsed.RefreshToken
	if wegostrings.IsBlank(refresh) {
		refresh = fallbackRefresh
	}

	switch {
	case wegostrings.IsBlank(parsed.AccessToken):
		return nil, errors.New("token response is missing access_token")
	case wegostrings.IsBlank(parsed.IDToken):
		return nil, errors.New("token response is missing id_token")
	case wegostrings.IsBlank(refresh):
		return nil, errors.New("token response is missing refresh_token")
	}

	return &TokenSet{
		AccessToken:  parsed.AccessToken,
		IDToken:      parsed.IDToken,
		RefreshToken: refresh,
		ExpiresAt:    c.now().Add(time.Duration(parsed.ExpiresIn) * time.Second),
	}, nil
}

// tokenEndpointError reports a non-2xx token response.
//
// Only the standard OAuth error fields are echoed. An arbitrary response body
// is deliberately withheld, so no token can ever ride out inside an error
// message that ends up in a log or a bug report.
func tokenEndpointError(status int, body []byte) error {
	var parsed struct {
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}

	if err := json.Unmarshal(body, &parsed); err == nil && wegostrings.IsNotBlank(parsed.Error) {
		if wegostrings.IsNotBlank(parsed.Description) {
			return fmt.Errorf("token endpoint returned %d: %s: %s", status, parsed.Error, parsed.Description)
		}
		return fmt.Errorf("token endpoint returned %d: %s", status, parsed.Error)
	}

	return fmt.Errorf("token endpoint returned %d (response body withheld: it may carry credentials)", status)
}

// buildAuthorizeURL assembles the hosted-UI URL the operator is sent to.
func (c Config) buildAuthorizeURL(state, challenge string) string {
	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", c.ClientID)
	query.Set("redirect_uri", c.RedirectURI)
	query.Set("scope", c.Scopes)
	query.Set("state", state)
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	if wegostrings.IsNotBlank(c.IdentityProvider) {
		query.Set("identity_provider", c.IdentityProvider)
	}

	separator := "?"
	if strings.Contains(c.AuthorizeURL, "?") {
		separator = "&"
	}
	return c.AuthorizeURL + separator + query.Encode()
}

// checkAllowedDomain rejects an operator signed in outside AllowedDomain.
//
// This is client-side UX, NOT a security boundary: it catches "you signed in
// with your personal Google account" before the CLI starts issuing calls that
// would fail confusingly. The server does not enforce it, so nothing may rely
// on it for authorization.
func (c Config) checkAllowedDomain(tokens *TokenSet) error {
	if wegostrings.IsBlank(c.AllowedDomain) {
		return nil
	}

	email, err := tokens.Email()
	if err != nil {
		return fmt.Errorf("check the signed-in email: %w", err)
	}

	if !strings.HasSuffix(strings.ToLower(email), strings.ToLower(c.AllowedDomain)) {
		return fmt.Errorf("signed in as %s, but a %s account is required", email, c.AllowedDomain)
	}
	return nil
}

// validateClient checks the values every token call needs.
func (c Config) validateClient() error {
	if wegostrings.IsBlank(c.ClientID) || strings.HasPrefix(c.ClientID, placeholderPrefix) {
		return errors.New("cognito app client is not provisioned yet: Config.ClientID is blank or still a placeholder, so the app client needs to be created and its id wired into the caller's config")
	}
	if wegostrings.IsBlank(c.TokenURL) {
		return errors.New("cognito token url is not configured")
	}
	return nil
}

// validateForLogin checks everything the interactive flow additionally needs.
func (c Config) validateForLogin() error {
	if err := c.validateClient(); err != nil {
		return err
	}
	if wegostrings.IsBlank(c.AuthorizeURL) {
		return errors.New("cognito authorize url is not configured")
	}
	if wegostrings.IsBlank(c.RedirectURI) {
		return errors.New("cognito redirect uri is not configured")
	}
	if wegostrings.IsBlank(c.CallbackAddr) {
		return errors.New("local callback address is not configured")
	}
	return nil
}

// httpClient is the client to call the token endpoint with.
func (c Config) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: defaultHTTPTimeout}
}

// now is the current time, from the injected clock when there is one.
func (c Config) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// openBrowserAt sends the operator to rawURL.
func (c Config) openBrowserAt(rawURL string) error {
	if c.OpenBrowser != nil {
		return c.OpenBrowser(rawURL)
	}
	return openBrowser(rawURL)
}
