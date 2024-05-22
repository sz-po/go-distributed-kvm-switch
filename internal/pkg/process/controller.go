package process

import (
	"context"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api"
	"time"
)

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

func (controller *Controller) InitInstance(object *api.Object[Specification, Status]) (*Runner, error) {
	runner := NewRunner(object.Specification.Execution.ExecutablePath,
		WithArgs(object.Specification.Execution.Arguments...),
	)

	return runner, nil
}

func (controller *Controller) ReconcileInstance(object *api.Object[Specification, Status], runner *Runner) (*Status, error) {
	status := Status{
		IsRunning: false,
		ProcessID: 0,
		ExitCode:  0,
		Error:     "",
	}

	if runner.IsRunning() {
		status.IsRunning = true
		status.ProcessID, _ = runner.GetPID()
	} else {
		status.IsRunning = false
		status.ExitCode, _ = runner.GetExitCode()
	}

	if !runner.IsRunning() {
		if err := runner.Start(); err != nil {
			status.Error = err.Error()
			return &status, err
		}
	}

	return &status, nil
}

func (controller *Controller) ShutdownInstance(instance *Runner) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	return instance.Stop(ctx)
}
