package device

import (
	"context"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
	"sync"
)

type ServiceConfig struct {
}

type Service struct {
	ctx    context.Context
	wg     *sync.WaitGroup
	config ServiceConfig

	processService *process.Service
}

func New(ctx context.Context, wg *sync.WaitGroup, config ServiceConfig, processService *process.Service) (*Service, error) {
	return &Service{
		ctx:            ctx,
		wg:             wg,
		config:         config,
		processService: processService,
	}, nil
}
