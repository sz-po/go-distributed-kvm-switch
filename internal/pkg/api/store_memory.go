package api

import (
	"github.com/brunoga/deep"
	"sync"
)

type MemoryObjectStore[TSpec Specification, TStatus Status] struct {
	objects      map[ObjectName]*Object[TSpec, TStatus]
	objectsMutex *sync.RWMutex
}

func NewMemoryObjectStore[TSpec Specification, TStatus Status]() *MemoryObjectStore[TSpec, TStatus] {
	return &MemoryObjectStore[TSpec, TStatus]{
		objects:      make(map[ObjectName]*Object[TSpec, TStatus]),
		objectsMutex: &sync.RWMutex{},
	}
}

func (store *MemoryObjectStore[TSpec, TStatus]) Create(name ObjectName, specification TSpec) error {
	store.objectsMutex.Lock()
	defer store.objectsMutex.Unlock()

	if object, exists := store.objects[name]; exists && !object.IsDeleted() {
		return ErrObjectWithNameAlreadyExists
	} else if exists && object.IsDeleted() {
		return ErrDeletedObjectWithNameAlreadyExists
	}

	store.objects[name] = &Object[TSpec, TStatus]{
		Metadata: Metadata{
			Name:   name,
			Labels: map[string]string{},

			CreatedAt:              Now(),
			SpecificationUpdatedAt: Now(),
		},
		Specification: deep.MustCopy(specification),
		Status:        nil,
	}

	return nil
}

func (store *MemoryObjectStore[TSpec, TStatus]) UpdateSpecification(name ObjectName, spec TSpec) error {
	store.objectsMutex.Lock()
	defer store.objectsMutex.Unlock()

	if object, exists := store.objects[name]; !exists {
		return ErrObjectNotFound
	} else if exists && object.IsDeleted() {
		return ErrObjectNotFound
	}

	store.objects[name].Specification = deep.MustCopy(spec)

	return nil
}

func (store *MemoryObjectStore[TSpec, TStatus]) UpdateStatus(name ObjectName, status TStatus) error {
	store.objectsMutex.Lock()
	defer store.objectsMutex.Unlock()

	if object, exists := store.objects[name]; !exists {
		return ErrObjectNotFound
	} else if exists && object.IsDeleted() {
		return ErrObjectNotFound
	}

	statusCopy := deep.MustCopy(status)

	object := store.objects[name]
	object.Metadata.StatusUpdatedAt = Now()
	object.Status = &statusCopy

	return nil
}

func (store *MemoryObjectStore[TSpec, TStatus]) Get(name ObjectName, queryOpts ...ObjectStoreQuery) (*Object[TSpec, TStatus], error) {
	store.objectsMutex.RLock()
	defer store.objectsMutex.RUnlock()

	query := objectStoreQuery{}
	query.apply(queryOpts)

	if object, exists := store.objects[name]; !exists {
		return nil, ErrObjectNotFound
	} else if object.IsDeleted() && !query.withDeleted {
		return nil, ErrObjectNotFound
	}

	object := store.objects[name]
	object.Metadata.SpecificationUpdatedAt = Now()

	return &Object[TSpec, TStatus]{
		Metadata:      deep.MustCopy(object.Metadata),
		Specification: deep.MustCopy(object.Specification),
		Status:        deep.MustCopy(object.Status),
	}, nil
}

func (store *MemoryObjectStore[TSpec, TStatus]) Delete(name ObjectName) error {
	store.objectsMutex.Lock()
	defer store.objectsMutex.Unlock()

	if object, exists := store.objects[name]; !exists {
		return ErrObjectNotFound
	} else if object.IsDeleted() {
		return ErrObjectAlreadyDeleted
	}

	store.objects[name].Metadata.DeletedAt = Now()

	return nil
}

func (store *MemoryObjectStore[TSpec, TStatus]) Prune(name ObjectName) error {
	store.objectsMutex.Lock()
	defer store.objectsMutex.Unlock()

	if object, exists := store.objects[name]; !exists {
		return ErrObjectNotFound
	} else if !object.IsDeleted() {
		return ErrObjectNotDeleted
	}

	delete(store.objects, name)

	return nil
}

func (store *MemoryObjectStore[TSpec, TStatus]) Find(queryOpts ...ObjectStoreQuery) []ObjectName {
	store.objectsMutex.RLock()
	defer store.objectsMutex.RUnlock()

	query := &objectStoreQuery{}
	query.apply(queryOpts)

	var result []ObjectName

	for objectName, object := range store.objects {
		if object.IsDeleted() && !query.withDeleted {
			continue
		}

		if store.objectMatch(query, object) {
			result = append(result, objectName)
		}
	}

	return result
}

func (store *MemoryObjectStore[TSpec, TStatus]) objectMatch(query *objectStoreQuery, object *Object[TSpec, TStatus]) bool {
	match := true

	if query.nameFilter != "" && query.nameFilter != string(object.Metadata.Name) {
		match = false
	}

	return match
}
