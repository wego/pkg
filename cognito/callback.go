package cognito

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	wegostrings "github.com/wego/pkg/strings"
)

const (
	// defaultCallbackTimeout bounds how long we hold the loopback listener
	// open waiting for the operator to finish signing in.
	defaultCallbackTimeout = 5 * time.Minute

	callbackReadHeaderTimeout = 5 * time.Second
	callbackShutdownTimeout   = 2 * time.Second
)

const successHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>Signed in</title>
<style>body{font-family:-apple-system,BlinkMacSystemFont,sans-serif;max-width:480px;margin:10vh auto;padding:0 24px;color:#222}
h1{font-size:20px}p{color:#555;line-height:1.5}</style></head>
<body><h1>Signed in</h1>
<p>You can close this window and return to your terminal.</p></body></html>`

// failureHTML deliberately carries no detail from the request. The operator
// reads the actual reason in their terminal, where it cannot be reflected back
// into a page, so nothing from the redirect is ever interpolated into HTML.
const failureHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>Sign-in failed</title>
<style>body{font-family:-apple-system,BlinkMacSystemFont,sans-serif;max-width:480px;margin:10vh auto;padding:0 24px;color:#222}
h1{font-size:20px;color:#b00}p{color:#555;line-height:1.5}</style></head>
<body><h1>Sign-in failed</h1>
<p>Return to your terminal for the reason, and try again.</p></body></html>`

// callbackResult is the single outcome a callback server reports.
type callbackResult struct {
	code  string
	state string
	err   error
}

// callbackServer is a single-use loopback listener for the OAuth redirect.
type callbackServer struct {
	results <-chan callbackResult
	srv     *http.Server
}

// callbackPath is the request path the redirect URI points at, falling back to
// root when the URI carries no path (or cannot be parsed at all).
func callbackPath(redirectURI string) string {
	parsed, err := url.Parse(redirectURI)
	if err != nil || wegostrings.IsBlank(parsed.Path) {
		return "/"
	}
	return parsed.Path
}

// startCallbackServer binds addr and serves the OAuth redirect at path.
//
// A bind failure is terminal on purpose: the Cognito app client registers
// exactly one callback URL, so listening somewhere else would produce a
// redirect the authorization server refuses. Fail loudly instead.
func startCallbackServer(addr, path string) (*callbackServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf(
			"bind the sign-in callback listener on %s: %w (that address must be free - the Cognito app client registers exactly one callback URL, so another port is not an option; close whatever holds it and retry)",
			addr, err)
	}

	results := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc(path, callbackHandler(results))

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: callbackReadHeaderTimeout,
	}
	go func() {
		_ = srv.Serve(listener)
	}()

	return &callbackServer{results: results, srv: srv}, nil
}

// callbackHandler serves the redirect, renders a minimal page for the
// operator, and reports the outcome exactly once.
func callbackHandler(results chan<- callbackResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if authErr := query.Get("error"); wegostrings.IsNotBlank(authErr) {
			message := authErr
			if desc := query.Get("error_description"); wegostrings.IsNotBlank(desc) {
				message += ": " + desc
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, failureHTML)
			deliver(results, callbackResult{err: fmt.Errorf("authorization server rejected the sign-in: %s", message)})
			return
		}

		code := query.Get("code")
		if wegostrings.IsBlank(code) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, failureHTML)
			deliver(results, callbackResult{err: errors.New("the sign-in redirect carried no authorization code")})
			return
		}

		fmt.Fprint(w, successHTML)
		deliver(results, callbackResult{code: code, state: query.Get("state")})
	}
}

// deliver reports the first result and silently drops later ones. The flow is
// single-use: a reloaded browser tab must neither block the handler nor
// overwrite the outcome we already acted on.
func deliver(results chan<- callbackResult, result callbackResult) {
	select {
	case results <- result:
	default:
	}
}

// wait blocks for the callback, honouring both ctx cancellation and timeout.
func (c *callbackServer) wait(ctx context.Context, timeout time.Duration) (code, state string, err error) {
	if timeout <= 0 {
		timeout = defaultCallbackTimeout
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return "", "", fmt.Errorf("waiting for the sign-in callback: %w", ctx.Err())
	case <-timer.C:
		return "", "", fmt.Errorf("timed out after %s waiting for the sign-in callback", timeout)
	case result := <-c.results:
		if result.err != nil {
			return "", "", result.err
		}
		return result.code, result.state, nil
	}
}

// shutdown releases the loopback listener.
func (c *callbackServer) shutdown() {
	if c.srv == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), callbackShutdownTimeout)
	defer cancel()
	_ = c.srv.Shutdown(ctx)
}
