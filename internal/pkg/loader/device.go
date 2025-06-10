package loader

import (
	"context"
	"fmt"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api/utils"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/device"
	"log/slog"
	"sync"
	"time"
)

type DeviceLoaderConfig struct {
	ScanInterval utils.Duration `kong:"default=5s"`
}

type DeviceSource interface {
	GetAvailableDevices() (map[device.Name]device.Specification, error)
}

type DeviceLoader struct {
	service device.Service
	source  DeviceSource
	config  DeviceLoaderConfig
	logger  *slog.Logger
}

func NewDeviceLoader(config DeviceLoaderConfig, deviceService device.Service, source DeviceSource) *DeviceLoader {
	return &DeviceLoader{
		service: deviceService,
		source:  source,
		config:  config,
		logger:  slog.Default().With(slog.String("componentName", "loader.DeviceLoader")),
	}
}

func (loader *DeviceLoader) Start(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	ticker := time.NewTicker(time.Duration(loader.config.ScanInterval))

	tick := func() error {
		if err := loader.scan(ctx); err != nil {
			loader.logger.Error("Failed to scan for devices.", slog.String("error", err.Error()))
			return fmt.Errorf("failed to scan for devices: %w", err)
		}

		return nil
	}

	if err := tick(); err != nil {
		return err
	}

	go func() {
		defer wg.Done()
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tick()
			}
		}
	}()

	loader.logger.Info("Device loader started.", slog.Duration("scanInterval", time.Duration(loader.config.ScanInterval)))

	return nil
}

func (loader *DeviceLoader) scan(ctx context.Context) error {
	devices, err := loader.source.GetAvailableDevices()
	if err != nil {
		return fmt.Errorf("failed to get available devices: %w", err)
	}

	for deviceName, deviceSpecification := range devices {
		if loader.service.HasDevice(deviceName) {
			continue
		}

		loader.logger.Info("Found a new device. Loading it.", slog.String("deviceName", string(deviceName)))

		newDevice, err := loader.service.CreateDevice(ctx, deviceName, deviceSpecification)
		if err != nil {
			return fmt.Errorf("failed to create device: %w", err)
		}

		loader.logger.Info("Device loaded.", slog.String("deviceName", string(deviceName)), slog.String("deviceKind", newDevice.GetSpecification().Kind))
	}

	return nil
}
