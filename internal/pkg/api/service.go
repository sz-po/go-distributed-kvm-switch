package api

import (
	"fmt"
	"github.com/brunoga/deep"
)

type ServiceHookKind string

const (
	BeforeCreate              ServiceHookKind = "BeforeCreate"
	AfterCreate                               = "AfterCreate"
	BeforeSpecificationUpdate                 = "BeforeSpecificationUpdate"
	AfterSpecificationUpdate                  = "AfterSpecificationUpdate"
	BeforeStatusUpdate                        = "BeforeStatusUpdate"
	AfterStatusUpdate                         = "AfterStatusUpdate"
	BeforeDelete                              = "BeforeDelete"
	AfterDelete                               = "AfterDelete"
	BeforePrune                               = "BeforePrune"
	AfterPrune                                = "AfterPrune"
)

type ServiceHook[TSpec Specification, TStatus Status] func(*Object[TSpec, TStatus], *Object[TSpec, TStatus]) error

type ServiceOpt[TSpec Specification, TStatus Status] func(*Service[TSpec, TStatus])

type Service[TSpec Specification, TStatus Status] struct {
	store ObjectStore[TSpec, TStatus]

	hooks map[ServiceHookKind][]ServiceHook[TSpec, TStatus]
}

func NewService[TSpec Specification, TStatus Status](store ObjectStore[TSpec, TStatus], opts ...ServiceOpt[TSpec, TStatus]) *Service[TSpec, TStatus] {
	service := &Service[TSpec, TStatus]{
		store: store,

		hooks: make(map[ServiceHookKind][]ServiceHook[TSpec, TStatus]),
	}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func (service *Service[TSpec, TStatus]) Create(name ObjectName, spec TSpec) (*Object[TSpec, TStatus], error) {
	err := service.callHook(BeforeCreate, &Object[TSpec, TStatus]{Specification: spec}, nil)
	if err != nil {
		return nil, err
	}

	newObject, err := service.store.Create(name, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to create object: %w", err)
	}

	err = service.callHook(AfterCreate, nil, newObject)
	if err != nil {
		return nil, err
	}

	return newObject, nil
}

func (service *Service[TSpec, TStatus]) UpdateSpecification(name ObjectName, spec TSpec) (*Object[TSpec, TStatus], error) {
	oldObject, err := service.store.Get(name)
	if err != nil {
		return nil, err
	}

	newObject := deep.MustCopy(oldObject)
	newObject.Specification = deep.MustCopy(spec)

	err = service.callHook(BeforeSpecificationUpdate, oldObject, newObject)
	if err != nil {
		return nil, err
	}

	newObject, err = service.store.UpdateSpecification(name, spec)
	if err != nil {
		return nil, err
	}

	err = service.callHook(AfterSpecificationUpdate, oldObject, newObject)
	if err != nil {
		return nil, err
	}

	return newObject, nil
}

func (service *Service[TSpec, TStatus]) UpdateStatus(name ObjectName, status TStatus) (*Object[TSpec, TStatus], error) {
	oldObject, err := service.store.Get(name)
	if err != nil {
		return nil, err
	}

	newObject := deep.MustCopy(oldObject)
	newObject.Status = deep.MustCopy(&status)

	err = service.callHook(BeforeStatusUpdate, oldObject, newObject)
	if err != nil {
		return nil, err
	}

	newObject, err = service.store.UpdateStatus(name, status)
	if err != nil {
		return nil, err
	}

	err = service.callHook(AfterStatusUpdate, oldObject, newObject)
	if err != nil {
		return nil, err
	}

	return newObject, nil
}

func (service *Service[TSpec, TStatus]) Get(name ObjectName, query ...ObjectStoreQuery) (*Object[TSpec, TStatus], error) {
	return service.store.Get(name, query...)
}

func (service *Service[TSpec, TStatus]) Delete(name ObjectName) (*Object[TSpec, TStatus], error) {
	oldObject, err := service.store.Get(name, WithDeleted())
	if err != nil {
		return nil, err
	}

	err = service.callHook(BeforeDelete, oldObject, nil)
	if err != nil {
		return nil, err
	}

	newObject, err := service.store.Delete(name)
	if err != nil {
		return nil, err
	}

	err = service.callHook(AfterDelete, oldObject, newObject)
	if err != nil {
		return nil, err
	}

	return newObject, nil
}

func (service *Service[TSpec, TStatus]) Prune(name ObjectName) error {
	oldObject, err := service.store.Get(name, WithDeleted())
	if err != nil {
		return err
	}

	err = service.callHook(BeforePrune, oldObject, nil)
	if err != nil {
		return err
	}

	err = service.store.Prune(name)
	if err != nil {
		return err
	}

	err = service.callHook(AfterPrune, oldObject, nil)
	if err != nil {
		return err
	}

	return nil
}

func (service *Service[TSpec, TStatus]) Find(query ...ObjectStoreQuery) []ObjectName {
	return service.store.Find(query...)
}

func (service *Service[TSpec, TStatus]) attachHook(kind ServiceHookKind, hook ServiceHook[TSpec, TStatus]) {
	if _, exists := service.hooks[kind]; !exists {
		service.hooks[kind] = make([]ServiceHook[TSpec, TStatus], 0)
	}

	service.hooks[kind] = append(service.hooks[kind], hook)
}

func (service *Service[TSpec, TStatus]) callHook(kind ServiceHookKind, oldObject *Object[TSpec, TStatus], newObject *Object[TSpec, TStatus]) error {
	if _, exists := service.hooks[kind]; !exists {
		return nil
	}

	oldObjectCopy := deep.MustCopy(oldObject)
	newObjectCopy := deep.MustCopy(newObject)

	for _, hook := range service.hooks[kind] {
		err := hook(oldObjectCopy, newObjectCopy)
		if err != nil {
			return fmt.Errorf("failed to call %s hook: %w", kind, err)
		}
	}

	return nil
}

func WithServiceHook[TSpec Specification, TStatus Status](kind ServiceHookKind, hook ServiceHook[TSpec, TStatus]) ServiceOpt[TSpec, TStatus] {
	return func(service *Service[TSpec, TStatus]) {
		service.attachHook(kind, hook)
	}
}
