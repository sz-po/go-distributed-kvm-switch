package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestError_Error(t *testing.T) {
	error := NewError("foo", 500)
	assert.Equal(t, "foo", error.Error())
	assert.Equal(t, 500, error.HttpCode())
}
