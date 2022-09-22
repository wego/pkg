// Package header provides useful fucntions to work with http header
package header

import (
	"net/http"
	"strings"
)

// ClientIP returns the true IP address of client
func ClientIP(req *http.Request) (ip string) {
	if req == nil {
		return
	}

	ip = req.Header.Get(RealIP)
	if len(ip) == 0 {
		ips := req.Header.Get(ForwaredFor)
		ip = strings.TrimSpace(strings.Split(ips, ",")[0])
	}
	return
}
