package device

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (service *ServiceMock) CreateDevice(ctx context.Context, name Name, specification Specification) (Device, error) {
	args := service.Called(ctx, name, specification)

	device, _ := args.Get(0).(Device)
	err := args.Error(1)

	return device, err
}

func (service *ServiceMock) DeleteDevice(ctx context.Context, name Name) error {
	args := service.Called(ctx, name)

	return args.Error(0)
}

func (service *ServiceMock) GetDeviceByName(name Name) (Device, error) {
	args := service.Called(name)

	device, _ := args.Get(0).(Device)
	err := args.Error(1)

	return device, err
}

func (service *ServiceMock) HasDevice(name Name) bool {
	args := service.Called(name)

	return args.Bool(0)
}
