package device

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type DeviceMock struct {
	mock.Mock
}

func (device *DeviceMock) GetStatus() Status {
	args := device.Called()

	return args.Get(0).(Status)
}

func (device *DeviceMock) GetSpecification() Specification {
	args := device.Called()

	return args.Get(0).(Specification)
}

func (device *DeviceMock) Enable(ctx context.Context) error {
	args := device.Called(ctx)

	return args.Error(0)
}

func (device *DeviceMock) Disable(ctx context.Context) error {
	args := device.Called(ctx)

	return args.Error(0)
}

func (device *DeviceMock) Restart(ctx context.Context) error {
	args := device.Called(ctx)

	return args.Error(0)
}
