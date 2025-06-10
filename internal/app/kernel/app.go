package kernel

import (
	"context"
	"fmt"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api/http"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/device"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/loader"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/loader/source"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
	"sync"
)

func Start(ctx context.Context, wg *sync.WaitGroup, config Config) error {
	processService, err := process.NewLocalService(config.Service.Process)
	if err != nil {
		return fmt.Errorf("failed to create process service: %w", err)
	}

	if err = processService.StartService(ctx, wg); err != nil {
		return fmt.Errorf("failed to start process service: %w", err)
	}

	deviceService, err := device.NewLocalService(config.Service.Device, processService)
	if err != nil {
		return fmt.Errorf("failed to create device service: %w", err)
	}

	directoryConfigSource, err := source.NewDirectorySource(config.Loader.DirectorySource)
	if err != nil {
		return fmt.Errorf("failed to create directory config source: %w", err)
	}

	deviceLoader := loader.NewDeviceLoader(config.Loader.DeviceLoader, deviceService, directoryConfigSource)
	err = deviceLoader.Start(ctx, wg)
	if err != nil {
		return fmt.Errorf("failed to start device loader: %w", err)
	}

	httpApiRouter := http.NewRouter(processService)

	httpApi := http.NewServer(config.Api.Http,
		http.WithHandler("/api", httpApiRouter),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP API server: %w", err)
	}

	if err := httpApi.Start(ctx, wg); err != nil {
		return fmt.Errorf("failed to start HTTP API server: %w", err)
	}

	return nil
}
