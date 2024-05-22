package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testController struct {
	requiredInitRetries      int
	requiredReconcileRetries int
	requiredShutdownRetries  int

	initialized bool
	reconciled  bool
	finished    bool
	instance    *string
}

func (controller *testController) InitInstance(object *Object[testSpec, testStatus]) (*string, error) {
	if controller.requiredInitRetries > 0 {
		controller.requiredInitRetries--
		return nil, fmt.Errorf("init error")
	}

	instance := ""

	controller.initialized = true
	controller.instance = &instance

	return &instance, nil
}

func (controller *testController) ReconcileInstance(object *Object[testSpec, testStatus], instance *string) (*testStatus, error) {
	if controller.requiredReconcileRetries > 0 {
		controller.requiredReconcileRetries--
		return nil, fmt.Errorf("reconcile error")
	}

	*instance = object.Specification.Foo

	controller.reconciled = true

	return &testStatus{
		Foo: *instance,
	}, nil
}

func (controller *testController) ShutdownInstance(instance *string) error {
	if controller.requiredShutdownRetries > 0 {
		controller.requiredShutdownRetries--
		return fmt.Errorf("shutdown error")
	}
	*instance = ""

	controller.finished = true

	return nil
}

func TestWithController(t *testing.T) {
	ticker := make(chan time.Time)
	controller := testController{
		requiredInitRetries:      1,
		requiredReconcileRetries: 1,
		requiredShutdownRetries:  1,
	}

	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithController[testSpec, testStatus, string](&controller, ticker))

	obj, err := service.Create("foo", testSpec{Foo: "bar"})
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Nil(t, obj.Status)

	ticker <- time.Now()
	ticker <- time.Now()
	time.Sleep(time.Millisecond * 50)

	assert.NotNil(t, controller.instance)
	assert.Empty(t, *controller.instance)
	assert.True(t, controller.initialized)
	assert.False(t, controller.reconciled)
	assert.False(t, controller.finished)

	ticker <- time.Now()
	ticker <- time.Now()
	time.Sleep(time.Millisecond * 50)

	assert.Equal(t, "bar", *controller.instance)
	assert.True(t, controller.initialized)
	assert.True(t, controller.reconciled)
	assert.False(t, controller.finished)

	obj, _ = service.Get("foo")
	assert.NotNil(t, obj)
	assert.Equal(t, "bar", obj.Status.Foo)

	service.Delete("foo")

	ticker <- time.Now()
	ticker <- time.Now()
	ticker <- time.Now()
	time.Sleep(time.Millisecond * 50)

	assert.Equal(t, "", *controller.instance)
	assert.True(t, controller.initialized)
	assert.True(t, controller.reconciled)
	assert.True(t, controller.finished)

	ticker <- time.Now()
	time.Sleep(time.Millisecond * 50)

	obj, err = service.Get("foo", WithDeleted())
	assert.Nil(t, obj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
}
