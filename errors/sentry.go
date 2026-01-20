package errors

import (
	"context"
	"encoding/json"
	"fmt"

	"maps"

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
	// Prepare tags and extras to set
	var tagsToSet = make(map[string]string)
	var extrasToSet = make(map[string]any)

	// Fetch error code
	errorCode := fmt.Sprint(Code(err))
	tagsToSet[SentryErrorCode] = errorCode

	// Fingerprinting is handed over to the SDK ({{default}} is the default fingerprint), with the additional error code field to add a dimension of uniqueness
	fingerprint := []string{"{{default}}", errorCode}

	// If the error is an Error type, we can enrich the scope with the basics and extras
	e, ok := err.(*Error)
	if ok {
		// For each basic key-value pair, set it as a tag
		for key, value := range e.basics() {
			if tag, err := json.Marshal(value); err == nil {
				tagsToSet[key] = string(tag)
			}
		}

		// For each operation, set it as an extra
		// Note: extra is not searchable in Sentry
		ops := ops(e)
		extrasToSet[SentryOperations] = ops

		// Merge e.extras() into extrasToSet
		// Note: maps.Copy overwrites existing keys in the destination map.
		maps.Copy(extrasToSet, e.extras())
	}

	// Get the request ID from the context
	reqID, ok := ctx.Value(SentryRequestID).(string)
	if ok {
		tagsToSet[SentryRequestID] = reqID
	}

	// Finally, update the scope with the tags, extras and fingerprint
	scope.SetTags(tagsToSet)
	scope.SetExtras(extrasToSet)
	scope.SetFingerprint(fingerprint)
}

func getHub(ctx context.Context) (hub *sentry.Hub) {
	hub = sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	return
}
