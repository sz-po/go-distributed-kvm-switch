package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_Create(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus]())
	obj, err := service.Create("foo", testSpec{Foo: "bar"})
	assert.NotNil(t, obj)
	assert.NoError(t, err)

	obj, err = service.Create("foo", testSpec{Foo: "bar"})
	assert.Nil(t, obj)
	assert.ErrorIs(t, err, ErrObjectWithNameAlreadyExists)
}

func TestService_Create_BeforeCreateHook(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforeCreate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			return fmt.Errorf("hook error")
		}))
	obj, err := service.Create("foo", testSpec{Foo: "bar"})
	assert.Nil(t, obj)
	assert.ErrorContains(t, err, "hook error")

	obj, err = service.Get("foo")
	assert.Nil(t, obj)
	assert.ErrorIs(t, err, ErrObjectNotFound)

	hookCalled := false
	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforeCreate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return nil
		}))
	obj, err = service.Create("foo", testSpec{Foo: "bar"})
	assert.NotNil(t, obj)
	assert.NoError(t, err)
	assert.True(t, hookCalled)

	obj, err = service.Get("foo")
	assert.NotNil(t, obj)
	assert.NoError(t, err)
}

func TestService_Create_AfterCreateHook(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(AfterCreate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			return fmt.Errorf("hook error")
		}))

	obj, err := service.Create("foo", testSpec{Foo: "bar"})
	assert.Nil(t, obj)
	assert.ErrorContains(t, err, "hook error")

	obj, err = service.Get("foo")
	assert.NotNil(t, obj)
	assert.NoError(t, err)

	hookCalled := false
	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(AfterCreate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return nil
		}))
	obj, err = service.Create("foo", testSpec{Foo: "bar"})
	assert.NotNil(t, obj)
	assert.NoError(t, err)
	assert.True(t, hookCalled)

	obj, err = service.Get("foo")
	assert.NotNil(t, obj)
	assert.NoError(t, err)
}

func TestService_Create_Mutator(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceMutator[testSpec, testStatus](WhileCreatingObject, func(object *Object[testSpec, testStatus]) (*Object[testSpec, testStatus], error) {
			object.Specification.Foo = "baz"
			return object, nil
		}))

	obj, err := service.Create("foo", testSpec{Foo: "bar"})
	assert.NoError(t, err)
	assert.Equal(t, "baz", obj.Specification.Foo)

	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceMutator[testSpec, testStatus](WhileCreatingObject, func(object *Object[testSpec, testStatus]) (*Object[testSpec, testStatus], error) {
			return nil, fmt.Errorf("mutator error")
		}))

	obj, err = service.Create("foo", testSpec{Foo: "bar"})
	assert.Nil(t, obj)
	assert.ErrorContains(t, err, "mutator error")
}

func TestService_Delete(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus]())

	obj, err := service.Delete("foo")
	assert.Nil(t, obj)
	assert.ErrorIs(t, err, ErrObjectNotFound)

	service.Create("foo", testSpec{Foo: "bar"})

	obj, err = service.Delete("foo")
	assert.NoError(t, err)
	assert.True(t, obj.IsDeleted())

	obj, err = service.Delete("foo")
	assert.ErrorIs(t, err, ErrObjectAlreadyDeleted)
	assert.Nil(t, obj)
}

func TestService_Delete_BeforeDeleteHook(t *testing.T) {
	hookCalled := false
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforeDelete, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return fmt.Errorf("hook error")
		}))

	deletedObj, err := service.Delete("foo")
	assert.Nil(t, deletedObj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.False(t, hookCalled)

	service.Create("foo", testSpec{Foo: "bar"})

	deletedObj, err = service.Delete("foo")
	assert.Nil(t, deletedObj)
	assert.ErrorContains(t, err, "hook error")
	assert.True(t, hookCalled)

	hookCalled = false
	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforeDelete, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return nil
		}))

	service.Create("foo", testSpec{Foo: "bar"})
	deletedObj, err = service.Delete("foo")
	assert.NotNil(t, deletedObj)
	assert.NoError(t, err)
	assert.True(t, hookCalled)
}

func TestService_Delete_AfterDeleteHook(t *testing.T) {
	hookCalled := false
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(AfterDelete, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return fmt.Errorf("hook error")
		}))

	deletedObj, err := service.Delete("foo")
	assert.Nil(t, deletedObj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.False(t, hookCalled)

	service.Create("foo", testSpec{Foo: "bar"})
	deletedObj, err = service.Delete("foo")
	assert.Nil(t, deletedObj)
	assert.ErrorContains(t, err, "hook error")
	assert.True(t, hookCalled)

	obj, err := service.Get("foo", WithDeleted())
	assert.NotNil(t, obj)
	assert.NoError(t, err)
	assert.True(t, obj.IsDeleted())

	hookCalled = false
	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(AfterDelete, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return nil
		}))

	service.Create("foo", testSpec{Foo: "bar"})

	deletedObj, err = service.Delete("foo")
	assert.NotNil(t, deletedObj)
	assert.NoError(t, err)
	assert.True(t, hookCalled)
}

