package logger

import (
	"go.uber.org/zap/zapcore"
)

// UltronExMsg ...
type UltronExMsg struct {
	Channel string
	Text    string
	Payload string
	Title   string
}

// MarshalLogObject marshal UltronExMsg to zap log object
// The struct need to implement this, so we can log it with zap.Object
func (um UltronExMsg) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("channel", um.Channel)
	enc.AddString("text", um.Text)
	enc.AddString("payload", um.Payload)
	enc.AddString("title", um.Title)
	return nil
}
