package device

import (
	"context"
	"errors"
	"github.com/coder/quartz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
	device_sdk "github.com/sz-po/go-distributed-kvm-switch/pkg/device"
	"testing"
	"time"
)

func TestNewLocalDevice(t *testing.T) {
	processFactoryErr := errors.New("process factory error")

	processFactory, processFactoryMock := process.CreateProcessFactoryMock(t)
	processFactoryMock.On("func1", process.Name("device-bar-foo"), mock.Anything, mock.Anything).Return(nil, processFactoryErr).Once()
	processFactoryMock.On("func1", process.Name("device-bar-foo"), mock.Anything, mock.Anything).Return(&process.ProcessMock{}, nil)

	clock := quartz.NewMock(t)

	device, err := NewLocalDevice(Name("foo"), Specification{
		Kind: "bar",
		Config: map[string]any{
			"foo": "bar",
		},
	}, WithLocalDeviceClock(clock), WithLocalDeviceProcessFactory(processFactory))
	assert.ErrorIs(t, err, processFactoryErr)
	assert.Nil(t, device)

	device, err = NewLocalDevice(Name("foo"), Specification{
		Kind: "bar",
		Config: map[string]any{
			"foo": "bar",
		},
	}, WithLocalDeviceClock(clock), WithLocalDeviceProcessFactory(processFactory), WithLocalDeviceProcessExecutableDirectory("/test"))
	assert.NoError(t, err)
	assert.NotNil(t, device)

	processFactoryMock.AssertCalled(t, "func1", process.Name("device-bar-foo"), process.Specification{
		ExecutablePath: "/test/dkvms-device-bar",
		EnvironmentVariables: map[string]string{
			device_sdk.DeviceNameEnvironmentKey: "foo",
		},
		RestartMode: process.Always,
		AutoEnable:  false,
	}, mock.Anything)
}

func TestLocalDevice_GetSpecification(t *testing.T) {
	processFactory := func(name process.Name, specification process.Specification, opts ...process.ProcessOpt) (process.Process, error) {
		return &process.ProcessMock{}, nil
	}

	specification := Specification{
		Kind: "foo",
		Config: map[string]any{
			"foo": "bar",
		},
	}

	device, err := NewLocalDevice(Name("foo"), specification, WithLocalDeviceProcessFactory(processFactory))
	assert.NoError(t, err)
	assert.Equal(t, specification, device.GetSpecification())
}

func TestLocalDevice_GetStatus(t *testing.T) {
	processFactory := func(name process.Name, specification process.Specification, opts ...process.ProcessOpt) (process.Process, error) {
		return &process.ProcessMock{}, nil
	}

	clock := quartz.NewMock(t)

	device, err := NewLocalDevice(Name("foo"), Specification{
		Kind:       "bar",
		AutoEnable: false,
	}, WithLocalDeviceProcessFactory(processFactory), WithLocalDeviceClock(clock))
	assert.NoError(t, err)
	assert.Equal(t, Status{
		ProcessName: "device-bar-foo",
		Phase:       Idle,
		Enabled:     false,
	}, device.GetStatus())

	device, err = NewLocalDevice(Name("foo"), Specification{
		Kind:       "bar",
		AutoEnable: true,
	}, WithLocalDeviceProcessFactory(processFactory), WithLocalDeviceClock(clock))
	assert.NoError(t, err)
	assert.True(t, device.GetStatus().Enabled)
}

func TestLocalDevice_Enable_Disable(t *testing.T) {
	processMock := &process.ProcessMock{}
	processMock.On("Enable", mock.Anything).Return(nil).Once()
	processMock.On("Disable", mock.Anything).Return(nil).Once()
	processMock.On("GetStatus").Return(process.Status{
		Enabled: false,
		Phase:   process.Idle,
	}).Once()
	processMock.On("GetStatus").Return(process.Status{
		Enabled: true,
		Phase:   process.Starting,
	}).Once()
	processMock.On("GetStatus").Return(process.Status{
		Enabled: true,
		Phase:   process.Running,
	}).Once()
	processMock.On("GetStatus").Return(process.Status{
		Enabled: true,
		Phase:   process.Running,
	}).Once()
	processMock.On("GetStatus").Return(process.Status{
		Enabled: false,
		Phase:   process.Terminating,
	}).Once()
	processMock.On("GetStatus").Return(process.Status{
		Enabled: false,
		Phase:   process.Idle,
	})

	processFactory := func(name process.Name, specification process.Specification, opts ...process.ProcessOpt) (process.Process, error) {
		return processMock, nil
	}

	clock := quartz.NewMock(t)

	ctx := context.Background()

	device, err := NewLocalDevice(Name("foo"), Specification{
		Kind:       "bar",
		AutoEnable: false,
	}, WithLocalDeviceProcessFactory(processFactory), WithLocalDeviceClock(clock))
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.False(t, device.GetStatus().Enabled)

	err = device.Enable(ctx)
	assert.NoError(t, err)
	assert.True(t, device.GetStatus().Enabled)

	clock.AdvanceNext()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, Starting, device.GetStatus().Phase)

	clock.AdvanceNext()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, Starting, device.GetStatus().Phase)

	clock.AdvanceNext()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, Running, device.GetStatus().Phase)

	err = device.Enable(ctx)
	assert.ErrorIs(t, err, ErrDeviceAlreadyEnabled)
	assert.True(t, device.GetStatus().Enabled)

	err = device.Disable(ctx)
	assert.NoError(t, err)
	assert.False(t, device.GetStatus().Enabled)
	clock.AdvanceNext()
	time.Sleep(10 * time.Millisecond)

	clock.AdvanceNext()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, Stopping, device.GetStatus().Phase)

	clock.AdvanceNext()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, Idle, device.GetStatus().Phase)

	err = device.Disable(ctx)
	assert.ErrorIs(t, err, ErrDeviceAlreadyDisabled)
	assert.False(t, device.GetStatus().Enabled)
}
