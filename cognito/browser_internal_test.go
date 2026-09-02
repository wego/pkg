package cognito

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrowserCommand(t *testing.T) {
	const authorizeURL = "https://cognito.test/oauth2/authorize?client_id=x"

	tests := []struct {
		name      string
		givenGOOS string
		wantName  string
		wantArgs  []string
	}{
		{
			name:      "darwin uses open",
			givenGOOS: "darwin",
			wantName:  "open",
			wantArgs:  []string{authorizeURL},
		},
		{
			name:      "windows goes through the protocol handler",
			givenGOOS: "windows",
			wantName:  "rundll32",
			wantArgs:  []string{"url.dll,FileProtocolHandler", authorizeURL},
		},
		{
			name:      "other platforms fall back to xdg-open",
			givenGOOS: "linux",
			wantName:  "xdg-open",
			wantArgs:  []string{authorizeURL},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotArgs := browserCommand(tt.givenGOOS, authorizeURL)

			assert.Equal(t, tt.wantName, gotName)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}
