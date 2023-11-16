package host

import (
	"net"

	"github.com/wego/pkg/errors"
)

// PrivateIPv4 get current host private IP v4 address
func PrivateIPv4() (net.IP, error) {
	const op errors.Op = "host.PrivateIP"
	interfaces, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, inf := range interfaces {
		inet, ok := inf.(*net.IPNet)
		if !ok || inet.IP.IsLoopback() {
			continue
		}

		ip := inet.IP
		if isPrivateIPv4(inet.IP) {
			return ip.To4(), nil
		}
	}
	return nil, errors.New(nil, op, "no private ip address")
}

// isPrivateIPv4 check if an ip address private
func isPrivateIPv4(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1]&0xf0 == 16) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}
	return false
}

// Lower16BitPrivateIP get lower 16 bits of current private IP address
func Lower16BitPrivateIP() (uint16, error) {
	ip, err := PrivateIPv4()
	if err != nil {
		return 0, err
	}
	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}
