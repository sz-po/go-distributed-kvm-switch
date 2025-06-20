package device

import (
	"context"
	"fmt"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
)

type Name string

type Phase string

const (
	Idle     Phase = "idle"
	Starting Phase = "starting"
	Running  Phase = "running"
	Stopping Phase = "stopping"
)

type Specification struct {
	Kind       string `json:"kind" validate:"required"`
	AutoEnable bool   `json:"autoEnable,omitempty"`
	Config     any    `json:"config,omitempty"`
}

type Status struct {
	ProcessName process.Name `json:"processName"`
	Enabled     bool         `json:"enabled"`
	Phase       Phase        `json:"phase"`
}

type Device interface {
	GetStatus() Status
	GetSpecification() Specification
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
	Restart(ctx context.Context) error
}

var ErrDeviceAlreadyEnabled = fmt.Errorf("device already enabled")
var ErrDeviceAlreadyDisabled = fmt.Errorf("device already disabled")
