package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testSpec struct {
	Foo string
}

type testStatus struct {
	Foo string
}

func testObjectStore_Create(t *testing.T, store ObjectStore[testSpec, testStatus]) {
	err := store.Create("foo", testSpec{Foo: "bar"})
	assert.NoError(t, err)

	obj, _ := store.Get("foo")
	assert.False(t, obj.IsDeleted())
	assert.True(t, obj.Metadata.DeletedAt.IsEmpty())
	assert.False(t, obj.Metadata.CreatedAt.IsEmpty())
	assert.False(t, obj.Metadata.SpecificationUpdatedAt.IsEmpty())
	assert.True(t, obj.Metadata.StatusUpdatedAt.IsEmpty())
	assert.Nil(t, obj.Status)

	err = store.Create("foo", testSpec{Foo: "bar"})
	assert.ErrorIs(t, err, ErrObjectWithNameAlreadyExists)

	store.Delete("foo")
	err = store.Create("foo", testSpec{Foo: "bar"})
	assert.ErrorIs(t, err, ErrDeletedObjectWithNameAlreadyExists)

	store.Prune("foo")
	err = store.Create("foo", testSpec{Foo: "bar"})
	assert.NoError(t, err)

	err = store.Create("bar", testSpec{Foo: "bar"})
	assert.NoError(t, err)

	spec := testSpec{
		Foo: "bar",
	}
	err = store.Create("baz", spec)
	assert.NoError(t, err)
	spec.Foo = "baz"

	obj, _ = store.Get("baz")
	assert.Equal(t, "bar", obj.Specification.Foo)
	obj.Specification.Foo = "bax"
	assert.Equal(t, "bax", obj.Specification.Foo)
	assert.Equal(t, "baz", spec.Foo)
}

func testObjectStore_Delete(t *testing.T, store ObjectStore[testSpec, testStatus]) {
	err := store.Delete("foo")
	assert.ErrorIs(t, err, ErrObjectNotFound)

	store.Create("foo", testSpec{Foo: "bar"})
	err = store.Delete("foo")
	assert.NoError(t, err)

	err = store.Delete("foo")
	assert.ErrorIs(t, err, ErrObjectAlreadyDeleted)

	err = store.Create("foo", testSpec{Foo: "bar"})
	assert.ErrorIs(t, err, ErrDeletedObjectWithNameAlreadyExists)

	obj, err := store.Get("foo")
	assert.Nil(t, obj)
	assert.ErrorIs(t, err, ErrObjectNotFound)

	obj, err = store.Get("foo", WithDeleted())
	assert.NotNil(t, obj)
	assert.NoError(t, err)
	assert.Equal(t, testSpec{Foo: "bar"}, obj.Specification)
	assert.True(t, obj.IsDeleted())
	assert.False(t, obj.Metadata.DeletedAt.IsEmpty())
}

func testObjectStore_Get(t *testing.T, store ObjectStore[testSpec, testStatus]) {
	obj, err := store.Get("foo")
	assert.Nil(t, obj)
	assert.ErrorIs(t, err, ErrObjectNotFound)

	spec := testSpec{Foo: "bar"}
	store.Create("foo", spec)

	obj, err = store.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, spec, obj.Specification)
	assert.Nil(t, obj.Status)

	spec.Foo = "baz"
	assert.NotEqual(t, spec, obj.Specification)
}

func testObjectStore_UpdateSpecification(t *testing.T, store ObjectStore[testSpec, testStatus]) {
	err := store.UpdateSpecification("foo", testSpec{Foo: "bar"})
	assert.ErrorIs(t, err, ErrObjectNotFound)

	store.Create("foo", testSpec{Foo: "bar"})
	store.Delete("foo")
	err = store.UpdateSpecification("foo", testSpec{Foo: "bar"})
	assert.ErrorIs(t, err, ErrObjectNotFound)
	store.Prune("foo")

	oldSpec := testSpec{Foo: "bar"}
	newSpec := testSpec{Foo: "baz"}

	err = store.Create("foo", oldSpec)
	assert.NoError(t, err)

	oldObj, _ := store.Get("foo")
	assert.False(t, oldObj.Metadata.SpecificationUpdatedAt.IsEmpty())
	assert.Equal(t, oldSpec, oldObj.Specification)

	time.Sleep(time.Millisecond)

	err = store.UpdateSpecification("foo", newSpec)
	assert.NoError(t, err)

	newObj, _ := store.Get("foo")
	assert.False(t, newObj.Metadata.SpecificationUpdatedAt.IsEmpty())
	assert.Equal(t, newSpec, newObj.Specification)

	assert.NotEqual(t, oldObj.Specification, newObj.Specification)
	assert.NotEqual(t, oldObj.Metadata.SpecificationUpdatedAt, newObj.Metadata.SpecificationUpdatedAt)
	assert.NotEqual(t, oldSpec, newSpec)
}

