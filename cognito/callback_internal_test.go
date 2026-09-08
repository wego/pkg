package cognito

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wegostrings "github.com/wego/pkg/strings"
)

func TestCallbackPath(t *testing.T) {
	tests := []struct {
		name     string
		givenURI string
		want     string
	}{
		{name: "unparseable falls back to root", givenURI: "http://[::1", want: "/"},
		{name: "no path falls back to root", givenURI: "http://localhost:8110", want: "/"},
		{name: "blank falls back to root", givenURI: "", want: "/"},
		{name: "explicit path is used", givenURI: "http://localhost:8110/callback", want: "/callback"},
		{name: "nested path is used", givenURI: "http://127.0.0.1:8110/oauth/cb", want: "/oauth/cb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, callbackPath(tt.givenURI))
		})
	}
}

func TestCallbackHandler(t *testing.T) {
	const wantedState = "the-state"

	tests := []struct {
		name            string
		givenQuery      string
		wantStatus      int
		wantCode        string
		wantState       string
		wantErrContains string
		wantNoDelivery  bool
	}{
		{
			name:            "authorize error is surfaced with its description",
			givenQuery:      "error=access_denied&error_description=user+said+no&state=" + wantedState,
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "user said no",
		},
		{
			name:            "authorize error without a description still reports",
			givenQuery:      "error=server_error&state=" + wantedState,
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "server_error",
		},
		{
			name:            "missing code is rejected",
			givenQuery:      "state=" + wantedState,
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "no authorization code",
		},
		{
			name:       "code and state are captured",
			givenQuery: "code=the-code&state=" + wantedState,
			wantStatus: http.StatusOK,
			wantCode:   "the-code",
			wantState:  wantedState,
		},
		// The next three are the denial-of-sign-in guard. Each is a request
		// that reaches the predictable callback port without belonging to this
		// attempt, and the assertion that matters is wantNoDelivery: the
		// single-use channel must still be empty afterwards, so the genuine
		// redirect can still be accepted.
		{
			name:           "a code carrying someone else's state is not delivered",
			givenQuery:     "code=blind-code&state=not-this-attempt",
			wantStatus:     http.StatusBadRequest,
			wantNoDelivery: true,
		},
		{
			name:           "a code with no state at all is not delivered",
			givenQuery:     "code=blind-code",
			wantStatus:     http.StatusBadRequest,
			wantNoDelivery: true,
		},
		{
			name:           "an authorize error from another attempt is not delivered",
			givenQuery:     "error=access_denied&state=not-this-attempt",
			wantStatus:     http.StatusBadRequest,
			wantNoDelivery: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan callbackResult, 1)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/callback?"+tt.givenQuery, nil)

			callbackHandler(wantedState, ch)(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")

			if tt.wantNoDelivery {
				assert.Empty(t, ch, "an uncorrelated request must not consume the single-use delivery")
				assert.NotContains(t, rec.Body.String(), "blind-code",
					"the page must not reflect anything from the request")
				return
			}

			select {
			case got := <-ch:
				if wegostrings.IsNotEmpty(tt.wantErrContains) {
					require.Error(t, got.err)
					assert.Contains(t, got.err.Error(), tt.wantErrContains)
					return
				}
				require.NoError(t, got.err)
				assert.Equal(t, tt.wantCode, got.code)
				assert.Equal(t, tt.wantState, got.state)
			default:
				t.Fatal("handler did not emit a callbackResult")
			}
		})
	}
}

