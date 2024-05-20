package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTimestamp_IsEmpty(t *testing.T) {
	timestamp := NewTimestamp(time.Time{})
	assert.False(t, timestamp.IsEmpty())

	timestamp = Timestamp("")
	assert.True(t, timestamp.IsEmpty())
}

func TestTimestamp_Time(t *testing.T) {
	now := time.Now()
	timestamp := NewTimestamp(now)
	assert.NotNil(t, timestamp.Time())
	assert.Equal(t, now.Format(TimestampFormat), timestamp.Time().Format(TimestampFormat))

	timestamp = Timestamp("")
	assert.Nil(t, timestamp.Time())
}
