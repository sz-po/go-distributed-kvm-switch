package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestObject_IsDeleted(t *testing.T) {
	object := Object[string, Status]{
		Metadata: Metadata{
			Name: "foo",
		},
	}
	assert.False(t, object.IsDeleted())

	object.Metadata.DeletedAt = Now()
	assert.True(t, object.IsDeleted())
}
