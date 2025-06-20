package process

import (
	"context"
	"github.com/stretchr/testify/mock"
	"testing"
)

func CreateProcessFactoryMock(t *testing.T) (ProcessFactory, *mock.Mock) {
	mock := &mock.Mock{}

	return func(name Name, specification Specification, opts ...ProcessOpt) (Process, error) {
		args := mock.Called(name, specification, opts)

		process, _ := args.Get(0).(Process)
		err := args.Error(1)

		return process, err
	}, mock
}

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
