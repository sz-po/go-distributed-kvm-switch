package api

import "github.com/creasty/defaults"

func WithDefaults[TSpec Specification, TStatus Status]() ServiceOpt[TSpec, TStatus] {
	return func(service *Service[TSpec, TStatus]) {
		service.attachMutator(WhileCreatingObject, func(object *Object[TSpec, TStatus]) (*Object[TSpec, TStatus], error) {
			defaults.MustSet(&object.Specification)

			return object, nil
		})
	}
}
