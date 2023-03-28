package logger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/wego/pkg/errors"
	"go.uber.org/zap"
)

// ContextWithRequestType returns a new context from a parent context with request type added into it
func ContextWithRequestType(ctx context.Context, reqType RequestType) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, contextKeyRequestType, reqType)
}

// RequestTypeFromContext gets the request type from context
func RequestTypeFromContext(ctx context.Context) (reqType RequestType) {
	if ctx == nil {
		return
	}

	reqType, _ = ctx.Value(contextKeyRequestType).(RequestType)
	return
}

// ContextWithRequest returns a new context from a parent context with request added into it
func ContextWithRequest(ctx context.Context, req *Request) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, contextKeyRequest, req)
}

// RequestFromContext gets the request from context
func RequestFromContext(ctx context.Context) (req *Request) {
	if ctx == nil {
		return
	}

	req, _ = ctx.Value(contextKeyRequest).(*Request)
	return
}

// LogUltronEx logs a msg to UltronEx local file
func LogUltronEx(msg UltronExMsg) {
	logger := loggers[logTypeUltronex]
	if logger != nil && msg != (UltronExMsg{}) {
		time.Sleep(slackPostingMsgLimitTime)
		// UltronEx require the key as `msg`
		logger.Info("", zap.Object("msg", msg))
	}
}

// LogPartnerRequest logs a partner request to local file
func LogPartnerRequest(log Request) {
	logger := loggers[logTypePartnerRequest]
	if logger != nil && len(log.Type) > 0 {
		logger.Info("", log.fields()...)
	}
}

// LogRequest logs a request to local file
func LogRequest(log Request) {
	logger := loggers[logTypeRequest]
	if logger != nil && len(log.Type) > 0 {
		logger.Info("", log.fields()...)
	}
}

// Init initializes loggers
func Init() error {
	loggers = make(map[logType]*zap.Logger, 2)

	uLog, err := initLogger(ultronExFileName)
	if err != nil {
		return errors.New("cannot init UltronEx logger", err)
	}
	loggers[logTypeUltronex] = uLog

	prLog, err := initLogger(partnerRequestsFileName)
	if err != nil {
		return errors.New("cannot init partner request logger", err)
	}
	loggers[logTypePartnerRequest] = prLog

	rLog, err := initLogger(requestsFileName)
	if err != nil {
		return errors.New("cannot init request logger", err)
	}
	loggers[logTypeRequest] = rLog
	return nil
}

// Sync syncs all loggers
func Sync() {
	for _, logger := range loggers {
		if logger != nil {
			logger.Sync()
		}
	}
}

func initLogger(fileName string) (logger *zap.Logger, err error) {
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		return
	}

	logPath := filepath.Join(logDir, strings.Replace(fileName, "{{env}}", viper.GetString("env"), 1))
	_, err = os.Create(logPath)
	if err != nil {
		return
	}

	logConfig := zap.NewProductionConfig()
	// remove unwanted keys
	logConfig.EncoderConfig.MessageKey = ""
	logConfig.EncoderConfig.LevelKey = ""
	logConfig.EncoderConfig.CallerKey = ""
	logConfig.EncoderConfig.TimeKey = ""
	// set output to file
	logConfig.OutputPaths = []string{logPath}
	return logConfig.Build()
}