// TestCallbackHandler_BlindRequestDoesNotAbortTheRealSignIn is the
// invalid-then-valid case. Before the state check moved ahead of delivery, the
// first request here consumed the one delivery the flow gets and the operator's
// actual redirect arrived to a channel that was already full, so a sign-in
// could be aborted by anything able to guess the callback port.
func TestCallbackHandler_BlindRequestDoesNotAbortTheRealSignIn(t *testing.T) {
	const wantedState = "real-attempt"

	ch := make(chan callbackResult, 1)
	handler := callbackHandler(wantedState, ch)

	blind := httptest.NewRecorder()
	handler(blind, httptest.NewRequest(http.MethodGet, "/callback?code=blind&state=guessed", nil))

	assert.Equal(t, http.StatusBadRequest, blind.Code)
	require.Empty(t, ch, "the blind request must leave the delivery unused")

	real := httptest.NewRecorder()
	handler(real, httptest.NewRequest(http.MethodGet, "/callback?code=genuine&state="+wantedState, nil))

	assert.Equal(t, http.StatusOK, real.Code)
	select {
	case got := <-ch:
		require.NoError(t, got.err)
		assert.Equal(t, "genuine", got.code, "the real callback must still be the one delivered")
		assert.Equal(t, wantedState, got.state)
	default:
		t.Fatal("the genuine callback was not delivered after a blind request")
	}
}

func TestCallbackHandler_IsSingleUse(t *testing.T) {
	ch := make(chan callbackResult, 1)
	handler := callbackHandler("s", ch)

	for range 3 {
		rec := httptest.NewRecorder()
		handler(rec, httptest.NewRequest(http.MethodGet, "/callback?code=c&state=s", nil))
	}

	require.Len(t, ch, 1, "a redelivered callback must neither block nor enqueue a second result")
}

func TestStartCallbackServer_PortAlreadyBoundFailsFast(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	addr := ln.Addr().String()
	cs, err := startCallbackServer(addr, "/callback", "s")

	require.Error(t, err, "binding an occupied port must fail rather than silently pick another")
	assert.Nil(t, cs)
	assert.Contains(t, err.Error(), addr, "the error must name the address the operator has to free")
}

func TestStartCallbackServer_ServesTheCallback(t *testing.T) {
	addr := mustFreeAddr(t)
	cs, err := startCallbackServer(addr, "/callback", "xyz")
	require.NoError(t, err)
	t.Cleanup(cs.shutdown)

	mustGet(t, "http://"+addr+"/callback?code=abc&state=xyz")

	code, state, err := cs.wait(context.Background(), 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "abc", code)
	assert.Equal(t, "xyz", state)
}

func TestCallbackServer_Wait(t *testing.T) {
	tests := []struct {
		name            string
		givenResult     *callbackResult
		givenTimeout    time.Duration
		givenCancel     bool
		wantCode        string
		wantErrContains string
	}{
		{
			name:            "timeout is reported",
			givenTimeout:    10 * time.Millisecond,
			wantErrContains: "timed out",
		},
		{
			name:            "cancellation is reported",
			givenTimeout:    time.Minute,
			givenCancel:     true,
			wantErrContains: "context canceled",
		},
		{
			name:            "callback error is propagated",
			givenResult:     &callbackResult{err: stubError("authorize: access_denied")},
			givenTimeout:    time.Minute,
			wantErrContains: "access_denied",
		},
		{
			name:         "code and state are returned",
			givenResult:  &callbackResult{code: "c", state: "s"},
			givenTimeout: time.Minute,
			wantCode:     "c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan callbackResult, 1)
			if tt.givenResult != nil {
				ch <- *tt.givenResult
			}
			cs := &callbackServer{results: ch}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.givenCancel {
				cancel()
			}

			code, _, err := cs.wait(ctx, tt.givenTimeout)
			if wegostrings.IsNotEmpty(tt.wantErrContains) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func TestCallbackServer_ShutdownIsSafeWithoutAServer(t *testing.T) {
	cs := &callbackServer{results: make(chan callbackResult, 1)}
	assert.NotPanics(t, cs.shutdown)
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

// mustGet issues a throwaway GET, standing in for the operator's browser.
func mustGet(t *testing.T, rawURL string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	require.NoError(t, err)
	//nolint:gosec // G704: a loopback URL this test built itself.
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

// stubError is a minimal error for table cases that need a pre-baked failure.
type stubError string

func (e stubError) Error() string { return string(e) }
