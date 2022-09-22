package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/pointer"
)

func Test_TimeChanged_BothNotNilAndEquals(t *testing.T) {
	assert := assert.New(t)

	now := CurrentUTCTime()

	changed, value := TimeChanged(pointer.To(now), pointer.To(now))
	assert.False(changed)
	assert.Equal(now, *value)
}

func Test_TimeChanged_BothNotNilAndChanged(t *testing.T) {
	assert := assert.New(t)

	now := CurrentUTCTime()
	newTime := now.Add(1 * time.Second)
	changed, value := TimeChanged(pointer.To(newTime), pointer.To(now))
	assert.True(changed)
	assert.Equal(newTime, *value)
}

func Test_TimeChanged_BothNil(t *testing.T) {
	assert := assert.New(t)
	changed, value := TimeChanged(nil, nil)
	assert.False(changed)
	assert.Nil(value)
}

func Test_TimeChanged_NewNotNilButOldNil(t *testing.T) {
	assert := assert.New(t)

	now := CurrentUTCTime()
	changed, value := TimeChanged(pointer.To(now), nil)
	assert.True(changed)
	assert.Equal(now, *value)
}

func Test_TimeChanged_NewNilButOldNotNil(t *testing.T) {
	assert := assert.New(t)

	now := CurrentUTCTime()
	changed, value := TimeChanged(nil, pointer.To(now))
	assert.False(changed)
	assert.Equal(now, *value)
}
