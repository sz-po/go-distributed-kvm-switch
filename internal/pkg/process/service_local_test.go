package process

import (
	"context"
	"errors"
	"github.com/coder/quartz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api/utils"
	"log/slog"
	"testing"
	"time"
)

func TestLocalService_CreateProcess(t *testing.T) {
	clock := quartz.NewMock(t)

	processFactoryErr := errors.New("failed to create process")

	processFactory := func(name Name, specification Specification, opts ...ProcessOpt) (Process, error) {
		if name == Name("foo") {
			return &ProcessMock{}, nil
		} else {
			return nil, processFactoryErr
		}
	}

	service := NewLocalService(WithLocalServiceProcessInstanceFactory(processFactory))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	processName := Name("foo")
	processSpecification := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
		AutoEnable:     true,
	}

	process, err := service.CreateProcess(ctx, processName, processSpecification, WithClock(clock))
	assert.NoError(t, err)
	assert.NotNil(t, process)

	process, err = service.CreateProcess(ctx, processName, processSpecification, WithClock(clock))
	assert.ErrorIs(t, err, ErrProcessNameAlreadyTaken)
	assert.Nil(t, process)

	process, err = service.CreateProcess(ctx, Name("bar"), processSpecification, WithClock(clock))
	assert.ErrorIs(t, err, processFactoryErr)
	assert.Nil(t, process)
}

func TestLocalService_DeleteProcess(t *testing.T) {
	disableErr := errors.New("failed to disable")
	waitErr := errors.New("failed to wait")

	processFactory := func(name Name, specification Specification, opts ...ProcessOpt) (Process, error) {
		processMock := &ProcessMock{}
		processMock.On("Wait", mock.Anything, Running).Return(nil)
		processMock.On("Wait", mock.Anything, Idle).Return(waitErr).Once()
		processMock.On("Wait", mock.Anything, Idle).Return(nil)
		processMock.On("Disable", mock.Anything).Return(disableErr).Once()
		processMock.On("Disable", mock.Anything).Return(nil)
		processMock.On("GetStatus").Return(Status{
			Enabled: true,
			Phase:   Running,
		}).Once()
		processMock.On("GetStatus").Return(Status{
			Enabled: false,
			Phase:   Idle,
		})

		return processMock, nil
	}

	service := NewLocalService(WithLocalServiceProcessInstanceFactory(processFactory))
	assert.NotNil(t, service)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	processName := Name("foo")
	processSpecification := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
		AutoEnable:     true,
	}

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
	assert.ErrorIs(t, err, disableErr)

	err = service.DeleteProcess(ctx, processName)
	assert.ErrorIs(t, err, waitErr)

	err = service.DeleteProcess(ctx, processName)
	assert.NoError(t, err)

	assert.False(t, process.GetStatus().Enabled)
	assert.Equal(t, Idle, process.GetStatus().Phase)

	err = service.DeleteProcess(ctx, processName)
	assert.ErrorIs(t, err, ErrProcessNotFound)
}

func TestLocalService_GetProcessByName(t *testing.T) {
	processName := Name("foo")
	processSpecification := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
		AutoEnable:     false,
	}

	processFactory := func(name Name, specification Specification, opts ...ProcessOpt) (Process, error) {
		processMock := &ProcessMock{}
		processMock.On("Wait", mock.Anything, Idle).Return(nil)
		processMock.On("GetSpecification").Return(processSpecification)
		processMock.On("GetStatus").Return(Status{
			Enabled: false,
			Phase:   Idle,
		})
		return processMock, nil
	}

	service := NewLocalService(WithLocalServiceProcessInstanceFactory(processFactory))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	processFactory := func(name Name, specification Specification, opts ...ProcessOpt) (Process, error) {
		return &ProcessMock{}, nil
	}

	service := NewLocalService(WithLocalServiceProcessInstanceFactory(processFactory))

	processName := Name("foo")
	processSpecification := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
		AutoEnable:     true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	process, err := service.CreateProcess(ctx, processName, processSpecification)
	assert.NoError(t, err)
	assert.NotNil(t, process)

	foundProcesses := service.FindProcess()
	assert.Len(t, foundProcesses, 1)
	assert.Contains(t, foundProcesses, processName)
}

func TestLocalService_createLocalProcessInstance(t *testing.T) {
	service := NewLocalService()

	processSpecification := Specification{
		ExecutablePath:       "/bin/sleep",
		Arguments:            []string{"10"},
		EnvironmentVariables: map[string]string{},
		RestartMode:          OnError,
		KillTimeout:          utils.Duration(time.Second),
	}

	process, err := service.createLocalProcessInstance(Name("foo"), processSpecification, string("foo"))
	assert.ErrorIs(t, err, ErrInvalidLocalProcessOpt)
	assert.Nil(t, process)

	process, err = service.createLocalProcessInstance(Name("foo"), processSpecification, WithLogger(slog.Default()))
	assert.NoError(t, err)
	assert.NotNil(t, process)

	process, err = service.createLocalProcessInstance(Name("foo_invalid_name"), processSpecification)
	assert.ErrorIs(t, err, ErrInvalidProcessName)
	assert.Nil(t, process)

	process, err = service.createLocalProcessInstance(Name("foo"), processSpecification)
	assert.NoError(t, err)
	assert.NotNil(t, process)
	assert.Equal(t, processSpecification, process.GetSpecification())
}