func TestService_Find(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus]())

	objects := service.Find()
	assert.Len(t, objects, 0)

	service.Create("foo", testSpec{Foo: "bar"})
	objects = service.Find()
	assert.Len(t, objects, 1)
}

func TestService_UpdateSpecification(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus]())

	obj, err := service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.Nil(t, obj)

	service.Create("foo", testSpec{Foo: "bar"})
	obj, err = service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func TestService_UpdateSpecification_Mutator(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceMutator[testSpec, testStatus](WhileUpdatingSpecification, func(object *Object[testSpec, testStatus]) (*Object[testSpec, testStatus], error) {
			object.Specification.Foo = "baz"
			return object, nil
		}))

	obj, err := service.Create("foo", testSpec{Foo: "bar"})
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Specification.Foo)

	obj, err = service.UpdateSpecification("foo", testSpec{Foo: "bar"})
	assert.NoError(t, err)
	assert.Equal(t, "baz", obj.Specification.Foo)

	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceMutator[testSpec, testStatus](WhileUpdatingSpecification, func(object *Object[testSpec, testStatus]) (*Object[testSpec, testStatus], error) {
			return nil, fmt.Errorf("mutator error")
		}))

	obj, err = service.Create("foo", testSpec{Foo: "bar"})
	assert.NotNil(t, obj)
	assert.NoError(t, err)

	obj, err = service.UpdateSpecification("foo", testSpec{Foo: "bar"})
	assert.Nil(t, obj)
	assert.ErrorContains(t, err, "mutator error")
}

func TestService_UpdateSpecification_BeforeSpecificationUpdateHook(t *testing.T) {
	hookCalled := false
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforeSpecificationUpdate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return fmt.Errorf("hook error")
		}))

	updatedObj, err := service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.Nil(t, updatedObj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.False(t, hookCalled)

	service.Create("foo", testSpec{Foo: "bar"})

	updatedObj, err = service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.Nil(t, updatedObj)
	assert.ErrorContains(t, err, "hook error")
	assert.True(t, hookCalled)

	hookCalled = false
	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforeSpecificationUpdate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return nil
		}))

	updatedObj, err = service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.Nil(t, updatedObj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.False(t, hookCalled)

	service.Create("foo", testSpec{Foo: "bar"})

	updatedObj, err = service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.NotNil(t, updatedObj)
	assert.NoError(t, err)
	assert.True(t, hookCalled)

	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforeSpecificationUpdate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			service.Delete(oldObject.Metadata.Name)
			return nil
		}))

	service.Create("foo", testSpec{Foo: "bar"})

	updatedObj, err = service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.Nil(t, updatedObj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
}

func TestService_UpdateSpecification_AfterSpecificationUpdateHook(t *testing.T) {
	hookCalled := false
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(AfterSpecificationUpdate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return fmt.Errorf("hook error")
		}))

	updatedObj, err := service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.Nil(t, updatedObj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.False(t, hookCalled)

	service.Create("foo", testSpec{Foo: "bar"})

	updatedObj, err = service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.Nil(t, updatedObj)
	assert.ErrorContains(t, err, "hook error")
	assert.True(t, hookCalled)

	obj, err := service.Get("foo")
	assert.Equal(t, "baz", obj.Specification.Foo)

	hookCalled = false
	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(AfterSpecificationUpdate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return nil
		}))

	updatedObj, err = service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.Nil(t, updatedObj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.False(t, hookCalled)

	service.Create("foo", testSpec{Foo: "bar"})

	updatedObj, err = service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.NotNil(t, updatedObj)
	assert.NoError(t, err)
	assert.True(t, hookCalled)
	assert.Equal(t, "baz", updatedObj.Specification.Foo)

	obj, _ = service.Get("foo")
	assert.Equal(t, "baz", obj.Specification.Foo)
}

func TestService_UpdateStatus(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus]())

	obj, err := service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.Nil(t, obj)

	service.Create("foo", testSpec{Foo: "bar"})
	obj, err = service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func TestService_UpdateStatus_Mutator(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceMutator[testSpec, testStatus](WhileUpdatingStatus, func(object *Object[testSpec, testStatus]) (*Object[testSpec, testStatus], error) {
			object.Status.Foo = "baz"
			return object, nil
		}))

	obj, err := service.Create("foo", testSpec{Foo: "bar"})
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Specification.Foo)

	obj, err = service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.NoError(t, err)
	assert.Equal(t, "baz", obj.Status.Foo)

	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceMutator[testSpec, testStatus](WhileUpdatingStatus, func(object *Object[testSpec, testStatus]) (*Object[testSpec, testStatus], error) {
			return nil, fmt.Errorf("mutator error")
		}))

	obj, err = service.Create("foo", testSpec{Foo: "bar"})
	assert.NotNil(t, obj)
	assert.NoError(t, err)

	obj, err = service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.Nil(t, obj)
	assert.ErrorContains(t, err, "mutator error")
}

