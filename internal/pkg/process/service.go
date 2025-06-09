package process

import (
	"context"
	"fmt"
)

type Service interface {
	CreateProcess(ctx context.Context, name Name, specification Specification) (Process, error)
	DeleteProcess(ctx context.Context, name Name) error
	GetProcessByName(name Name) (Process, error)
	FindProcess() map[Name]Process
}

var ErrProcessNameAlreadyTaken = fmt.Errorf("process name already taken")
var ErrProcessNotFound = fmt.Errorf("process not found")
