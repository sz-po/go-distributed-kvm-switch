package api

import "fmt"

func WithImmutableSpecification[TSpec Specification, TStatus Status]() ServiceOpt[TSpec, TStatus] {
	return func(service *Service[TSpec, TStatus]) {
		service.attachHook(BeforeSpecificationUpdate, func(oldObject *Object[TSpec, TStatus], newObject *Object[TSpec, TStatus]) error {
			return ErrObjectSpecificationIsImmutable
		})
	}
}

var ErrObjectSpecificationIsImmutable = fmt.Errorf("object specification is immutable")
