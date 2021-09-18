package logger

import (
	"encoding/json"
	"github.com/wego/pkg/common"
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
	Request
}

// Request contains general information of a request
type Request struct {
	Type            RequestType
	Basics          common.Basics
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
	return r.Request.fields()
}

func (r *Request) fields() []zapcore.Field {
	var fields []zapcore.Field
	for key, value := range r.Basics {
		if v, err := json.Marshal(value); err == nil {
			fields = append(fields, zap.String(key, string(v)))
		}
	}
	fields = append(fields, []zapcore.Field{
		zap.String("type", string(r.Type)),
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
	}...)

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
		err := enc.AppendObject(header{
			name:  k,
			value: v,
		})
		if err != nil {
			return err
		}
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
