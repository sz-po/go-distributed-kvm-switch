package process

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewService(t *testing.T) {
	service := NewService()
	assert.NotNil(t, service)
}

func TestNewService_BeforeObjectCreatedHook(t *testing.T) {
	service := NewService()
	assert.NotNil(t, service)

	obj, err := service.Create("non-existing", Specification{Execution: ExecutionSpecification{
		ExecutablePath: "/bin/non-existing-executable",
	}})
	assert.Nil(t, obj)
	assert.ErrorContains(t, err, "no such file or directory")

	obj, err = service.Create("existing", Specification{Execution: ExecutionSpecification{
		ExecutablePath: "/bin/bash",
		Arguments:      []string{"-c", "echo 'hello world'"},
	}})
	assert.NotNil(t, obj)
	assert.NoError(t, err)
}
