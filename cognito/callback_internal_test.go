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
	tests := []struct {
		name            string
		givenQuery      string
		wantStatus      int
		wantCode        string
		wantState       string
		wantErrContains string
	}{
		{
			name:            "authorize error is surfaced with its description",
			givenQuery:      "error=access_denied&error_description=user+said+no",
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "user said no",
		},
		{
			name:            "authorize error without a description still reports",
			givenQuery:      "error=server_error",
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "server_error",
		},
		{
			name:            "missing code is rejected",
			givenQuery:      "state=abc",
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "no authorization code",
		},
		{
			name:       "code and state are captured",
			givenQuery: "code=the-code&state=the-state",
			wantStatus: http.StatusOK,
			wantCode:   "the-code",
			wantState:  "the-state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan callbackResult, 1)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/callback?"+tt.givenQuery, nil)

			callbackHandler(ch)(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")

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

func TestCallbackHandler_IsSingleUse(t *testing.T) {
	ch := make(chan callbackResult, 1)
	handler := callbackHandler(ch)

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
	cs, err := startCallbackServer(addr, "/callback")

	require.Error(t, err, "binding an occupied port must fail rather than silently pick another")
	assert.Nil(t, cs)
	assert.Contains(t, err.Error(), addr, "the error must name the address the operator has to free")
}

func TestStartCallbackServer_ServesTheCallback(t *testing.T) {
	addr := mustFreeAddr(t)
	cs, err := startCallbackServer(addr, "/callback")
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
