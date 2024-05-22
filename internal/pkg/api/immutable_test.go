package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithImmutableSpecification(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](),
		WithImmutableSpecification[testSpec, testStatus]())

	obj, err := service.Create("foo", testSpec{Foo: "bar"})
	assert.NotNil(t, obj)
	assert.NoError(t, err)

	obj, err = service.UpdateSpecification("foo", testSpec{Foo: "baz"})
	assert.Nil(t, obj)
	assert.ErrorIs(t, err, ErrObjectSpecificationIsImmutable)
}
