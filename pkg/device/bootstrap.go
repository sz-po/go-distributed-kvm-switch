package device

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"os"
	"os/signal"
	"sync"
)

const DeviceNameEnvironmentKey = "DKVMS_DEVICE_NAME"
const DeviceConfigEnvironmentKey = "DKVMS_DEVICE_CONFIG"

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
	var config CONFIG

	configBufferEncoded := []byte(os.Getenv(DeviceConfigEnvironmentKey))
	if len(configBufferEncoded) == 0 {
		return nil, ErrMissingDeviceConfig
	}

	var configBuffer []byte
	configBuffer, err := base64.StdEncoding.DecodeString(string(configBufferEncoded))
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	err = json.Unmarshal(configBuffer, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	err = defaults.Set(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	configValidator := validator.New(validator.WithRequiredStructEnabled())

	err = configValidator.Struct(config)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &config, nil
}

var ErrMissingDeviceName = errors.New("missing device name")
var ErrMissingDeviceConfig = errors.New("missing device config")

func NewBootstrap[CONFIG DeviceConfig](opts ...BootstrapOpt) (*BootstrapConfig, error) {
	return &BootstrapConfig{}, nil
}
