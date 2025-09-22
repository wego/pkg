package errors

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/common"
)

const (
	parentOp         = "parent op"
	parentErrMsg     = "parent error message"
	childOp          = "child op"
	childErrMsg      = "child error message"
	grandchildOp     = "grandchild op"
	grandchildErrMsg = "grandchild error message"
)

type testContextKey string

func TestNew(t *testing.T) {
	childErr := New(Op(childOp), BadRequest, childErrMsg)

	// Test for data race condition when trying to `e.propagateContexts()`.
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := New(Op(parentOp), parentErrMsg, childErr)
			wantErr := &Error{
				Op:  parentOp,
				msg: parentErrMsg,
				Err: &Error{
					Op:   childOp,
					Kind: BadRequest,
					msg:  childErrMsg,
				},
			}

			assert.Equal(t, wantErr, err)
		}()
	}

	wg.Wait()
}

func TestNew_NestedContext_ParentContext(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	ctx = common.SetBasics(ctx, common.Basics{"key1": "value1-original"})
	ctx = common.SetBasic(ctx, "key2", "value2")

	childCtx := context.WithValue(ctx, testContextKey("keyNotStored1"), "valueNotStored1")
	childCtx = common.SetBasics(childCtx, common.Basics{"key1": "value1-child"})

	grandchildCtx := context.WithValue(childCtx, testContextKey("keyNotStored2"), "valueNotStored2")
	grandchildCtx = common.SetBasics(grandchildCtx, common.Basics{"key1": "value1-grandchild"})
	grandchildCtx = common.SetExtras(grandchildCtx, common.Extras{"key3": "value3"})

	grandChildErr := New(Op(grandchildOp), NotFound, grandchildErrMsg).
		WithContext(grandchildCtx)

	childErr := New(Op(childOp), BadRequest, childErrMsg, grandChildErr).
		WithContext(childCtx)

	err := New(Op(parentOp), parentErrMsg, childErr).
		WithContext(ctx)

	wantErr := &Error{
		Op:  parentOp,
		msg: parentErrMsg,
		Err: &Error{
			Op:   childOp,
			Kind: BadRequest,
			msg:  childErrMsg,
			Err: &Error{
				Op:   grandchildOp,
				Kind: NotFound,
				msg:  grandchildErrMsg,
			},
		},
		ctx: map[string]any{
			"basics": common.Basics{
				"key1": "value1-grandchild", // We should keep the latest value which is usually set by the lowest node.
				"key2": "value2",
			},
			"extras": common.Extras{
				"key3": "value3",
			},
		},
	}

	assert.Equal(wantErr, err)
}

func TestNew_NestedContext_ChildsOwnContext(t *testing.T) {
	assert := assert.New(t)

	// Child setting their own `context.Background()`
	childCtx := context.WithValue(context.Background(), testContextKey("key-not-stored1"), "value-not-stored1")
	childCtx = common.SetBasics(childCtx, common.Basics{"key1": "value1-child"})

	// Child setting their own `context.Background()`
	grandchildCtx := context.WithValue(context.Background(), testContextKey("key-not-stored2"), "value-not-stored2")
	grandchildCtx = common.SetBasics(grandchildCtx, common.Basics{"key1": "value1-grandchild"})
	grandchildCtx = common.SetExtras(grandchildCtx, common.Extras{"key2": "value2"})

	grandChildErr := New(Op(grandchildOp), NotFound, grandchildErrMsg).
		WithContext(grandchildCtx)

	childErr := New(Op(childOp), BadRequest, childErrMsg, grandChildErr).
		WithContext(childCtx)

	err := New(Op(parentOp), parentErrMsg, childErr)

	wantErr := &Error{
		Op:  parentOp,
		msg: parentErrMsg,
		Err: &Error{
			Op:   childOp,
			Kind: BadRequest,
			msg:  childErrMsg,
			Err: &Error{
				Op:   grandchildOp,
				Kind: NotFound,
				msg:  grandchildErrMsg,
			},
		},
		ctx: map[string]any{
			"basics": common.Basics{
				"key1": "value1-grandchild", // We should keep the latest value which is usually set by the lowest node.
			},
			"extras": common.Extras{
				"key2": "value2",
			},
		},
	}

	assert.Equal(wantErr, err)
}

func TestNew_NestedContext_NilChildContext(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = common.SetBasics(ctx, common.Basics{"key1": "value1"})

	childErr := New(Op(childOp), BadRequest, childErrMsg)
	err := New(Op(parentOp), parentErrMsg, childErr).
		WithContext(ctx)

	wantErr := &Error{
		Op:  parentOp,
		msg: parentErrMsg,
		Err: &Error{
			Op:   childOp,
			Kind: BadRequest,
			msg:  childErrMsg,
		},
		ctx: map[string]any{
			"basics": common.Basics{
				"key1": "value1",
			},
		},
	}

	assert.Equal(wantErr, err)
}

