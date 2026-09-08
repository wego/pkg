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
	"net"
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

	schemeHTTP  = "http"
	schemeHTTPS = "https"
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

	// NoBrowser suppresses the browser launch. Login reports the authorize URL
	// through PromptURL instead and then waits on the loopback listener exactly
	// as it otherwise would, for a headless shell, a terminal on a remote host,
	// or an operator who would rather open the URL themselves.
	//
	// The redirect still lands on CallbackAddr, so when the browser runs on a
	// different machine than the CLI that port has to be reachable from it —
	// usually `ssh -L <port>:localhost:<port>`. Suppressing the launch does not
	// move where the code is delivered.
	NoBrowser bool

	// PromptURL receives the authorize URL in place of a browser launch, so the
	// caller decides how to surface it: print it, render a QR code, hand it to
	// another process. Required when NoBrowser or ReadRedirect is set — a
	// sign-in whose URL the operator never sees cannot complete — and ignored
	// otherwise.
	PromptURL func(url string) error

	// ReadRedirect turns the sign-in into a paste-back exchange: instead of a
	// loopback listener receiving the redirect, the caller returns the URL the
	// browser was redirected to and Login reads the code out of it.
	//
	// This is the only path that works with NO browser on this machine AND no
	// way to reach its callback port. Cognito redirects to a loopback URL that
	// nothing is listening on, the browser shows a connection error, and the
	// address bar holds ?code=...&state=... for the operator to copy back.
	//
	// Cognito has no device authorization grant (RFC 8628): its discovery
	// document advertises no device_authorization_endpoint, /oauth2/device_
	// authorization is 404, and the token endpoint answers a device_code grant
	// with unsupported_grant_type. So this is the substitute, and it needs no
	// extra Cognito configuration because it is still the authorization code
	// grant with PKCE, only with the redirect carried by hand.
	//
	// Setting it implies NoBrowser: no browser is launched here. PromptURL is
	// required alongside it; CallbackAddr is not, since nothing binds a port. The pasted URL carries a single-use authorization
	// code, so it should not travel through a shared channel; state is still
	// checked, so a code from a different attempt is rejected.
	ReadRedirect func() (redirectedURL string, err error)

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

	code, callbackState, err := cfg.collectCode(ctx, state, generateChallenge(verifier))
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
	case parsed.ExpiresIn <= 0:
		// Silently accepting this is worse than failing. ExpiresAt would land on
		// exactly now(), and IsExpired subtracts a leeway on top, so a login that
		// just succeeded would read as already expired and the next command would
		// refresh or send the operator back through sign-in. Refusing here names
		// the real problem instead.
		return nil, errors.New("token response has no usable expires_in")
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
	return requireHTTPS("cognito token url", c.TokenURL)
}

// validateForLogin checks everything the interactive flow additionally needs.
func (c Config) validateForLogin() error {
	if err := c.validateClient(); err != nil {
		return err
	}
	if wegostrings.IsBlank(c.AuthorizeURL) {
		return errors.New("cognito authorize url is not configured")
	}
	if err := requireHTTPS("cognito authorize url", c.AuthorizeURL); err != nil {
		return err
	}
	if wegostrings.IsBlank(c.RedirectURI) {
		return errors.New("cognito redirect uri is not configured")
	}
	if err := requireLoopbackRedirect(c.RedirectURI); err != nil {
		return err
	}
	// Only the listener path needs a port; paste-back binds nothing.
	if c.ReadRedirect == nil && wegostrings.IsBlank(c.CallbackAddr) {
		return errors.New("local callback address is not configured")
	}
	if c.ReadRedirect == nil {
		if err := requireLoopbackAddr(c.CallbackAddr); err != nil {
			return err
		}
	}
	if c.ReadRedirect != nil && c.PromptURL == nil {
		return errors.New("paste-back sign-in needs Config.PromptURL: the operator has to be shown the url they are meant to open")
	}
	if c.NoBrowser && c.PromptURL == nil {
		return errors.New("no-browser sign-in needs Config.PromptURL: with no browser launched and no way to report the url, the operator has nothing to open")
	}
	return nil
}

// httpClient is the client to call the token endpoint with. It never follows
// redirects, whoever supplied it.
//
// The returned client is a COPY, so a caller-supplied HTTPClient is neither
// mutated nor able to reinstate redirect-following. That override is
// deliberate: the token request carries the authorization code, the PKCE
// verifier and the client id on sign-in and the refresh token on renewal, and
// Go's default policy follows up to ten redirects, re-sending the body
// verbatim on a 307 or 308. One redirect would therefore hand a complete
// credential set to whatever host the response named. It is also an injection
// route inwards, since the body that came back would be parsed as the session
// to use.
//
// A redirect from the token endpoint has no legitimate meaning here: TokenURL
// is a Cognito domain the operator configured, and Cognito answers it
// directly. Refusing turns the redirect into the error it should be.
func (c Config) httpClient() *http.Client {
	base := c.HTTPClient
	if base == nil {
		base = &http.Client{Timeout: defaultHTTPTimeout}
	}

	refusing := *base
	refusing.CheckRedirect = refuseRedirect

	return &refusing
}

