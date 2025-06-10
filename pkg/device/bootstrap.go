package device

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
)

const DeviceNameEnvironmentKey = "DKVMS_DEVICE_NAME"

type BootstrapOpt func(*BootstrapConfig)

type BootstrapConfig struct {
	runtimeOptions []RuntimeOpt
}

func WithRuntimeOpt(opt RuntimeOpt) BootstrapOpt {
	return func(config *BootstrapConfig) {
		config.runtimeOptions = append(config.runtimeOptions, opt)
	}
}

func Bootstrap[CONFIG DeviceConfig](opts ...BootstrapOpt) {
	bootstrapConfig := &BootstrapConfig{}

	deviceName, err := readDeviceName()
	if err != nil {
		panic(fmt.Errorf("failed to read device name: %w", err))
	}

	deviceConfig, err := readDeviceConfig[CONFIG]()
	if err != nil {
		panic(fmt.Errorf("failed to read device config: %w", err))
	}

	wg := &sync.WaitGroup{}

	ctx, stopRuntime := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopRuntime()

	runtime, err := NewRuntime(*deviceName, deviceConfig, bootstrapConfig.runtimeOptions...)
	if err != nil {
		panic(fmt.Errorf("failed to create runtime: %w", err))
	}

	err = runtime.Start(ctx, wg)
	if err != nil {
		panic(fmt.Errorf("failed to start runtime: %w", err))
	}

	wg.Wait()
}

func readDeviceName() (*DeviceName, error) {
	deviceName := DeviceName(os.Getenv(DeviceNameEnvironmentKey))
	if len(deviceName) == 0 {
		return nil, ErrMissingDeviceName
	}
	return &deviceName, nil
}

func readDeviceConfig[CONFIG DeviceConfig]() (*CONFIG, error) {
	panic("not implemented")
}

var ErrMissingDeviceName = errors.New("missing device name")

func NewBootstrap[CONFIG DeviceConfig](opts ...BootstrapOpt) (*BootstrapConfig, error) {
	return &BootstrapConfig{}, nil
}
