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
func (m UltronExMsg) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	title := "\n>" + m.Title + "\n"
	payload := "```" + m.Payload + "```"

	enc.AddString("channel", m.Channel)
	enc.AddString("text", m.Text+title+payload)
	return nil
}
