package errors

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wego/pkg/common"
)

func TestCaptureErrorMetadataContext(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	ctx := common.SetBasics(context.Background(), common.Basics{"service_tag": "probe"})
	ctx = common.SetExtras(ctx, common.Extras{"nested": map[string]any{"count": 3, "enabled": true}, "nullable": nil})
	structured := New(ctx, Op("outer"), New(Op("inner"), Unexpected, "synthetic failure"))
	collisionCtx := common.SetExtras(context.Background(), common.Extras{SentryOperations: "caller override"})

	tests := []struct {
		name        string
		givenError  error
		wantContext sentry.Context
		wantTag     string
	}{
		{name: "plain error has no structured context", givenError: fmt.Errorf("plain failure")},
		{name: "empty structured error retains empty operations", givenError: &Error{}, wantContext: sentry.Context{SentryOperations: []Op(nil)}},
		{name: "caller metadata retains precedence over operations", givenError: New(collisionCtx, Op("operation"), Unexpected, "failure"), wantContext: sentry.Context{SentryOperations: "caller override"}},
		{name: "nested values and operation chain survive", givenError: structured, wantContext: sentry.Context{
			"nested": map[string]any{"count": 3, "enabled": true}, "nullable": nil,
			SentryOperations: []Op{"outer", "inner"},
		}, wantTag: `"probe"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &eventRecordingTransport{}
			client, err := sentry.NewClient(sentry.ClientOptions{Transport: transport})
			require.NoError(t, err)
			scope := sentry.NewScope()
			scope.SetContext("unrelated", sentry.Context{"keep": true})
			hub := sentry.NewHub(client, scope)
			// The capture API reads this existing string key from caller contexts.
			//revive:disable-next-line:context-keys-type
			ctx := context.WithValue(context.Background(), SentryRequestID, "metadata-test-request")
			ctx = sentry.SetHubOnContext(ctx, hub)
			CaptureError(ctx, tt.givenError)
			require.Len(t, transport.events, 1)
			event := transport.events[0]
			assert.Equal(t, tt.wantContext, event.Contexts["wego_error"])
			assert.Equal(t, sentry.Context{"keep": true}, event.Contexts["unrelated"])
			assert.Equal(t, tt.wantTag, event.Tags["service_tag"])
			assert.Equal(t, "metadata-test-request", event.Tags[SentryRequestID])
			assert.Equal(t, "500", event.Tags[SentryErrorCode])
			assert.Equal(t, []string{"{{default}}", "500"}, event.Fingerprint)
			assert.Equal(t, sentry.LevelError, event.Level)
		})
	}
}

func TestCaptureErrorMetadataDoesNotLeak(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	transport := &eventRecordingTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{Transport: transport})
	require.NoError(t, err)
	hub := sentry.NewHub(client, sentry.NewScope())
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	errCtx := common.SetExtras(ctx, common.Extras{"only_first": "synthetic metadata"})
	CaptureError(ctx, New(errCtx, Unexpected, "first"))
	CaptureWarning(ctx, fmt.Errorf("second"))
	require.Len(t, transport.events, 2)
	assert.Equal(t, "synthetic metadata", transport.events[0].Contexts["wego_error"]["only_first"])
	assert.NotContains(t, transport.events[1].Contexts, "wego_error")
	assert.Equal(t, sentry.LevelWarning, transport.events[1].Level)
}

type eventRecordingTransport struct{ events []*sentry.Event }

func (*eventRecordingTransport) Configure(sentry.ClientOptions)        {}
func (*eventRecordingTransport) Flush(time.Duration) bool              { return true }
func (*eventRecordingTransport) FlushWithContext(context.Context) bool { return true }
func (*eventRecordingTransport) Close()                                {}
func (t *eventRecordingTransport) SendEvent(event *sentry.Event)       { t.events = append(t.events, event) }
