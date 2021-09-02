package common_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/wego/pkg/common/mocks"

	"github.com/stretchr/testify/suite"
)

type RequestIDServiceSuite struct {
	suite.Suite
	requestIDService *mocks.RequestIDService
}

// SetupTest runs before each Test
func (s *RequestIDServiceSuite) SetupTest() {
	s.requestIDService = &mocks.RequestIDService{}
}

func TestRequestIDService(t *testing.T) {
	suite.Run(t, new(RequestIDServiceSuite))
}

func (s *RequestIDServiceSuite) Test_GenerateRequestID() {
	requestID := uuid.New().String()
	s.requestIDService.On("GenerateRequestID").Return(requestID)

	generated := s.requestIDService.GenerateRequestID()

	s.Equal(requestID, generated)
}

func (s *RequestIDServiceSuite) Test_GenerateRequestIDWithCtx() {
	ctx := context.Background()
	requestID := uuid.New().String()
	s.requestIDService.On("GenerateRequestIDWithCtx", ctx).Return(requestID)

	generated := s.requestIDService.GenerateRequestIDWithCtx(ctx)

	s.Equal(requestID, generated)
}
