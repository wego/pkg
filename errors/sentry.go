package errors

import (
	"context"
	"encoding/json"
	"fmt"

	"maps"

	"github.com/getsentry/sentry-go"
	"github.com/wego/pkg/env"
)

// ErrorData carries optional extra data to attach to the error event
type ErrorData struct {
	Extra map[string]any
	Tags  map[string]string
}

// CaptureError captures an error to sentry & set level as error.
// It is backward compatible: info is an optional parameter.
func CaptureError(ctx context.Context, err error, info ...*ErrorData) {
	var ei *ErrorData
	if len(info) > 0 {
		ei = info[0]
	}
	capture(ctx, err, sentry.LevelError, ei)
}

// CaptureWarning captures an error to sentry & set level as warning.
// It is backward compatible: info is an optional parameter.
func CaptureWarning(ctx context.Context, err error, info ...*ErrorData) {
	var ei *ErrorData
	if len(info) > 0 {
		ei = info[0]
	}
	capture(ctx, err, sentry.LevelWarning, ei)
}

func capture(ctx context.Context, err error, level sentry.Level, info *ErrorData) {
	if !env.IsProduction() && !env.IsStaging() {
		return
	}

	hub := getHub(ctx)
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
		enrichScope(ctx, scope, err, info)
		hub.CaptureException(err)
	})
}

func enrichScope(ctx context.Context, scope *sentry.Scope, err error, info *ErrorData) {
	// Prepare tags and extras to set
	var tagsToSet = make(map[string]string)
	var extrasToSet = make(map[string]any)

	// Fetch error code
	errorCode := fmt.Sprint(Code(err))
	tagsToSet[SentryErrorCode] = errorCode

	// Fingeprinting is handed over to the SDK ({{default}} is the default fingerprint), with the additional error code field to add a dimension of uniqueness
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
		maps.Copy(extrasToSet, e.extras())
	}

	// If any info is provided, add them to the event scope
	if info != nil {
		// Merge info.Extra into extrasToSet
		maps.Copy(extrasToSet, info.Extra)

		// Merge info.Tags into tagsToSet
		maps.Copy(tagsToSet, info.Tags)
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
