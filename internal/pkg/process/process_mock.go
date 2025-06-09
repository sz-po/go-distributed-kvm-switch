package process

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type ProcessMock struct {
	mock.Mock
}

func (process *ProcessMock) GetStatus() Status {
	args := process.Called()

	return args.Get(0).(Status)
}

func (process *ProcessMock) GetSpecification() Specification {
	args := process.Called()

	return args.Get(0).(Specification)
}

func (process *ProcessMock) Enable(ctx context.Context) error {
	args := process.Called(ctx)

	return args.Error(0)
}

func (process *ProcessMock) Disable(ctx context.Context) error {
	args := process.Called(ctx)

	return args.Error(0)
}

func (process *ProcessMock) Restart(ctx context.Context) error {
	args := process.Called(ctx)

	return args.Error(0)
}

func (process *ProcessMock) Wait(ctx context.Context, phase Phase) error {
	args := process.Called(ctx, phase)

	return args.Error(0)
}
