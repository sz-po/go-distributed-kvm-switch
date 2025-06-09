package process

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (service *ServiceMock) CreateProcess(ctx context.Context, name Name, specification Specification) (Process, error) {
	args := service.Called(ctx, name, specification)

	process, _ := args.Get(0).(Process)
	err := args.Error(1)

	return process, err
}

func (service *ServiceMock) DeleteProcess(ctx context.Context, name Name) error {
	args := service.Called(ctx, name)

	return args.Error(0)
}

func (service *ServiceMock) GetProcessByName(name Name) (Process, error) {
	args := service.Called(name)

	process, _ := args.Get(0).(Process)
	err := args.Error(1)

	return process, err
}

func (service *ServiceMock) FindProcess() map[Name]Process {
	args := service.Called()

	return args.Get(0).(map[Name]Process)
}
