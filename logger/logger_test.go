package logger_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/logger"
)

var (
	requestType logger.RequestType = "RequestType"
)

func Test_ContextWithRequestType(t *testing.T) {
	assert := assert.New(t)

	ctx := logger.ContextWithRequestType(nil, "")
	assert.NotNil(ctx)

	ctx = logger.ContextWithRequestType(context.Background(), "")
	assert.NotNil(ctx)

	ctx = logger.ContextWithRequestType(context.Background(), requestType)
	assert.NotNil(ctx)
}

func Test_RequestTypeFromContext(t *testing.T) {
	assert := assert.New(t)

	reqType := logger.RequestTypeFromContext(nil)
	assert.Zero(reqType)

	reqType = logger.RequestTypeFromContext(context.Background())
	assert.Zero(reqType)

	ctx := logger.ContextWithRequestType(nil, requestType)
	reqType = logger.RequestTypeFromContext(ctx)
	assert.Equal(requestType, reqType)
}
