package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testSpec struct {
	Foo        string
	DefaultFoo string `default:"bar"`
}

type testStatus struct {
	Foo        string
	DefaultFoo string `default:"bar"`
}

func testObjectStore_Create(t *testing.T, store ObjectStore[testSpec, testStatus]) {
	createdObj, err := store.Create("foo", testSpec{Foo: "bar"})
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)

	obj, _ := store.Get("foo")
	assert.NotNil(t, obj)
	assert.False(t, obj.IsDeleted())
	assert.True(t, obj.Metadata.DeletedAt.IsEmpty())
	assert.False(t, obj.Metadata.CreatedAt.IsEmpty())
	assert.False(t, obj.Metadata.SpecificationUpdatedAt.IsEmpty())
	assert.True(t, obj.Metadata.StatusUpdatedAt.IsEmpty())
	assert.Nil(t, obj.Status)

	assert.Equal(t, createdObj, obj)

	obj, err = store.Create("foo", testSpec{Foo: "bar"})
	assert.ErrorIs(t, err, ErrObjectWithNameAlreadyExists)
	assert.Nil(t, obj)

	store.Delete("foo")
	obj, err = store.Create("foo", testSpec{Foo: "bar"})
	assert.ErrorIs(t, err, ErrDeletedObjectWithNameAlreadyExists)
	assert.Nil(t, obj)

	store.Prune("foo")
	obj, err = store.Create("foo", testSpec{Foo: "bar"})
	assert.NoError(t, err)
	assert.NotNil(t, obj)

	obj, err = store.Create("bar", testSpec{Foo: "bar"})
	assert.NoError(t, err)
	assert.NotNil(t, obj)

	spec := testSpec{
		Foo: "bar",
	}
	obj, err = store.Create("baz", spec)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	spec.Foo = "baz"

	obj, _ = store.Get("baz")
	assert.Equal(t, "bar", obj.Specification.Foo)
	obj.Specification.Foo = "bax"
	assert.Equal(t, "bax", obj.Specification.Foo)
	assert.Equal(t, "baz", spec.Foo)
}

func testObjectStore_Delete(t *testing.T, store ObjectStore[testSpec, testStatus]) {
	obj, err := store.Delete("foo")
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.Nil(t, obj)

	store.Create("foo", testSpec{Foo: "bar"})
	deletedObj, err := store.Delete("foo")
	assert.NoError(t, err)
	assert.NotNil(t, deletedObj)
	assert.True(t, deletedObj.IsDeleted())

	obj, err = store.Get("foo", WithDeleted())
	assert.NotNil(t, obj)
	assert.NoError(t, err)
	assert.Equal(t, deletedObj, obj)

	obj, err = store.Delete("foo")
	assert.ErrorIs(t, err, ErrObjectAlreadyDeleted)
	assert.Nil(t, obj)

	_, err = store.Create("foo", testSpec{Foo: "bar"})
	assert.ErrorIs(t, err, ErrDeletedObjectWithNameAlreadyExists)

	obj, err = store.Get("foo")
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
	obj, err := store.UpdateSpecification("foo", testSpec{Foo: "bar"})
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.Nil(t, obj)

	store.Create("foo", testSpec{Foo: "bar"})
	store.Delete("foo")
	obj, err = store.UpdateSpecification("foo", testSpec{Foo: "bar"})
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.Nil(t, obj)
	store.Prune("foo")

	oldSpec := testSpec{Foo: "bar"}
	newSpec := testSpec{Foo: "baz"}

	_, err = store.Create("foo", oldSpec)
	assert.NoError(t, err)

	oldObj, _ := store.Get("foo")
	assert.False(t, oldObj.Metadata.SpecificationUpdatedAt.IsEmpty())
	assert.Equal(t, oldSpec, oldObj.Specification)

	time.Sleep(time.Millisecond)

	updatedObj, err := store.UpdateSpecification("foo", newSpec)
	assert.NoError(t, err)
	assert.NotNil(t, updatedObj)

	newObj, _ := store.Get("foo")
	assert.False(t, newObj.Metadata.SpecificationUpdatedAt.IsEmpty())
	assert.Equal(t, newSpec, newObj.Specification)
	assert.Equal(t, newObj, updatedObj)

	assert.NotEqual(t, oldObj.Specification, newObj.Specification)
	assert.NotEqual(t, oldObj.Metadata.SpecificationUpdatedAt, newObj.Metadata.SpecificationUpdatedAt)
	assert.NotEqual(t, oldSpec, newSpec)
}

func testObjectStore_UpdateStatus(t *testing.T, store ObjectStore[testSpec, testStatus]) {
	obj, err := store.UpdateStatus("foo", testStatus{Foo: "bar"})
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.Nil(t, obj)

	store.Create("foo", testSpec{Foo: "bar"})

	obj, _ = store.Get("foo")
	assert.True(t, obj.Metadata.StatusUpdatedAt.IsEmpty())
	assert.Nil(t, obj.Status)

	oldStatus := testStatus{Foo: "bar"}
	newStatus := testStatus{Foo: "baz"}

	updatedObj, err := store.UpdateStatus("foo", oldStatus)
	assert.NoError(t, err)

	oldObj, _ := store.Get("foo")
	assert.Equal(t, oldStatus, *oldObj.Status)
	assert.False(t, oldObj.Metadata.StatusUpdatedAt.IsEmpty())
	assert.Equal(t, updatedObj, oldObj)

	time.Sleep(time.Millisecond)

	updatedObj, err = store.UpdateStatus("foo", newStatus)
	assert.NoError(t, err)

	newObj, _ := store.Get("foo")
	assert.False(t, newObj.Metadata.StatusUpdatedAt.IsEmpty())
	assert.NotEqual(t, oldObj.Status, newObj.Status)
	assert.NotEqual(t, oldObj.Metadata.StatusUpdatedAt, newObj.Metadata.StatusUpdatedAt)
	assert.NotEqual(t, oldStatus, newStatus)
	assert.Equal(t, updatedObj, newObj)

	store.Delete("foo")

	obj, err = store.UpdateStatus("foo", newStatus)
	assert.ErrorIs(t, err, ErrObjectNotFound)
	assert.Nil(t, obj)
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
