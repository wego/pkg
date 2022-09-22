package header_test

import (
	"net/http"
	"testing"

	"github.com/wego/pkg/http/header"
)

func Test_ClientIP(t *testing.T) {
	testCases := map[string]struct {
		req *http.Request
		ip  string
	}{
		"nil request": {},
		"real ip": {
			req: &http.Request{
				Header: http.Header{
					header.RealIP:      []string{"1.2.3.4", "5.6.7.8"},
					header.ForwaredFor: []string{"4.3.2.1", "8.7.6.5"},
				},
			},
			ip: "1.2.3.4",
		},
		"forwarded for": {
			req: &http.Request{
				Header: http.Header{
					header.ForwaredFor: []string{" 4.3.2.1 , 1.1.1.1", "8.7.6.5"},
				},
			},
			ip: "4.3.2.1",
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			actual := header.ClientIP(tc.req)
			if actual != tc.ip {
				t.Errorf("expected IP: <%s>, but got: <%s>", tc.ip, actual)
			}
		})
	}
}
