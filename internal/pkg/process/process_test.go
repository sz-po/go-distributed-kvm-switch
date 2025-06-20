package process

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestName_Validate(t *testing.T) {
	assert.NoError(t, Name("test").Validate())
	assert.NoError(t, Name("test-1").Validate())
	assert.NoError(t, Name("test-1-test").Validate())

	assert.Error(t, Name("TEST-1-test").Validate())
	assert.Error(t, Name("TEST_test").Validate())
}
