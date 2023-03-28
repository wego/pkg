package logger

import (
	"time"

	"go.uber.org/zap"
)

// RequestType ...
type RequestType string
type logType string
type contextKey string

const (
	logDir                  = "log"
	ultronExFileName        = "ultronex.{{env}}.log"
	partnerRequestsFileName = "partner_requests.{{env}}.log"
	requestsFileName        = "requests.{{env}}.log"
	defaultReplacement      = "[Filtered by Wego]"
	defaultMaskChar         = "*"

	logTypeUltronex       logType = "ultronEx"
	logTypePartnerRequest logType = "partnerRequest"
	logTypeRequest        logType = "request"

	contextKeyRequest     contextKey = "request"
	contextKeyRequestType contextKey = "requestType"

	slackPostingMsgLimitTime = time.Second
)

var (
	loggers          map[logType]*zap.Logger
	sensitiveHeaders = map[string]bool{
		"authorization":   true,
		"x-forter-siteid": true,
	}
)
