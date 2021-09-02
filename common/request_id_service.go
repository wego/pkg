//go:generate mockery --name=RequestIDService

package common

import (
	"context"

	"github.com/google/uuid"
)

// RequestIDService generate request ID
type RequestIDService interface {
	GenerateRequestIDWithCtx(ctx context.Context) string
	GenerateRequestID() string
}

type requestIDService struct {
}

// NewRequestIDService create a requestIDService
func NewRequestIDService() RequestIDService {
	return &requestIDService{}
}

func (r *requestIDService) GenerateRequestIDWithCtx(ctx context.Context) string {
	return uuid.New().String()
}

func (r *requestIDService) GenerateRequestID() string {
	return uuid.New().String()
}
