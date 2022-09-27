package integration_test

import (
	"os"
	"strconv"
)

// TestEnabled determines if integration testing is enabled by checking env variable flag 'ENABLE_INTEGRATION_TEST'. Default is false.
func TestEnabled() bool {
	enabled, err := strconv.ParseBool(os.Getenv("ENABLE_INTEGRATION_TEST"))
	if err != nil {
		enabled = false
	}

	return enabled
}
