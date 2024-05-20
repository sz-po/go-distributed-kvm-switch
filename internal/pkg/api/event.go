package api

type EventKind string

type Event interface {
	Kind() EventKind
	Message() string
	RelatedObjects() []ObjectRef
}
