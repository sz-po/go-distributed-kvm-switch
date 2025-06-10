package device

import (
	"context"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
)

type Name string

type Specification struct {
	Kind       string `json:"kind"`
	AutoEnable bool   `json:"autoEnable,omitempty"`
	Config     any    `json:"config,omitempty"`
}

type Status struct {
	ProcessName process.Name `json:"processName"`
	Enabled     bool         `json:"enabled"`
}

type Device interface {
	GetStatus() Status
	GetSpecification() Specification
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
	Restart(ctx context.Context) error
}
