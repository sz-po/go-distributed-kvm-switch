package process

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type LocalServiceOpt func(*LocalService)

type LocalServiceProcessInstanceFactory func(name Name, specification Specification, opts ...ProcessOpt) (Process, error)

type LocalService struct {
	logger *slog.Logger

	processStore      map[Name]Process
	processStoreMutex sync.Mutex

	createProcessInstance LocalServiceProcessInstanceFactory
}

var _ Service = (*LocalService)(nil)

func WithLocalServiceProcessInstanceFactory(factory LocalServiceProcessInstanceFactory) LocalServiceOpt {
	return func(service *LocalService) {
		service.createProcessInstance = factory
	}
}

func NewLocalService(opts ...LocalServiceOpt) *LocalService {
	service := &LocalService{
		logger: slog.Default().With(slog.String("componentName", "process.LocalService")),

		processStore:      map[Name]Process{},
		processStoreMutex: sync.Mutex{},
	}

	service.createProcessInstance = service.createLocalProcessInstance

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func (service *LocalService) CreateProcess(ctx context.Context, name Name, specification Specification, opts ...ProcessOpt) (Process, error) {
	service.processStoreMutex.Lock()
	defer service.processStoreMutex.Unlock()

	if _, exists := service.processStore[name]; exists {
		return nil, ErrProcessNameAlreadyTaken
	}

	process, err := service.createProcessInstance(name, specification, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

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

// createLocalProcessInstance is factory method for creating LocalProcess instances.
func (service *LocalService) createLocalProcessInstance(name Name, specification Specification, opts ...ProcessOpt) (Process, error) {
	var processOpts []LocalProcessOpt

	for _, opt := range opts {
		processOpt, ok := opt.(LocalProcessOpt)
		if !ok {
			return nil, fmt.Errorf("%w: %T", ErrInvalidLocalProcessOpt, opt)
		}

		processOpts = append(processOpts, processOpt)
	}

	process, err := NewLocalProcess(name, specification, processOpts...)
	if err != nil {
		return nil, err
	}

	return process, nil
}
