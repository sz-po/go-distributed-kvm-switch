package device

import (
	"context"
	"fmt"
	"github.com/sz-po/go-distributed-kvm-switch/pkg/device/protocol"
	"os"
	"sync"
)

type DeviceName string
type DeviceConfig any

type RuntimeOpt func(*RuntimeOptions)

type RuntimeOptions struct {
	serviceProviders []protocol.ServiceProvider
}

type Runtime struct {
	serviceRegistry *protocol.ServiceRegistry
	protocolServer  *protocol.Server
	protocolClient  *protocol.Client
}

func WithServiceProvider(serviceProvider protocol.ServiceProvider) RuntimeOpt {
	return func(runtimeOptions *RuntimeOptions) {
		runtimeOptions.serviceProviders = append(runtimeOptions.serviceProviders, serviceProvider)
	}
}

func NewRuntime[CONFIG DeviceConfig](deviceName DeviceName, deviceConfig CONFIG, opts ...RuntimeOpt) (*Runtime, error) {
	runtimeOptions := &RuntimeOptions{}

	for _, opt := range opts {
		opt(runtimeOptions)
	}

	var serviceRegistryOpts []protocol.ServiceRegistryOpt

	for _, serviceProvider := range runtimeOptions.serviceProviders {
		serviceRegistryOpts = append(serviceRegistryOpts, protocol.WithServiceProvider(serviceProvider))
	}

	serviceRegistry := protocol.NewServiceRegistry(serviceRegistryOpts...)

	stdioMux := NewStdioMux(os.Stdin, os.Stdout)

	protocolServer := protocol.NewServer(stdioMux.ServerPipe(), serviceRegistry)
	protocolClient := protocol.NewClient(stdioMux.ClientPipe())

	runtime := &Runtime{
		serviceRegistry: serviceRegistry,
		protocolServer:  protocolServer,
		protocolClient:  protocolClient,
	}

	return runtime, nil
}

func (runtime *Runtime) Start(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)

	if err := runtime.protocolServer.Start(ctx, wg); err != nil {
		return fmt.Errorf("failed to start protocol server: %w", err)
	}

	if err := runtime.protocolClient.Start(ctx, wg); err != nil {
		return fmt.Errorf("failed to start protocol client: %w", err)
	}

	go func() {
		<-ctx.Done()
		wg.Done()
	}()

	return nil
}
