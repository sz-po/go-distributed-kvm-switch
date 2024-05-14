package process

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	runner := NewRunner("/bin/")
	assert.NotNil(t, runner)
}

func TestRunner_GetUptime(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "sleep 10"))

	runner.Start()
	time.Sleep(time.Millisecond * 50)
	assertDurationBetween(t, time.Millisecond*45, time.Millisecond*55, runner.GetUptime())

	time.Sleep(time.Millisecond * 50)
	assertDurationBetween(t, time.Millisecond*95, time.Millisecond*105, runner.GetUptime())

	runner.Stop()
	assertDurationBetween(t, time.Millisecond*95, time.Millisecond*105, runner.GetUptime())

	time.Sleep(time.Millisecond * 50)
	assertDurationBetween(t, time.Millisecond*95, time.Millisecond*105, runner.GetUptime())
}

func TestRunner_Start(t *testing.T) {
	runner := NewRunner("/bin/bash-not-exists")

	err := runner.Start()
	assert.ErrorContains(t, err, "no such file or directory")

	runner = NewRunner("/bin/bash", WithArgs("-c", "sleep 10"))

	err = runner.Start()
	assert.NoError(t, err)

	err = runner.Start()
	assert.ErrorIs(t, err, ErrProcessIsAlreadyRunning)
}

func TestRunner_Stop(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "sleep 10"))

	runner.Start()

	err := runner.Stop()
	assert.NoError(t, err)

	err = runner.Stop()
	assert.ErrorIs(t, err, ErrProcessIsNotRunning)
}

func TestRunner_IsRunning(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "sleep 10"))

	runner.Start()
	assert.True(t, runner.IsRunning())

	runner.Stop()
	assert.False(t, runner.IsRunning())

	time.Sleep(time.Millisecond * 100)
}

func assertDurationBetween(t *testing.T, lowerLimit time.Duration, upperLimit time.Duration, actual time.Duration) {
	assert.GreaterOrEqual(t, actual, lowerLimit)
	assert.LessOrEqual(t, actual, upperLimit)
}
