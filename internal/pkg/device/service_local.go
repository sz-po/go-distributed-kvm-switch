package device

import (
	"context"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
	"sync"
)

type LocalServiceConfig struct {
	ExecutableDirectory string `kong:"default=/usr/local/bin"`
}

type LocalService struct {
	config LocalServiceConfig

	devices      map[Name]Device
	devicesMutex *sync.Mutex

	processService process.Service
}

var _ Service = (*LocalService)(nil)

func NewLocalService(config LocalServiceConfig, processService process.Service) (*LocalService, error) {
	return &LocalService{
		config:         config,
		devices:        make(map[Name]Device),
		devicesMutex:   &sync.Mutex{},
		processService: processService,
	}, nil
}

func (service *LocalService) CreateDevice(ctx context.Context, name Name, specification Specification) (Device, error) {
	service.devicesMutex.Lock()
	defer service.devicesMutex.Unlock()

	if _, found := service.devices[name]; found {
		return nil, ErrDeviceNameAlreadyTaken
	}

	device, err := NewLocalDevice(name, specification,
		WithLocalDeviceProcessFactory(process.CreateServiceProcessFactory(service.processService)),
		WithLocalDeviceProcessExecutableDirectory(service.config.ExecutableDirectory),
	)
	if err != nil {
		return nil, err
	}

	service.devices[name] = device

	return device, nil
}

func (service *LocalService) DeleteDevice(ctx context.Context, name Name) error {
	//TODO implement me
	panic("implement me")
}

func (service *LocalService) GetDeviceByName(name Name) (Device, error) {
	service.devicesMutex.Lock()
	defer service.devicesMutex.Unlock()

	if device, found := service.devices[name]; found {
		return device, nil
	}

	return nil, ErrDeviceNotFound
}

func (service *LocalService) HasDevice(name Name) bool {
	service.devicesMutex.Lock()
	defer service.devicesMutex.Unlock()

	_, found := service.devices[name]
	return found
}
