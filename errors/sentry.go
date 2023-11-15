package errors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/wego/pkg/env"
)

// CaptureError captures an error to sentry & set level as error
func CaptureError(ctx context.Context, err error) {
	capture(ctx, err, sentry.LevelError)
}

// CaptureWarning captures an error to sentry & set level as warning
func CaptureWarning(ctx context.Context, err error) {
	capture(ctx, err, sentry.LevelWarning)
}

func capture(ctx context.Context, err error, level sentry.Level) {
	if !env.IsProduction() && !env.IsStaging() {
		return
	}

	hub := getHub(ctx)
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
		enrichScope(ctx, scope, err)
		hub.CaptureException(err)
	})
}

func enrichScope(ctx context.Context, scope *sentry.Scope, err error) {
	errorCode := fmt.Sprint(Code(err))
	scope.SetTag(SentryErrorCode, errorCode)
	fingerprint := []string{errorCode, err.Error()}

	e, ok := err.(*Error)
	if ok {
		for key, value := range e.getBasics() {
			if tag, err := json.Marshal(value); err == nil {
				scope.SetTag(key, string(tag))
				fingerprint = append(fingerprint, key)
			}
		}

		// extra is not searchable in sentry
		ops := ops(e)
		scope.SetExtra(SentryOperations, ops)
		for _, o := range ops {
			fingerprint = append(fingerprint, string(o))
		}

		for k, v := range e.getExtras() {
			scope.SetExtra(k, v)
		}
	}

	reqID, ok := ctx.Value(SentryRequestID).(string)
	if ok {
		scope.SetTag(SentryRequestID, reqID)
		fingerprint = append(fingerprint, reqID)
	}

	scope.SetFingerprint(fingerprint)
}

func getHub(ctx context.Context) (hub *sentry.Hub) {
	hub = sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	return
}
