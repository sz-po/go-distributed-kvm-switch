package process

import (
	"context"
	"fmt"
)

func CreateServiceProcessFactory(service Service) ProcessFactory {
	return func(name Name, specification Specification, opts ...ProcessOpt) (Process, error) {
		return service.CreateProcess(context.Background(), name, specification, opts...)
	}
}

type Service interface {
	CreateProcess(ctx context.Context, name Name, specification Specification, opts ...ProcessOpt) (Process, error)
	DeleteProcess(ctx context.Context, name Name) error
	GetProcessByName(name Name) (Process, error)
	FindProcess() map[Name]Process
}

var ErrProcessNameAlreadyTaken = fmt.Errorf("process name already taken")
var ErrProcessNotFound = fmt.Errorf("process not found")
