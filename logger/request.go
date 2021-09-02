package logger

import (
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Headers ...
type Headers map[string]string

type header struct {
	name  string
	value string
}

// PartnerRequest contains information of requests we sent to our partners
type PartnerRequest struct {
	PartnerCode string
	PartnerRef  string
	Request
}

// Request contains general information of a request
type Request struct {
	Type            RequestType
	PaymentRef      string
	DeploymentRef   string
	ClientCode      string
	TransactionRef  string
	Method          string
	URL             string
	RequestHeaders  Headers
	RequestBody     string
	IP              string
	StatusCode      int32
	ResponseHeaders Headers
	ResponseBody    string
	RequestedAt     time.Time
	Duration        time.Duration
	Error           error
}

func (r *PartnerRequest) fields() []zapcore.Field {
	fields := []zapcore.Field{
		zap.String("partner_code", r.PartnerCode),
		zap.String("partner_ref", r.PartnerRef),
	}
	fields = append(fields, r.Request.fields()...)

	return fields
}

func (r *Request) fields() []zapcore.Field {
	fields := []zapcore.Field{
		zap.String("type", string(r.Type)),
		zap.String("payment_ref", r.PaymentRef),
		zap.String("virtual_card_deployment_ref", r.DeploymentRef),
		zap.String("client_code", r.ClientCode),
		zap.String("transaction_ref", r.TransactionRef),
		zap.String("method", r.Method),
		zap.String("url", r.URL),
		zap.Array("request_headers", r.RequestHeaders),
		zap.String("request_body", r.RequestBody),
		zap.String("ip", r.IP),
		zap.Int32("status_code", r.StatusCode),
		zap.Array("response_headers", r.ResponseHeaders),
		zap.String("response_body", r.ResponseBody),
		zap.String("requested_at", r.RequestedAt.Format(time.RFC3339)),
		zap.Int64("duration_in_ms", r.Duration.Milliseconds()),
	}
	if r.Error != nil {
		fields = append(fields, zap.String("error", r.Error.Error()))
	}

	return fields
}

// MarshalLogArray marshal Headers to zap log array
// Need to implement this to log it with zap.Array
func (h Headers) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for k, v := range h {
		if sensitive := sensitiveHeaders[strings.ToLower(k)]; sensitive {
			v = defaultReplacement
		}
		enc.AppendObject(header{
			name:  k,
			value: v,
		})
	}
	return nil
}

// MarshalLogObject marshal header to zap log object
// The struct need to implement this, so we can log it as object
func (h header) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", h.name)
	enc.AddString("value", h.value)
	return nil
}
