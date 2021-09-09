package host_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/host"
	"testing"
)

func Test_PrivateIPv4(t *testing.T) {
	assert := assert.New(t)
	ip, err :=host.PrivateIPv4()
	assert.NoError(err)
	assert.NotNil(ip)

	lsb, err := host.Lower16BitPrivateIP()
	assert.NoError(err)
	assert.GreaterOrEqual(lsb, uint16(0))
}
