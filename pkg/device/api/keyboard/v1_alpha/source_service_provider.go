package v1_alpha

import (
	"context"
	"github.com/sz-po/go-distributed-kvm-switch/pkg/device/protocol"
)

type SourceServiceProvider struct {
	service SourceService
}

var _ protocol.ServiceProvider = (*SourceServiceProvider)(nil)

func NewSourceServiceProvider(service SourceService) *SourceServiceProvider {
	return &SourceServiceProvider{
		service: service,
	}
}

func (provider *SourceServiceProvider) GetServiceName() protocol.ServiceName {
	return protocol.ServiceName("keyboard-source")
}

func (provider *SourceServiceProvider) GetMethods() map[protocol.MethodName]protocol.ServiceMethod {
	return map[protocol.MethodName]protocol.ServiceMethod{
		"get-num-lock-state":    provider.getNumLockState,
		"get-capslock-state":    provider.getCapslockState,
		"get-scroll-lock-state": provider.getScrollLockState,
	}
}

func (provider *SourceServiceProvider) getNumLockState(ctx context.Context, payload any) (any, error) {
	return provider.service.GetNumLockState(), nil
}

func (provider *SourceServiceProvider) getScrollLockState(ctx context.Context, payload any) (any, error) {
	return provider.service.GetScrollLockState(), nil
}

func (provider *SourceServiceProvider) getCapslockState(ctx context.Context, payload any) (any, error) {
	return provider.service.GetCapslockState(), nil
}
