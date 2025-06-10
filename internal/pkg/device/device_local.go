package device

import (
	"context"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
	"log/slog"
	"sync"
)

type LocalDeviceOpt func(*LocalDevice)

type LocalDevice struct {
	specification Specification
	name          Name
	status        Status
	statusMutex   *sync.Mutex

	processExecutablePath string
	processInstance       process.Process
	processName           process.Name

	logger *slog.Logger
}

func WithProcessExecutablePath(path string) LocalDeviceOpt {
	return func(device *LocalDevice) {
		device.processExecutablePath = path
	}
}

func NewLocalDevice(name Name, specification Specification, processService process.Service, opts ...LocalDeviceOpt) (*LocalDevice, error) {
	if err := defaults.Set(&specification); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	validator := validator.New(validator.WithRequiredStructEnabled())

	if err := validator.Struct(specification); err != nil {
		return nil, fmt.Errorf("failed to validate specification: %w", err)
	}

	processName := process.Name(fmt.Sprintf("device:%s/%s", specification.Kind, name))

	device := &LocalDevice{
		name:          name,
		specification: specification,
		status: Status{
			ProcessName: processName,
			Enabled:     specification.AutoEnable,
		},
		statusMutex: &sync.Mutex{},

		processName:           processName,
		processExecutablePath: fmt.Sprintf("/usr/local/bin/%s", specification.Kind),

		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(device)
	}

	processSpecification := process.Specification{
		ExecutablePath: device.processExecutablePath,
		RestartMode:    process.Always,
		AutoEnable:     false,
	}

	processInstance, err := processService.CreateProcess(context.Background(), processName, processSpecification)
	if err != nil {
		return nil, fmt.Errorf("failed to create process: %w", err)
	}
	device.processInstance = processInstance

	device.logger = device.logger.With(slog.String("componentName", "LocalDevice"))

	return device, nil
}

func (device *LocalDevice) GetStatus() Status {
	//TODO implement me
	panic("implement me")
}

func (device *LocalDevice) GetSpecification() Specification {
	return device.specification
}

func (device *LocalDevice) Enable(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (device *LocalDevice) Disable(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (device *LocalDevice) Restart(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
