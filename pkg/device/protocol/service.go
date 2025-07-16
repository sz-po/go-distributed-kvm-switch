package protocol

import (
	"context"
	"errors"
)

type ServiceMethod func(ctx context.Context, payload any) (any, error)

type ServiceProvider interface {
	GetServiceName() ServiceName
	GetMethods() map[MethodName]ServiceMethod
}

type ServiceRegistryOpt func(*ServiceRegistry)

type ServiceRegistry struct {
	services map[ServiceName]ServiceProvider
}

func WithServiceProvider(provider ServiceProvider) ServiceRegistryOpt {
	return func(registry *ServiceRegistry) {
		registry.services[provider.GetServiceName()] = provider
	}
}

func NewServiceRegistry(opts ...ServiceRegistryOpt) *ServiceRegistry {
	registry := &ServiceRegistry{
		services: make(map[ServiceName]ServiceProvider),
	}

	for _, opt := range opts {
		opt(registry)
	}

	return registry
}

func (registry *ServiceRegistry) CallServiceMethod(ctx context.Context, serviceName ServiceName, methodName MethodName, payload any) (any, error) {
	provider, ok := registry.services[serviceName]
	if !ok {
		return nil, ErrServiceNotFound
	}

	methods := provider.GetMethods()

	method, ok := methods[methodName]
	if !ok {
		return nil, ErrServiceMethodNotFound
	}

	return method(ctx, payload)
}

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceMethodNotFound = errors.New("service method not found")