func TestService_UpdateStatus_BeforeStatusUpdateHook(t *testing.T) {
	hookCalled := false
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforeStatusUpdate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return fmt.Errorf("hook error")
		}))

	updatedObj, err := service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.Nil(t, updatedObj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.False(t, hookCalled)

	service.Create("foo", testSpec{Foo: "bar"})

	updatedObj, err = service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.Nil(t, updatedObj)
	assert.ErrorContains(t, err, "hook error")
	assert.True(t, hookCalled)

	obj, err := service.Get("foo")
	assert.NoError(t, err)
	assert.Nil(t, obj.Status)

	hookCalled = false
	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforeStatusUpdate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return nil
		}))

	service.Create("foo", testSpec{Foo: "bar"})

	updatedObj, err = service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.NotNil(t, updatedObj)
	assert.NoError(t, err)
	assert.True(t, hookCalled)
	assert.Equal(t, testStatus{Foo: "bar"}, *updatedObj.Status)

	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforeStatusUpdate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			service.Delete(oldObject.Metadata.Name)
			return nil
		}))

	service.Create("foo", testSpec{Foo: "bar"})

	updatedObj, err = service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.Nil(t, updatedObj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
}

func TestService_UpdateStatus_AfterStatusUpdateHook(t *testing.T) {
	hookCalled := false
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(AfterStatusUpdate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return fmt.Errorf("hook error")
		}))

	updatedObj, err := service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.Nil(t, updatedObj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.False(t, hookCalled)

	service.Create("foo", testSpec{Foo: "bar"})

	updatedObj, err = service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.Nil(t, updatedObj)
	assert.ErrorContains(t, err, "hook error")
	assert.True(t, hookCalled)

	obj, err := service.Get("foo")
	assert.NoError(t, err)
	assert.NotNil(t, obj.Status)
	assert.Equal(t, testStatus{Foo: "bar"}, *obj.Status)

	hookCalled = false
	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(AfterStatusUpdate, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return nil
		}))

	service.Create("foo", testSpec{Foo: "bar"})

	updatedObj, err = service.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.NotNil(t, updatedObj)
	assert.NoError(t, err)
	assert.True(t, hookCalled)
	assert.Equal(t, testStatus{Foo: "bar"}, *updatedObj.Status)
}

func TestService_Prune(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus]())

	err := service.Prune("foo")
	assert.ErrorIs(t, err, ErrObjectNotFound)

	service.Create("foo", testSpec{Foo: "bar"})
	service.Delete("foo")

	err = service.Prune("foo")
	assert.NoError(t, err)
}

func TestService_Prune_BeforePruneHook(t *testing.T) {
	hookCalled := false
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforePrune, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return fmt.Errorf("hook error")
		}))

	service.Create("foo", testSpec{Foo: "bar"})
	service.Delete("foo")

	err := service.Prune("foo")
	assert.ErrorContains(t, err, "hook error")
	assert.True(t, hookCalled)

	obj, err := service.Get("foo", WithDeleted())
	assert.NotNil(t, obj)
	assert.NoError(t, err)
	assert.True(t, obj.IsDeleted())

	hookCalled = false
	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforePrune, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return nil
		}))

	service.Create("foo", testSpec{Foo: "bar"})
	service.Delete("foo")

	err = service.Prune("foo")
	assert.NoError(t, err)
	assert.True(t, hookCalled)

	obj, err = service.Get("foo", WithDeleted())
	assert.Nil(t, obj)
	assert.ErrorIs(t, err, ErrObjectNotFound)

	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(BeforePrune, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			service.store.Prune(oldObject.Metadata.Name)
			return nil
		}))

	service.Create("foo", testSpec{Foo: "bar"})
	service.Delete("foo")

	err = service.Prune("foo")
	assert.ErrorIs(t, err, ErrObjectNotFound)
}

func TestService_Prune_AfterPruneHook(t *testing.T) {
	hookCalled := false
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(AfterPrune, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return fmt.Errorf("hook error")
		}))

	service.Create("foo", testSpec{Foo: "bar"})

	err := service.Prune("foo")
	assert.ErrorIs(t, err, ErrObjectNotDeleted)
	assert.False(t, hookCalled)

	service.Delete("foo")

	err = service.Prune("foo")
	assert.ErrorContains(t, err, "hook error")
	assert.True(t, hookCalled)

	obj, err := service.Get("foo", WithDeleted())
	assert.Nil(t, obj)
	assert.ErrorIs(t, err, ErrObjectNotFound)

	hookCalled = false
	service = NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithServiceHook(AfterPrune, func(oldObject *Object[testSpec, testStatus], newObject *Object[testSpec, testStatus]) error {
			hookCalled = true
			return nil
		}))

	service.Create("foo", testSpec{Foo: "bar"})
	service.Delete("foo")
	err = service.Prune("foo")
	assert.NoError(t, err)
	assert.True(t, hookCalled)

	obj, err = service.Get("foo", WithDeleted())
	assert.Nil(t, obj)
	assert.ErrorIs(t, err, ErrObjectNotFound)
}
