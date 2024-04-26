package kernel

import (
	"context"
	"fmt"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/device"
)

func Start(ctx context.Context) error {
	deviceManager := device.NewManager()

	if err := deviceManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start device manager: %w", err)
	}

	return nil
}