// refuseRedirect stops the client at the redirect response instead of
// following it. Returning ErrUseLastResponse rather than an error of our own
// hands postToken the 3xx itself, which its status check then reports with the
// endpoint and status an operator needs to debug the misconfiguration.
func refuseRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

// now is the current time, from the injected clock when there is one.
func (c Config) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// openBrowserAt sends the operator to rawURL.
// presentAuthorizeURL gets the operator to the authorize URL, by launching a
// browser or, under NoBrowser, by handing the URL to PromptURL.
func (c Config) presentAuthorizeURL(rawURL string) error {
	// ReadRedirect implies the prompt. A machine that cannot receive the
	// callback generally cannot open a browser either, and where it could, the
	// listener path would have worked and been less work for the operator.
	// Launching a browser and then also asking for a paste is a footgun.
	if c.NoBrowser || c.ReadRedirect != nil {
		if err := c.PromptURL(rawURL); err != nil {
			return fmt.Errorf("present the sign-in url: %w", err)
		}

		return nil
	}

	if err := c.openBrowserAt(rawURL); err != nil {
		return fmt.Errorf("open browser for sign-in: %w", err)
	}

	return nil
}

// collectCode shows the operator the authorize URL and returns the code they
// came back with, by whichever route this Config selected.
//
// The loopback listener is the default. ReadRedirect replaces it with a
// paste-back exchange for a machine that can neither open a browser nor be
// reached on its callback port.
func (c Config) collectCode(ctx context.Context, state, challenge string) (code, gotState string, err error) {
	if c.ReadRedirect != nil {
		// No port is bound, so there is nothing to fail early on: show the URL,
		// then wait on the operator rather than on the network.
		if err := c.presentAuthorizeURL(c.buildAuthorizeURL(state, challenge)); err != nil {
			return "", "", err
		}

		pasted, err := c.readRedirect(ctx)
		if err != nil {
			return "", "", err
		}

		return codeFromRedirect(pasted)
	}

	// Bind the callback port BEFORE sending the operator to Cognito. If the
	// port is unavailable the redirect could never land, and there is no
	// fallback port to try, so failing here saves a pointless round trip.
	server, err := startCallbackServer(c.CallbackAddr, callbackPath(c.RedirectURI), state)
	if err != nil {
		return "", "", err
	}
	defer server.shutdown()

	if err := c.presentAuthorizeURL(c.buildAuthorizeURL(state, challenge)); err != nil {
		return "", "", err
	}

	return server.wait(ctx, defaultCallbackTimeout)
}

// codeFromRedirect reads the authorization code and state out of a redirect URL
// an operator pasted back.
//
// The host and scheme are deliberately not checked. The operator is copying
// from their own address bar, and what binds the code to this attempt is the
// state check the caller performs, not the shape of the URL. A bare code is
// refused for the same reason: with no state there is nothing to bind it to.
func codeFromRedirect(pasted string) (code, state string, err error) {
	trimmed := strings.TrimSpace(pasted)
	if wegostrings.IsBlank(trimmed) {
		return "", "", errors.New("nothing was pasted: copy the whole URL the browser was redirected to, including the ?code=... part")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", "", fmt.Errorf("parse the pasted redirect: %w", err)
	}

	if wegostrings.IsBlank(parsed.RawQuery) {
		return "", "", errors.New("the pasted value has no query string: copy the whole URL from the address bar, including the ?code=... part")
	}

	query := parsed.Query()

	// Cognito reports a refusal in the redirect rather than as a failed
	// request, so this is where a denied sign-in surfaces.
	if reported := query.Get("error"); wegostrings.IsNotBlank(reported) {
		if description := query.Get("error_description"); wegostrings.IsNotBlank(description) {
			return "", "", fmt.Errorf("the sign-in was refused: %s (%s)", reported, description)
		}

		return "", "", fmt.Errorf("the sign-in was refused: %s", reported)
	}

	code = query.Get("code")
	if wegostrings.IsBlank(code) {
		return "", "", errors.New("the pasted redirect carries no authorization code")
	}

	state = query.Get("state")
	if wegostrings.IsBlank(state) {
		return "", "", errors.New("the pasted redirect carries no state, so it cannot be tied to this sign-in attempt")
	}

	return code, state, nil
}

func (c Config) openBrowserAt(rawURL string) error {
	if c.OpenBrowser != nil {
		return c.OpenBrowser(rawURL)
	}
	return openBrowser(rawURL)
}

// ErrInsecureEndpoint is returned for a configured URL or bind address that
// must not carry OAuth credentials.
var ErrInsecureEndpoint = errors.New("insecure cognito endpoint")

