package errors

import (
	"context"
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
