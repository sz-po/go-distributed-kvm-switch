package device

import (
	"context"
	"fmt"
)

type Service interface {
	CreateDevice(ctx context.Context, name Name, specification Specification) (Device, error)
	DeleteDevice(ctx context.Context, name Name) error
	GetDeviceByName(name Name) (Device, error)
	HasDevice(name Name) bool
}

var ErrDeviceNameAlreadyTaken = fmt.Errorf("device name already taken")
var ErrDeviceNotFound = fmt.Errorf("device not found")
