package process

import (
	"github.com/stretchr/testify/assert"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api"
	"testing"
	"time"
)

func TestController_InitInstance(t *testing.T) {
	controller := NewController()

	object := api.Object[Specification, Status]{
		Metadata: api.Metadata{
			Name: "test",
		},
		Specification: Specification{
			Execution: ExecutionSpecification{
				ExecutablePath: "/bin/bash",
			},
		},
		Status: nil,
	}

	runner, err := controller.InitInstance(&object)
	assert.NotNil(t, runner)
	assert.NoError(t, err)
}

func TestController_ReconcileInstance(t *testing.T) {
	controller := NewController()

	object := &api.Object[Specification, Status]{
		Metadata: api.Metadata{
			Name: "test",
		},
		Specification: Specification{
			Execution: ExecutionSpecification{
				ExecutablePath: "/bin/bash",
				Arguments:      []string{"-c", "sleep 1"},
			},
		},
		Status: nil,
	}

	runner, err := controller.InitInstance(object)
	assert.NoError(t, err)
	assert.NotNil(t, runner)

	status, err := controller.ReconcileInstance(object, runner)
	assert.NoError(t, err)
	assert.NotNil(t, status)

	assert.False(t, status.IsRunning)
	assert.Zero(t, status.ProcessID)
	assert.Zero(t, status.ExitCode)
	assert.Empty(t, status.Error)

	status, err = controller.ReconcileInstance(object, runner)
	object.Status = status
	assert.NoError(t, err)
	assert.NotNil(t, status)

	assert.True(t, status.IsRunning)
	assert.NotZero(t, status.ProcessID)
	assert.Zero(t, status.ExitCode)
	assert.Empty(t, status.Error)

	time.Sleep(2 * time.Second)

	status, err = controller.ReconcileInstance(object, runner)
	object.Status = status
	assert.NoError(t, err)
	assert.NotNil(t, status)

	assert.False(t, status.IsRunning)
	assert.Zero(t, status.ProcessID)
	assert.Zero(t, status.ExitCode)
	assert.Empty(t, status.Error)

	err = controller.ShutdownInstance(runner)
	assert.NoError(t, err)
	assert.False(t, runner.IsRunning())
}

func TestController_ReconcileInstance_StartFailure(t *testing.T) {
	controller := NewController()

	object := &api.Object[Specification, Status]{
		Metadata: api.Metadata{
			Name: "test",
		},
		Specification: Specification{
			Execution: ExecutionSpecification{
				ExecutablePath: "/bin/non-existing-executable",
			},
		},
		Status: nil,
	}

	runner, err := controller.InitInstance(object)
	assert.NoError(t, err)
	assert.NotNil(t, runner)

	status, err := controller.ReconcileInstance(object, runner)
	object.Status = status
	assert.ErrorContains(t, err, "no such file or directory")
	assert.NotNil(t, status)
	assert.Contains(t, status.Error, "no such file or directory")

}
