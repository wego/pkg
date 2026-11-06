package logger_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/logger"
)

func Test_JSON_InvalidInput_DoesNotEchoInput(t *testing.T) {
	const secret = "4111111111111111"
	tests := []struct {
		name  string
		given string
	}{
		{name: "syntax error before the secret", given: `{"amount":1x,"number":"` + secret + `"}`},
		{name: "unexpected value quotes the rest", given: `{"amount":tru,"number":"` + secret + `"}`},
		{name: "trailing data after a valid document", given: `{"a":1} {"number":"` + secret + `"}`},
		{name: "long body after the syntax error", given: `{"amount":1x,"padding":"` + strings.Repeat("a", 200) + `","number":"` + secret + `"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := logger.RedactJSON(tt.given, "", [][]string{{"number"}})
			masked := logger.MaskJSON(tt.given, "", []logger.MaskData{{JSONKeys: []string{"number"}}})
			chained := logger.MaskJSON(redacted, "", []logger.MaskData{{JSONKeys: []string{"number"}}})

			for _, output := range []string{redacted, masked, chained} {
				assert.NotContains(t, output, secret)
				assert.Equal(t, "cannot parse JSON", output)
			}
		})
	}
}
