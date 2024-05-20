package api

type objectStoreQuery struct {
	nameFilter  string
	withDeleted bool
}

type ObjectStoreQuery func(*objectStoreQuery)

type ObjectStore[TSpec Specification, TStatus Status] interface {
	Create(ObjectName, TSpec) error
	UpdateSpecification(ObjectName, TSpec) error
	UpdateStatus(ObjectName, TStatus) error
	Get(ObjectName, ...ObjectStoreQuery) (*Object[TSpec, TStatus], error)
	Delete(name ObjectName) error
	Prune(name ObjectName) error
	Find(...ObjectStoreQuery) []ObjectName
}

type EventStoreQuery func(Event) bool

type EventStore interface {
	Add(Event) error
	Find(EventStoreQuery) []Event
}

func (query *objectStoreQuery) apply(queryOpts []ObjectStoreQuery) {
	for _, queryOpt := range queryOpts {
		queryOpt(query)
	}
}

func WhereObjectName(name string) ObjectStoreQuery {
	return func(query *objectStoreQuery) {
		query.nameFilter = name
	}
}

func WithDeleted() ObjectStoreQuery {
	return func(query *objectStoreQuery) {
		query.withDeleted = true
	}
}
