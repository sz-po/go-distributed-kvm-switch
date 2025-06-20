package device

import (
	"context"
	"fmt"
	"github.com/coder/quartz"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
	device_sdk "github.com/sz-po/go-distributed-kvm-switch/pkg/device"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"
)

const ProcessBinaryNamePrefix = "dkvms-device-"

type LocalDeviceOpt func(*LocalDevice)

type LocalDevice struct {
	specification Specification
	name          Name
	status        Status
	statusMutex   *sync.Mutex

	createProcess process.ProcessFactory

	processExecutableDirectory string
	processExecutableName      string
	processInstance            process.Process
	processName                process.Name

	controlLoopInterval time.Duration
	controlLoopReady    *sync.WaitGroup
	controlLoopStop     context.CancelFunc

	clock  quartz.Clock
	logger *slog.Logger
}

func WithLocalDeviceProcessExecutableDirectory(directory string) LocalDeviceOpt {
	return func(device *LocalDevice) {
		device.processExecutableDirectory = directory
	}
}

func WithLocalDeviceProcessFactory(factory process.ProcessFactory) LocalDeviceOpt {
	return func(device *LocalDevice) {
		device.createProcess = factory
	}
}

func WithLocalDeviceLogger(logger *slog.Logger) LocalDeviceOpt {
	return func(device *LocalDevice) {
		device.logger = logger
	}
}

func WithLocalDeviceClock(clock quartz.Clock) LocalDeviceOpt {
	return func(device *LocalDevice) {
		device.clock = clock
	}
}

func NewLocalDevice(name Name, specification Specification, opts ...LocalDeviceOpt) (*LocalDevice, error) {
	if err := defaults.Set(&specification); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	validator := validator.New(validator.WithRequiredStructEnabled())

	if err := validator.Struct(specification); err != nil {
		return nil, fmt.Errorf("failed to validate specification: %w", err)
	}

	processName := process.Name(fmt.Sprintf("device-%s-%s", specification.Kind, name))

	device := &LocalDevice{
		name:          name,
		specification: specification,
		status: Status{
			ProcessName: processName,
			Enabled:     specification.AutoEnable,
			Phase:       Idle,
		},
		statusMutex: &sync.Mutex{},

		processName:                processName,
		processExecutableDirectory: "/usr/local/bin",
		processExecutableName:      fmt.Sprintf("%s%s", ProcessBinaryNamePrefix, specification.Kind),

		controlLoopInterval: 100 * time.Millisecond,
		controlLoopReady:    &sync.WaitGroup{},

		logger: slog.Default(),
		clock:  quartz.NewReal(),
	}

	device.createProcess = process.CreateLocalProcessFactory()

	for _, opt := range opts {
		opt(device)
	}

	processSpecification := process.Specification{
		ExecutablePath: path.Join(device.processExecutableDirectory, device.processExecutableName),
		EnvironmentVariables: map[string]string{
			device_sdk.DeviceNameEnvironmentKey: string(device.name),
		},
		RestartMode: process.Always,
		AutoEnable:  false,
	}

	processInstance, err := device.createProcess(processName, processSpecification,
		process.WithStdin(os.Stdin),
		process.WithStdout(os.Stdout),
		process.WithStderr(os.Stderr),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create process: %w", err)
	}
	device.processInstance = processInstance

	device.logger = device.logger.With(
		slog.String("componentName", "LocalDevice"),
		slog.String("deviceName", string(device.name)),
	)

	ctx := context.Background()

	device.controlLoopReady.Add(1)
	controlLoopCtx, controlLoopStop := context.WithCancel(ctx)
	device.controlLoopStop = controlLoopStop
	go device.controlLoop(controlLoopCtx)
	device.controlLoopReady.Wait()

	return device, nil
}

func (device *LocalDevice) GetStatus() Status {
	device.statusMutex.Lock()
	defer device.statusMutex.Unlock()

	return device.status
}

func (device *LocalDevice) GetSpecification() Specification {
	return device.specification
}

func (device *LocalDevice) Enable(ctx context.Context) error {
	device.statusMutex.Lock()
	defer device.statusMutex.Unlock()

	if device.status.Enabled {
		return ErrDeviceAlreadyEnabled
	}

	device.status.Enabled = true
	return nil
}

func (device *LocalDevice) Disable(ctx context.Context) error {
	device.statusMutex.Lock()
	defer device.statusMutex.Unlock()

	if !device.status.Enabled {
		return ErrDeviceAlreadyDisabled
	}

	device.status.Enabled = false
	return nil
}

func (device *LocalDevice) Restart(ctx context.Context) error {
	device.statusMutex.Lock()
	defer device.statusMutex.Unlock()

	if !device.status.Enabled {
		return ErrDeviceAlreadyDisabled
	}

	err := device.processInstance.Restart(ctx)
	if err != nil {
		return fmt.Errorf("failed to restart process: %w", err)
	}

	return nil
}

func (device *LocalDevice) controlLoop(ctx context.Context) {
	device.logger.Debug("Starting the control loop.")

	ticker := device.clock.NewTicker(device.controlLoopInterval)
	defer ticker.Stop()

	device.controlLoopReady.Done()

	for {
		select {
		case <-ctx.Done():
			device.logger.Debug("The control loop has been stopped.")
			return
		case <-ticker.C:
			if err := device.controlFn(); err != nil {
				device.logger.Warn("Failed to control the device.", slog.String("error", err.Error()))
			}
		}
	}
}

func (device *LocalDevice) controlFn() error {
	device.statusMutex.Lock()
	defer device.statusMutex.Unlock()

	processStatus := device.processInstance.GetStatus()

	if device.status.Enabled {
		switch processStatus.Phase {
		case process.Running:
			device.status.Phase = Running
		default:
			device.status.Phase = Starting
		}

		if !processStatus.Enabled {
			device.logger.Debug("Enabling the device process.")
			return device.processInstance.Enable(context.Background())
		}
	} else {
		switch processStatus.Phase {
		case process.Idle:
			device.status.Phase = Idle
		default:
			device.status.Phase = Stopping
		}

		if processStatus.Enabled {
			device.logger.Debug("Disabling the device process.")
			return device.processInstance.Disable(context.Background())
		}
	}

	return nil
}