func Test_basics(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	childCtx := common.SetBasic(ctx, "key1", "value1")
	childCtx = common.SetBasic(childCtx, "key2", "value2")
	childCtx = common.SetExtra(childCtx, "key3", "value3")

	childErr := New(Op(childOp), childErrMsg).
		WithContext(childCtx)

	err := New(Op(parentOp), parentErrMsg, childErr).
		WithContext(ctx)

	assert.NotNil(err)
	wantBasics := common.Basics{
		"key1": "value1",
		"key2": "value2",
	}
	assert.Equal(wantBasics, err.basics())
}

func Test_extras(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	childCtx := common.SetBasic(ctx, "key1", "value1")
	childCtx = common.SetExtra(childCtx, "key2", "value2")
	childCtx = common.SetExtra(childCtx, "key3", "value3")

	childErr := New(Op(childOp), childErrMsg).
		WithContext(childCtx)

	err := New(Op(parentOp), parentErrMsg, childErr).
		WithContext(ctx)

	assert.NotNil(err)
	wantExtras := common.Extras{
		"key2": "value2",
		"key3": "value3",
	}
	assert.Equal(wantExtras, err.extras())
}

func TestError_ContextHandling(t *testing.T) {
	tests := []struct {
		name              string
		setupFunc         func() *Error
		expectedBasics    common.Basics
		expectedExtras    common.Extras
		expectContextNil  bool
		expectChildCtxNil bool
	}{
		{
			name: "context with multiple layers and overrides",
			setupFunc: func() *Error {
				ctx := context.Background()
				ctx = common.SetBasics(ctx, common.Basics{
					"request_id": "req-123",
					"user_id":    "user-456",
				})
				ctx = common.SetExtras(ctx, common.Extras{
					"trace_id": "trace-789",
					"span_id":  "span-abc",
				})

				childCtx := common.SetBasics(ctx, common.Basics{
					"user_id":    "user-override", // This should override parent
					"session_id": "session-def",   // This should be added
				})
				childCtx = common.SetExtras(childCtx, common.Extras{
					"span_id":   "span-override", // This should override parent
					"operation": "child-op",      // This should be added
				})

				childErr := New(Op("child.operation"), BadRequest, "child error").
					WithContext(childCtx)

				return New(Op("parent.operation"), "parent error", childErr).
					WithContext(ctx)
			},
			expectedBasics: common.Basics{
				"request_id": "req-123",      // From parent, unchanged
				"user_id":    "user-override", // From child, overridden
				"session_id": "session-def",  // From child, new
			},
			expectedExtras: common.Extras{
				"trace_id":  "trace-789",      // From parent, unchanged
				"span_id":   "span-override",  // From child, overridden
				"operation": "child-op",       // From child, new
			},
			expectContextNil:  false,
			expectChildCtxNil: true,
		},
		{
			name: "context without child error",
			setupFunc: func() *Error {
				ctx := context.Background()
				ctx = common.SetBasics(ctx, common.Basics{"key": "value"})
				ctx = common.SetExtras(ctx, common.Extras{"extra": "data"})

				return New(Op("test.operation"), BadRequest, "test error").
					WithContext(ctx)
			},
			expectedBasics: common.Basics{"key": "value"},
			expectedExtras: common.Extras{"extra": "data"},
			expectContextNil: false,
		},
		{
			name: "empty context",
			setupFunc: func() *Error {
				ctx := context.Background()
				return New(Op("test.operation"), BadRequest, "test error").
					WithContext(ctx)
			},
			expectedBasics:   nil,
			expectedExtras:   nil,
			expectContextNil: false, // WithContext always creates ctx map
		},
		{
			name: "no context set",
			setupFunc: func() *Error {
				return New(Op("test.operation"), BadRequest, "test error")
			},
			expectedBasics:   nil,
			expectedExtras:   nil,
			expectContextNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			err := tt.setupFunc()

			assert.Equal(tt.expectedBasics, err.basics())
			assert.Equal(tt.expectedExtras, err.extras())

			if tt.expectContextNil {
				assert.Nil(err.ctx)
			} else {
				assert.NotNil(err.ctx)
				if tt.expectedBasics != nil {
					assert.Contains(err.ctx, "basics")
				}
				if tt.expectedExtras != nil {
					assert.Contains(err.ctx, "extras")
				}
			}

			if tt.expectChildCtxNil {
				childErrTyped, ok := err.Err.(*Error)
				if ok {
					assert.Nil(childErrTyped.ctx, "child error context should be cleared after propagation")
				}
			}
		})
	}
}
