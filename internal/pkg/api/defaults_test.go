package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithDefaults(t *testing.T) {
	service := NewService[testSpec, testStatus](NewMemoryObjectStore[testSpec, testStatus](), WithDefaults[testSpec, testStatus]())
	createdObj, err := service.Create("foo", testSpec{Foo: "bar"})

	assert.NoError(t, err)
	assert.Equal(t, "bar", createdObj.Specification.DefaultFoo)
}
