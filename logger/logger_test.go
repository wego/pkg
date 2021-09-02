package logger_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/logger"
)

func Test_ContextWithRequestType(t *testing.T) {
	assert := assert.New(t)

	ctx := logger.ContextWithRequestType(nil, "")
	assert.NotNil(ctx)

	ctx = logger.ContextWithRequestType(context.Background(), "")
	assert.NotNil(ctx)

	ctx = logger.ContextWithRequestType(context.Background(), logger.RequestTypeAuthorizePayment)
	assert.NotNil(ctx)
}

func Test_RequestTypeFromContext(t *testing.T) {
	assert := assert.New(t)

	reqType := logger.RequestTypeFromContext(nil)
	assert.Zero(reqType)

	reqType = logger.RequestTypeFromContext(context.Background())
	assert.Zero(reqType)

	ctx := logger.ContextWithRequestType(nil, logger.RequestTypeAuthorizePayment)
	reqType = logger.RequestTypeFromContext(ctx)
	assert.Equal(logger.RequestTypeAuthorizePayment, reqType)
}
