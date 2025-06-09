package kernel

import (
	"context"
	"fmt"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api/http"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/device"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
	"sync"
)

func Start(ctx context.Context, wg *sync.WaitGroup, config Config) error {
	processService, err := process.NewLocalService(config.Service.Process)
	if err != nil {
		return fmt.Errorf("failed to create process service: %w", err)
	}

	if err = processService.Start(ctx, wg); err != nil {
		return fmt.Errorf("failed to start process service: %w", err)
	}

	_, err = device.New(ctx, wg, config.Service.Device, processService)
	if err != nil {
		return fmt.Errorf("failed to create device service: %w", err)
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
