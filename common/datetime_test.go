package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_TimeChanged_BothNotNilAndEquals(t *testing.T) {
	assert := assert.New(t)

	now := CurrentUTCTime()

	changed, value := TimeChanged(TimeRef(now), TimeRef(now))
	assert.False(changed)
	assert.Equal(now, *value)
}

func Test_TimeChanged_BothNotNilAndChanged(t *testing.T) {
	assert := assert.New(t)

	now := CurrentUTCTime()
	newTime := now.Add(1 * time.Second)
	changed, value := TimeChanged(TimeRef(newTime), TimeRef(now))
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
	changed, value := TimeChanged(TimeRef(now), nil)
	assert.True(changed)
	assert.Equal(now, *value)
}

func Test_TimeChanged_NewNilButOldNotNil(t *testing.T) {
	assert := assert.New(t)

	now := CurrentUTCTime()
	changed, value := TimeChanged(nil, TimeRef(now))
	assert.False(changed)
	assert.Equal(now, *value)
}
