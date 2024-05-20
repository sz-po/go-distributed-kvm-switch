package api

type ObjectId string

type ObjectKind string

type ObjectName string

type ObjectRef struct {
	Id   ObjectId
	Kind ObjectKind
}

type Metadata struct {
	Name   ObjectName
	Labels map[string]string

	CreatedAt              Timestamp
	SpecificationUpdatedAt Timestamp
	StatusUpdatedAt        Timestamp
	DeletedAt              Timestamp
}

type Specification interface {
}

type Status interface {
}

type Object[TSpec Specification, TStatus Status] struct {
	Metadata      Metadata
	Specification TSpec
	Status        *TStatus
}

func (object *Object[TSpec, TStatus]) IsDeleted() bool {
	return !object.Metadata.DeletedAt.IsEmpty()
}
