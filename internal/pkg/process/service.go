package process

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"sync"
)

type ServiceConfig struct {
}

type Service struct {
	ctx       context.Context
	wg        *sync.WaitGroup
	config    ServiceConfig
	validator *validator.Validate
	logger    *slog.Logger

	processStore      map[Name]*Process
	processStoreMutex sync.Mutex
}

func NewService(ctx context.Context, wg *sync.WaitGroup, config ServiceConfig) (*Service, error) {
	return &Service{
		ctx:       ctx,
		wg:        wg,
		config:    config,
		validator: validator.New(validator.WithRequiredStructEnabled()),
		logger:    slog.Default().With(slog.String("componentName", "process.Service")),

		processStore:      map[Name]*Process{},
		processStoreMutex: sync.Mutex{},
	}, nil
}

func (s *Service) Create(ctx context.Context, name Name, specification Specification) (*Process, error) {
	s.processStoreMutex.Lock()
	defer s.processStoreMutex.Unlock()

	if err := s.validator.Struct(specification); err != nil {
		return nil, fmt.Errorf("failed to validate specification: %w", err)
	}

	if _, exists := s.processStore[name]; exists {
		return nil, ErrProcessNameAlreadyTaken
	}

	processLogger := slog.Default().With(
		slog.String("processName", string(name)),
	)

	process, err := NewProcess(specification, WithLogger(processLogger))
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	s.processStore[name] = process
	return process, nil
}

func (s *Service) Delete(ctx context.Context, name Name) error {
	panic("not implemented")
}

func (s *Service) Get(name Name) (*Process, error) {
	s.processStoreMutex.Lock()
	defer s.processStoreMutex.Unlock()

	if _, exists := s.processStore[name]; !exists {
		return nil, ErrProcessNotFound
	}

	return s.processStore[name], nil
}

func (s *Service) Find() map[Name]*Process {
	s.processStoreMutex.Lock()
	defer s.processStoreMutex.Unlock()

	processes := map[Name]*Process{}

	for processName, process := range s.processStore {
		processes[processName] = process
	}

	return processes
}
