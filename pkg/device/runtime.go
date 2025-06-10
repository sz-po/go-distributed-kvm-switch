package device

import (
	"context"
	"sync"
)

type DeviceName string
type DeviceConfig any

type RuntimeOpt func(*Runtime)

type Runtime struct {
}

func NewRuntime[CONFIG DeviceConfig](deviceName DeviceName, deviceConfig CONFIG, opts ...RuntimeOpt) (*Runtime, error) {
	return &Runtime{}, nil
}

func (runtime *Runtime) Start(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)

	go func() {
		<-ctx.Done()
		wg.Done()
	}()

	return nil
}
