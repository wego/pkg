package common

import (
	"context"

	"github.com/DataDog/datadog-go/statsd"
)

// ContextKey is key to be used to store values to context
type ContextKey string

// Extras are extra information to be added into context
type Extras map[string]interface{}

const (
	CtxClientCode ContextKey = "clientCode"
	// keep it private to avoid conflicts
	ctxExtras ContextKey = "extras"
	ctxStatsD ContextKey = "statsD"
)

// GetString gets string from a context by ContextKey
func GetString(ctx context.Context, key ContextKey) (value string) {
	if ctx != nil {
		value, _ = ctx.Value(key).(string)
	}
	return
}

// GetExtras gets extras from the context if any
func GetExtras(ctx context.Context) (value Extras) {
	if ctx != nil {
		value, _ = ctx.Value(ctxExtras).(Extras)
	}
	return
}

// SetExtras returns a copy of parent context with extras added into it
func SetExtras(parent context.Context, extras Extras) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, ctxExtras, extras)
}

// GetStatsD gets statsD client from the context if any
func GetStatsD(ctx context.Context) (statsD *statsd.Client) {
	if ctx != nil {
		statsD, _ = ctx.Value(ctxStatsD).(*statsd.Client)
	}
	return
}

// SetStatsD returns a copy of parent context with statsD client added into it
func SetStatsD(parent context.Context, statsD *statsd.Client) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, ctxStatsD, statsD)
}
