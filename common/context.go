package common

import (
	"context"
	"github.com/DataDog/datadog-go/statsd"
)

// ContextKey is key to be used to store values to context
type ContextKey string

// Extras are extra information to be added into context
type Extras map[string]interface{}

// Basics are basics information to be added into context
type Basics map[string]interface{}

const (
	// keep it private to avoid conflicts
	ctxBasics ContextKey = "basics"
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

// GetBasic gets basic from the context with key
func GetBasic(ctx context.Context, key string) (value interface{}) {
	if ctx != nil {
		if basics, ok := ctx.Value(ctxBasics).(Basics); ok {
			value, _ = basics[key]
		}
	}
	return
}

// GetBasics gets basics from the context if any
func GetBasics(ctx context.Context) (value Basics) {
	if ctx != nil {
		value, _ = ctx.Value(ctxBasics).(Basics)
	}
	return
}

// SetBasic returns a copy of parent context with basic key value added into it
func SetBasic(parent context.Context, key string, value interface{}) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	basics, ok := parent.Value(ctxBasics).(Basics)
	if !ok {
		basics = make(map[string]interface{})
	}
	basics[key] = value

	return context.WithValue(parent, ctxBasics, basics)
}

// SetBasics returns a copy of parent context with basics added into it
func SetBasics(parent context.Context, basics Basics) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, ctxBasics, basics)
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

// GetExtra gets Extra from the context with key
func GetExtra(ctx context.Context, key string) (value interface{}) {
	if ctx != nil {
		if Extras, ok := ctx.Value(ctxExtras).(Extras); ok {
			value, _ = Extras[key]
		}
	}
	return
}

// SetExtra returns a copy of parent context with Extra key value added into it
func SetExtra(parent context.Context, key string, value interface{}) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	Extras, ok := parent.Value(ctxExtras).(Extras)
	if !ok {
		Extras = make(map[string]interface{})
	}
	Extras[key] = value

	return context.WithValue(parent, ctxExtras, Extras)
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
