package device

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetadata_Id(t *testing.T) {
	metadata := Metadata{
		Name: "name",
		Kind: "kind",
	}
	assert.Equal(t, Id("name.kind"), metadata.Id())
}
