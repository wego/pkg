package retry

import (
	"github.com/eapache/go-resiliency/retrier"
	"github.com/wego/pkg/errors"
	"net"
	"net/http"
	"strings"
	"time"
)

type retryClassifier struct{}

func (r *retryClassifier) Classify(err error) retrier.Action {
	if err == nil {
		return retrier.Succeed
	}
	switch err := err.(type) {
	case net.Error:
		return retrier.Retry
	case *errors.Error:
		if _, ok := err.Err.(net.Error); ok {
			return retrier.Retry
		}
		switch errors.Code(err) {
		// retry only on 500,502/504/429/401/403
		case
			int(errors.Retry),
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusGatewayTimeout,
			http.StatusTooManyRequests:
			return retrier.Retry
		}
	default:
		if strings.Contains(err.Error(), "error unmarshalling") {
			return retrier.Retry
		}
	}
	return retrier.Fail
}

// NewRetrier create a ExponentialBackoff retrier
func NewRetrier(maxRetries int, initialBackoff time.Duration) *retrier.Retrier {
	return retrier.New(retrier.ExponentialBackoff(maxRetries, initialBackoff), &retryClassifier{})
}