func testObjectStore_UpdateStatus(t *testing.T, store ObjectStore[testSpec, testStatus]) {
	err := store.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.ErrorIs(t, err, ErrObjectNotFound)

	store.Create("foo", testSpec{Foo: "bar"})

	obj, _ := store.Get("foo")
	assert.True(t, obj.Metadata.StatusUpdatedAt.IsEmpty())
	assert.Nil(t, obj.Status)

	oldStatus := testStatus{Foo: "bar"}
	newStatus := testStatus{Foo: "baz"}

	err = store.UpdateStatus("foo", oldStatus)
	assert.NoError(t, err)

	oldObj, _ := store.Get("foo")
	assert.Equal(t, oldStatus, *oldObj.Status)
	assert.False(t, oldObj.Metadata.StatusUpdatedAt.IsEmpty())

	time.Sleep(time.Millisecond)

	err = store.UpdateStatus("foo", newStatus)
	assert.NoError(t, err)

	newObj, _ := store.Get("foo")
	assert.False(t, newObj.Metadata.StatusUpdatedAt.IsEmpty())
	assert.NotEqual(t, oldObj.Status, newObj.Status)
	assert.NotEqual(t, oldObj.Metadata.StatusUpdatedAt, newObj.Metadata.StatusUpdatedAt)
	assert.NotEqual(t, oldStatus, newStatus)

	store.Delete("foo")

	err = store.UpdateStatus("foo", newStatus)
	assert.ErrorIs(t, err, ErrObjectNotFound)
}

func testObjectStore_Prune(t *testing.T, store ObjectStore[testSpec, testStatus]) {
	err := store.Prune("foo")
	assert.ErrorIs(t, err, ErrObjectNotFound)

	store.Create("foo", testSpec{Foo: "bar"})

	err = store.Prune("foo")
	assert.ErrorIs(t, err, ErrObjectNotDeleted)

	store.Delete("foo")

	err = store.Prune("foo")
	assert.NoError(t, err)

	_, err = store.Get("foo")
	assert.ErrorIs(t, err, ErrObjectNotFound)
}

func testObjectStore_Find(t *testing.T, store ObjectStore[testSpec, testStatus]) {
	objects := store.Find()
	assert.Empty(t, objects)

	store.Create("foo", testSpec{Foo: "bar"})

	objects = store.Find()
	assert.Len(t, objects, 1)
	assert.Equal(t, ObjectName("foo"), objects[0])

	store.Create("bar", testSpec{Foo: "bar"})
	objects = store.Find()
	assert.Len(t, objects, 2)
	assert.Equal(t, ObjectName("foo"), objects[0])
	assert.Equal(t, ObjectName("bar"), objects[1])

	store.Delete("foo")
	objects = store.Find()
	assert.Len(t, objects, 1)
	assert.Equal(t, ObjectName("bar"), objects[0])

	objects = store.Find(WithDeleted())
	assert.Len(t, objects, 2)
	assert.Equal(t, ObjectName("foo"), objects[0])
	assert.Equal(t, ObjectName("bar"), objects[1])

	objects = store.Find(WithDeleted(), WhereObjectName("foo"))
	assert.Len(t, objects, 1)
	assert.Equal(t, ObjectName("foo"), objects[0])

	store.Prune("foo")
	objects = store.Find(WithDeleted())
	assert.Len(t, objects, 1)
	assert.Equal(t, ObjectName("bar"), objects[0])

	objects = store.Find(WhereObjectName("foo"))
	assert.Empty(t, objects)
}

func TestMemoryObjectStore_UpdateStatus(t *testing.T) {
	store := NewMemoryObjectStore[testSpec, testStatus]()
	testObjectStore_UpdateStatus(t, store)
}

func TestMemoryObjectStore_UpdateSpecification(t *testing.T) {
	store := NewMemoryObjectStore[testSpec, testStatus]()
	testObjectStore_UpdateSpecification(t, store)
}

func TestMemoryObjectStore_Delete(t *testing.T) {
	store := NewMemoryObjectStore[testSpec, testStatus]()
	testObjectStore_Delete(t, store)
}

func TestMemoryObjectStore_Create(t *testing.T) {
	store := NewMemoryObjectStore[testSpec, testStatus]()
	testObjectStore_Create(t, store)
}

func TestMemoryObjectStore_Get(t *testing.T) {
	store := NewMemoryObjectStore[testSpec, testStatus]()
	testObjectStore_Get(t, store)
}

func TestMemoryObjectStore_Prune(t *testing.T) {
	store := NewMemoryObjectStore[testSpec, testStatus]()
	testObjectStore_Prune(t, store)
}

func TestMemoryObjectStore_Find(t *testing.T) {
	store := NewMemoryObjectStore[testSpec, testStatus]()
	testObjectStore_Find(t, store)
}
