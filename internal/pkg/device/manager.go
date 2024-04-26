package device

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

type runtime struct {
	reader io.Reader
	writer io.Writer
}

type descriptor struct {
	metadata   *Metadata
	config     Config
	status     *Status
	descriptor *descriptor
}

type Manager struct {
	devices      map[Id]*descriptor
	devicesMutex *sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		devices:      make(map[Id]*descriptor),
		devicesMutex: &sync.RWMutex{},
	}
}

func (manager *Manager) Start(ctx context.Context) error {
	panic("not implemented")
}

func (manager *Manager) Create(metadata Metadata, config Config) (Id, error) {
	manager.devicesMutex.Lock()
	defer manager.devicesMutex.Unlock()

	id := metadata.Id()

	if exists, _ := manager.devices[id]; exists != nil {
		return IdEmpty, ErrDeviceNameAlreadyTaken
	}

	deviceDescriptor := &descriptor{
		metadata: &metadata,
		config:   config,
		status: &Status{
			CreatedAt: time.Now().String(),
			Process:   "",
		},
	}

	manager.devices[id] = deviceDescriptor

	return id, nil
}

var ErrDeviceNameAlreadyTaken = fmt.Errorf("device name already taken")
