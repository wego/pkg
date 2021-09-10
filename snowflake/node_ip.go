package snowflake

import "github.com/wego/pkg/host"

func privateIP() (uint16, error) {
	return host.Lower16BitPrivateIP()
}
