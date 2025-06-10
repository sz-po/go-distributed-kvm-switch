package process

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"sync"
)

type LocalServiceConfig struct {
}

type LocalService struct {
	config LocalServiceConfig
	logger *slog.Logger

	processStore      map[Name]*LocalProcess
	processStoreMutex sync.Mutex

	wg *sync.WaitGroup
}

var _ Service = (*LocalService)(nil)

func NewLocalService(config LocalServiceConfig) (*LocalService, error) {
	validator := validator.New(validator.WithRequiredStructEnabled())

	service := &LocalService{
		config: config,
		logger: slog.Default().With(slog.String("componentName", "process.LocalService")),

		processStore:      map[Name]*LocalProcess{},
		processStoreMutex: sync.Mutex{},
	}

	if validator.Struct(config) != nil {
		return nil, fmt.Errorf("failed to validate service config: %w", validator.Struct(config))
	}

	return service, nil
}

func (service *LocalService) StartService(ctx context.Context, wg *sync.WaitGroup) error {
	service.wg = wg

	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()

		service.processStoreMutex.Lock()
		defer service.processStoreMutex.Unlock()

		for _, process := range service.processStore {
			go func() {
				if process.GetStatus().Enabled {
					if err := process.Disable(ctx); err != nil {
						service.logger.Error("Failed to disable process.", err)
					}
				}
				if err := process.Wait(context.Background(), Idle); err != nil {
					service.logger.Error("Failed to wait for process to exit.", err)
				}
				wg.Done()
			}()
		}
	}()

	return nil
}

func (service *LocalService) CreateProcess(ctx context.Context, name Name, specification Specification) (Process, error) {
	if service.wg == nil {
		return nil, ErrLocalServiceNotStarted
	}

	service.processStoreMutex.Lock()
	defer service.processStoreMutex.Unlock()

	if _, exists := service.processStore[name]; exists {
		return nil, ErrProcessNameAlreadyTaken
	}

	processLogger := slog.Default().With(
		slog.String("processName", string(name)),
	)

	process, err := NewProcess(specification, WithLogger(processLogger))
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	service.wg.Add(1)

	service.processStore[name] = process
	return process, nil
}

func (service *LocalService) DeleteProcess(ctx context.Context, name Name) error {
	service.processStoreMutex.Lock()
	defer service.processStoreMutex.Unlock()

	if process, exists := service.processStore[name]; !exists {
		return ErrProcessNotFound
	} else {
		if process.GetStatus().Enabled {
			if err := process.Disable(ctx); err != nil {
				return fmt.Errorf("failed to disable process: %w", err)
			}
		}

		if err := process.Wait(ctx, Idle); err != nil {
			return fmt.Errorf("failed to wait for process to exit: %w", err)
		}
	}

	delete(service.processStore, name)

	return nil
}

func (service *LocalService) GetProcessByName(name Name) (Process, error) {
	service.processStoreMutex.Lock()
	defer service.processStoreMutex.Unlock()

	if _, exists := service.processStore[name]; !exists {
		return nil, ErrProcessNotFound
	}

	return service.processStore[name], nil
}

func (service *LocalService) FindProcess() map[Name]Process {
	service.processStoreMutex.Lock()
	defer service.processStoreMutex.Unlock()

	processes := map[Name]Process{}

	for processName, process := range service.processStore {
		processes[processName] = process
	}

	return processes
}

var ErrLocalServiceNotStarted = fmt.Errorf("local service not started")
