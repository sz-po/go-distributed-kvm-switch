package process

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestNewLocalService(t *testing.T) {
	service, err := NewLocalService(LocalServiceConfig{})
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestLocalService_Start(t *testing.T) {
	service, err := NewLocalService(LocalServiceConfig{})
	assert.NoError(t, err)
	assert.NotNil(t, service)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = service.Start(ctx, wg)
	assert.NoError(t, err)
}

func TestLocalService_CreateProcess(t *testing.T) {
	service, err := NewLocalService(LocalServiceConfig{})
	assert.NoError(t, err)
	assert.NotNil(t, service)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	processName := Name("foo")
	processSpecification := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
		AutoEnable:     true,
	}

	process, err := service.CreateProcess(ctx, processName, processSpecification)
	assert.Nil(t, process)
	assert.ErrorIs(t, err, ErrLocalServiceNotStarted)

	err = service.Start(ctx, wg)
	assert.NoError(t, err)

	process, err = service.CreateProcess(ctx, processName, processSpecification)
	assert.NoError(t, err)
	assert.NotNil(t, process)

	waitCtx, waitCancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer waitCancel()
	err = process.Wait(waitCtx, Running)
	assert.NoError(t, err)

	wgCh := make(chan struct{})
	go func() {
		wg.Wait()
		wgCh <- struct{}{}
	}()

	cancel()

	select {
	case <-wgCh:
		assert.Equal(t, process.GetStatus().Enabled, false)
		assert.Equal(t, process.GetStatus().Phase, Idle)
	case <-time.After(500 * time.Millisecond):
		t.Fail()
	}
}

func TestLocalService_DeleteProcess(t *testing.T) {
	service, err := NewLocalService(LocalServiceConfig{})
	assert.NoError(t, err)
	assert.NotNil(t, service)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	processName := Name("foo")
	processSpecification := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
		AutoEnable:     true,
	}

	err = service.Start(ctx, wg)
	assert.NoError(t, err)

	process, err := service.CreateProcess(ctx, processName, processSpecification)
	assert.NoError(t, err)
	assert.NotNil(t, process)

	waitCtx, waitCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer waitCancel()
	err = process.Wait(waitCtx, Running)
	assert.NoError(t, err)

	err = service.DeleteProcess(ctx, Name("not-exists"))
	assert.ErrorIs(t, err, ErrProcessNotFound)

	err = service.DeleteProcess(ctx, processName)
	assert.NoError(t, err)

	assert.Equal(t, process.GetStatus().Enabled, false)
	assert.Equal(t, process.GetStatus().Phase, Idle)

	err = service.DeleteProcess(ctx, processName)
	assert.ErrorIs(t, err, ErrProcessNotFound)
}

func TestLocalService_GetProcessByName(t *testing.T) {
	service, err := NewLocalService(LocalServiceConfig{})
	assert.NoError(t, err)
	assert.NotNil(t, service)

	processName := Name("foo")
	processSpecification := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
		AutoEnable:     true,
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = service.Start(ctx, wg)
	assert.NoError(t, err)

	process, err := service.CreateProcess(ctx, processName, processSpecification)
	assert.NoError(t, err)
	assert.NotNil(t, process)

	getProcessResult, err := service.GetProcessByName(processName)
	assert.NoError(t, err)
	assert.NotNil(t, getProcessResult)

	assert.Equal(t, process.GetSpecification(), getProcessResult.GetSpecification())

	err = service.DeleteProcess(ctx, processName)
	assert.NoError(t, err)

	getProcessResult, err = service.GetProcessByName(processName)
	assert.ErrorIs(t, err, ErrProcessNotFound)
	assert.Nil(t, getProcessResult)
}

func TestLocalService_FindProcess(t *testing.T) {
	service, err := NewLocalService(LocalServiceConfig{})
	assert.NoError(t, err)
	assert.NotNil(t, service)

	processName := Name("foo")
	processSpecification := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
		AutoEnable:     true,
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = service.Start(ctx, wg)
	assert.NoError(t, err)

	process, err := service.CreateProcess(ctx, processName, processSpecification)
	assert.NoError(t, err)
	assert.NotNil(t, process)

	foundProcesses := service.FindProcess()
	assert.Len(t, foundProcesses, 1)
	assert.Contains(t, foundProcesses, processName)
}
