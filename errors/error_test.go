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
	ctx0 := context.Background()

	ctx1 := context.WithValue(ctx0, testContextKey("keyNotStored1"), "valueNotStored1")
	ctx1 = common.SetBasics(ctx1, common.Basics{"key1": "value1"})

	ctx2 := context.WithValue(ctx1, testContextKey("keyNotStored2"), "valueNotStored2")
	ctx2 = common.SetBasics(ctx2, common.Basics{"key1": "value1-replaced"})
	ctx2 = common.SetExtras(ctx2, common.Extras{"key2": "value2"})

	childErr2 := New(ctx2, Op(grandchildOp), NotFound, grandchildErrMsg)
	childErr1 := New(ctx1, Op(childOp), BadRequest, childErrMsg, childErr2)
	gotErr := New(ctx0, Op(parentOp), parentErrMsg, childErr1)

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
				"key1": "value1-replaced",
			},
			"extras": common.Extras{
				"key2": "value2",
			},
		},
	}

	assert.Equal(wantErr, gotErr)
}

func TestNew_NestedContext_ChildsOwnContext(t *testing.T) {
	assert := assert.New(t)

	// Child setting their own `context.Background()`
	ctx1 := context.WithValue(context.Background(), testContextKey("keyNotStored1"), "valueNotStored1")
	ctx1 = common.SetBasics(ctx1, common.Basics{"key1": "value1"})

	// Child setting their own `context.Background()`
	ctx2 := context.WithValue(context.Background(), testContextKey("keyNotStored2"), "valueNotStored2")
	ctx2 = common.SetBasics(ctx2, common.Basics{"key1": "value1-replaced"})
	ctx2 = common.SetExtras(ctx2, common.Extras{"key2": "value2"})

	childErr2 := New(ctx2, Op(grandchildOp), NotFound, grandchildErrMsg)
	childErr1 := New(ctx1, Op(childOp), BadRequest, childErrMsg, childErr2)
	gotErr := New(context.Background(), Op(parentOp), parentErrMsg, childErr1)

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
				"key1": "value1-replaced",
			},
			"extras": common.Extras{
				"key2": "value2",
			},
		},
	}

	assert.Equal(wantErr, gotErr)
}

func TestNew_NestedContext_NilChildContext(t *testing.T) {
	assert := assert.New(t)

	ctx0 := context.Background()
	ctx0 = common.SetBasics(ctx0, common.Basics{"key1": "value1"})

	childErr1 := New(nil, Op(childOp), BadRequest, childErrMsg)
	gotErr := New(ctx0, Op(parentOp), parentErrMsg, childErr1)

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

	assert.Equal(wantErr, gotErr)
}

func Test_basics(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	childCtx := common.SetBasic(ctx, "key1", "value1")
	childCtx = common.SetBasic(childCtx, "key2", "value2")
	childCtx = common.SetExtra(childCtx, "key3", "value3")

	childErr := New(childCtx, Op(childOp), childErrMsg)
	err := New(ctx, Op(parentOp), "msg", childErr)

	e, ok := err.(*Error)
	assert.True(ok)

	wantBasics := common.Basics{
		"key1": "value1",
		"key2": "value2",
	}
	assert.Equal(wantBasics, e.basics())
}

func Test_extras(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	childCtx := common.SetBasic(ctx, "key1", "value1")
	childCtx = common.SetExtra(childCtx, "key2", "value2")
	childCtx = common.SetExtra(childCtx, "key3", "value3")

	childErr := New(childCtx, Op(childOp), childErrMsg)
	err := New(ctx, Op(parentOp), "msg", childErr)

	e, ok := err.(*Error)
	assert.True(ok)

	wantExtras := common.Extras{
		"key2": "value2",
		"key3": "value3",
	}
	assert.Equal(wantExtras, e.extras())
}