// requireHTTPS rejects a Cognito endpoint that is not https.
//
// Both endpoints carry credentials in the clear if the transport does not
// protect them: the authorize URL puts the PKCE challenge and state on the
// wire, and the token endpoint receives the authorization code, the PKCE
// verifier and the client id on sign-in and the refresh token on renewal. A
// cleartext http:// value hands all of that to anyone on the path, so it is
// refused rather than warned about.
//
// There is no loopback exemption here, unlike the redirect: these are Cognito's
// own endpoints on a domain the operator configured, and Cognito serves them
// over https only. An http:// value is a misconfiguration in every case.
func requireHTTPS(name, raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: %s %q cannot be parsed: %w", ErrInsecureEndpoint, name, raw, err)
	}
	if parsed.Scheme != schemeHTTPS {
		return fmt.Errorf(
			"%w: %s must use https, got %q (the sign-in exchange carries the authorization code, the pkce verifier and the refresh token, so cleartext would expose the whole credential set)",
			ErrInsecureEndpoint, name, raw)
	}
	if wegostrings.IsBlank(parsed.Host) {
		return fmt.Errorf("%w: %s %q has no host", ErrInsecureEndpoint, name, raw)
	}
	if parsed.User != nil {
		return fmt.Errorf(
			"%w: %s %q embeds credentials in the url, which would be logged and sent on every request",
			ErrInsecureEndpoint, name, raw)
	}
	return nil
}

// requireLoopbackRedirect rejects a redirect URI that is not a loopback
// address.
//
// Cleartext http IS allowed here, and only here: RFC 8252 §7.3 has a native app
// receive the redirect on a loopback interface, where the request never leaves
// the machine and TLS would buy nothing but a certificate problem. What must
// hold is that the host is loopback — a redirect to a routable host would send
// the authorization code across the network to whoever answers it, which is the
// exact exposure this package's loopback-only design exists to prevent.
func requireLoopbackRedirect(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: cognito redirect uri %q cannot be parsed: %w", ErrInsecureEndpoint, raw, err)
	}
	switch parsed.Scheme {
	case schemeHTTP, schemeHTTPS:
	default:
		return fmt.Errorf(
			"%w: cognito redirect uri %q must be http or https, got scheme %q",
			ErrInsecureEndpoint, raw, parsed.Scheme)
	}
	if !isLoopbackHost(parsed.Hostname()) {
		return fmt.Errorf(
			"%w: cognito redirect uri %q must be a loopback address (127.0.0.1, [::1] or localhost) — a routable host would receive the authorization code over the network",
			ErrInsecureEndpoint, raw)
	}
	return nil
}

// requireLoopbackAddr rejects a callback bind address that is not loopback.
//
// Binding 0.0.0.0 or a LAN interface would let anything that can reach the
// machine deliver a redirect to the single-use callback, so the listener is
// held to the same loopback rule as the redirect it serves.
func requireLoopbackAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf(
			"%w: local callback address %q must be host:port, e.g. 127.0.0.1:8110: %w",
			ErrInsecureEndpoint, addr, err)
	}
	if wegostrings.IsBlank(port) {
		return fmt.Errorf("%w: local callback address %q has no port", ErrInsecureEndpoint, addr)
	}
	if !isLoopbackHost(host) {
		return fmt.Errorf(
			"%w: local callback address %q must bind a loopback interface (127.0.0.1 or [::1]) — binding 0.0.0.0 or a routable interface would let any host that can reach this machine deliver the sign-in callback",
			ErrInsecureEndpoint, addr)
	}
	return nil
}

// isLoopbackHost reports whether host is LITERALLY loopback.
//
// Only a literal loopback IP counts, plus the name "localhost". A name that
// merely resolves to 127.0.0.1 does not: resolution can change between this
// check and the request, so trusting it would make the check decorative. Go
// itself treats "localhost" as loopback (net/http.isLocalhost, and the URL
// spec's special-casing), and a poisoned "localhost" costs the attacker nothing
// they do not already have on a machine they can edit /etc/hosts on.
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

// readRedirect runs ReadRedirect without letting it pin Login to a blocked
// reader.
//
// ReadRedirect is synchronous and typically reads a line from stdin, which
// ignores context: os.Stdin.Read stays blocked until the operator types
// something or the descriptor closes. Called directly, that made a cancelled
// or timed-out Login unable to return — the caller asked to stop and the
// process sat there anyway, which is the whole defect.
//
// Running it on a goroutine and selecting on ctx.Done() means Login returns
// when the caller says so. It does NOT unblock the read itself, and nothing
// here can: the goroutine survives until the reader yields, then sends into a
// buffered channel and exits, so it parks rather than leaks and never blocks on
// a receiver that has gone. A caller that needs the read torn down too should
// close its own reader on cancellation — the signature stays
// func() (string, error) so existing callers keep working.
func (c Config) readRedirect(ctx context.Context) (string, error) {
	type pasteResult struct {
		url string
		err error
	}
	// Buffered: the goroutine must be able to finish and exit after Login has
	// already returned on ctx.Done() and stopped receiving.
	done := make(chan pasteResult, 1)

	go func() {
		url, err := c.ReadRedirect()
		done <- pasteResult{url: url, err: err}
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("waiting for the pasted redirect: %w", ctx.Err())
	case result := <-done:
		if result.err != nil {
			return "", fmt.Errorf("read the pasted redirect: %w", result.err)
		}
		return result.url, nil
	}
}
