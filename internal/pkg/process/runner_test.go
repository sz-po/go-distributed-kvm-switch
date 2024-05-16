package process

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	runner := NewRunner("/bin/bash")
	assert.NotNil(t, runner)
}

func TestRunner_GetUptime(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "sleep 10"))

	runner.Start()
	time.Sleep(time.Millisecond * 50)
	assertDurationBetween(t, time.Millisecond*45, time.Millisecond*55, runner.GetUptime())

	time.Sleep(time.Millisecond * 50)
	assertDurationBetween(t, time.Millisecond*95, time.Millisecond*105, runner.GetUptime())

	runner.Stop(context.Background())
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
	assert.ErrorIs(t, err, ErrProcessIsRunning)
}

func TestRunner_Stop_Interrupt(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "sleep 10"))

	runner.Start()

	timeoutCtx, _ := context.WithTimeout(context.Background(), time.Millisecond*100)
	stopAt := time.Now()
	err := runner.Stop(timeoutCtx)
	assert.NoError(t, err)
	assert.LessOrEqual(t, time.Since(stopAt), time.Millisecond*10)

	err = runner.Stop(context.Background())
	assert.ErrorIs(t, err, ErrProcessIsNotRunning)
}

func TestRunner_Stop_Kill(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "trap 'sleep 10' INT; sleep 10"))

	runner.Start()

	time.Sleep(time.Millisecond * 10)

	timeoutCtx, _ := context.WithTimeout(context.Background(), time.Millisecond*100)
	stopAt := time.Now()
	err := runner.Stop(timeoutCtx)
	assert.NoError(t, err)
	assertDurationBetween(t, time.Millisecond*90, time.Millisecond*110, time.Since(stopAt))

	err = runner.Stop(context.Background())
	assert.ErrorIs(t, err, ErrProcessIsNotRunning)
}

func TestRunner_Stop_AlreadyFinished(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "trap 'sleep 10' INT; sleep 10"))

	runner.Start()
	pid, _ := runner.GetPID()

	err := exec.Command("/bin/bash", "-c", fmt.Sprintf("kill -9 %d", pid)).Run()
	assert.NoError(t, err)

	err = runner.Stop(context.Background())
	assert.ErrorContains(t, err, "already finished")
}

func TestRunner_IsRunning(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "sleep 10"))

	runner.Start()
	assert.True(t, runner.IsRunning())

	runner.Stop(context.Background())
	assert.False(t, runner.IsRunning())
}

func TestRunner_GetExitCode(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "sleep 1 && exit 123"))
	runner.Start()

	exitCode, err := runner.GetExitCode()
	assert.ErrorIs(t, err, ErrProcessIsRunning)
	assert.Equal(t, 0, exitCode)

	runner.Wait()

	exitCode, err = runner.GetExitCode()
	assert.NoError(t, err)
	assert.Equal(t, 123, exitCode)
}

func TestRunner_Wait(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "sleep 1"))

	err := runner.Wait()
	assert.ErrorIs(t, err, ErrProcessIsNotRunning)

	runner.Start()
	err = runner.Wait()
	assert.NoError(t, err)
	assert.False(t, runner.IsRunning())
}

func TestRunner_GetPID(t *testing.T) {
	runner := NewRunner("/bin/bash", WithArgs("-c", "sleep 1"))

	pid, err := runner.GetPID()
	assert.ErrorIs(t, err, ErrProcessIsNotRunning)
	assert.Equal(t, 0, pid)

	runner.Start()
	pid, err = runner.GetPID()
	assert.NoError(t, err)
	assert.True(t, runner.IsRunning())
	assert.NotEqual(t, 0, pid)

	runner.Wait()
	pid, err = runner.GetPID()
	assert.ErrorIs(t, err, ErrProcessIsNotRunning)
	assert.False(t, runner.IsRunning())
	assert.Equal(t, 0, pid)
}

func TestWithStdout(t *testing.T) {
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")

	runner := NewRunner("/bin/bash", WithArgs("-c", "echo hello"), WithStdout(stdout), WithStderr(stderr))
	runner.Start()
	runner.Wait()

	assert.Equal(t, "hello\n", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestWithStderr(t *testing.T) {
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")

	runner := NewRunner("/bin/bash", WithArgs("-c", ">&2 echo hello"), WithStdout(stdout), WithStderr(stderr))
	runner.Start()
	runner.Wait()

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "hello\n", stderr.String())
}

func TestWithStdin(t *testing.T) {
	stdout := bytes.NewBufferString("")
	stdin := bytes.NewBufferString("hello")

	runner := NewRunner("/bin/bash", WithArgs("-c", "read line; echo $line"), WithStdout(stdout), WithStdin(stdin))
	runner.Start()
	runner.Wait()

	assert.Equal(t, "hello\n", stdout.String())
}

func assertDurationBetween(t *testing.T, lowerLimit time.Duration, upperLimit time.Duration, actual time.Duration) {
	assert.GreaterOrEqual(t, actual, lowerLimit)
	assert.LessOrEqual(t, actual, upperLimit)
}
