package protocol

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"testing"
)

type mockedServiceProvider struct {
	mock.Mock
}

func (serviceProvider *mockedServiceProvider) GetServiceName() ServiceName {
	args := serviceProvider.MethodCalled("GetServiceName")

	return args.Get(0).(ServiceName)
}

func (serviceProvider *mockedServiceProvider) GetMethods() map[MethodName]ServiceMethod {
	args := serviceProvider.MethodCalled("GetMethods")

	return args.Get(0).(map[MethodName]ServiceMethod)
}

func TestServiceRegistry_CallServiceMethod(t *testing.T) {
	ctx := context.Background()

	serviceProvider := &mockedServiceProvider{}
	serviceProvider.On("GetServiceName").Return(ServiceName("foo"))
	serviceProvider.On("GetMethods").Return(map[MethodName]ServiceMethod{
		"bar": func(ctx context.Context, payload any) (any, error) {
			return payload, nil
		},
	})

	serviceRegistry := NewServiceRegistry(WithServiceProvider(serviceProvider))

	payload, err := serviceRegistry.CallServiceMethod(ctx, ServiceName("non-existing"), MethodName("bar"), nil)
	assert.ErrorIs(t, err, ErrServiceNotFound)
	assert.Nil(t, payload)

	payload, err = serviceRegistry.CallServiceMethod(ctx, ServiceName("foo"), MethodName("non-existing"), nil)
	assert.ErrorIs(t, err, ErrServiceMethodNotFound)
	assert.Nil(t, payload)

	payload, err = serviceRegistry.CallServiceMethod(ctx, ServiceName("foo"), MethodName("bar"), "baz")
	assert.NoError(t, err)
	assert.Equal(t, "baz", payload)
}
