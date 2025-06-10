package source

import (
	"github.com/stretchr/testify/assert"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/device"
	"os"
	"path"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestNewDirectorySource(t *testing.T) {
	source, err := NewDirectorySource(DirectorySourceConfig{
		RootPath: "invalid-path",
	})
	assert.Nil(t, source)
	assert.Error(t, err)

	source, err = NewDirectorySource(DirectorySourceConfig{
		RootPath: "/bin/sleep",
	})
	assert.Nil(t, source)
	assert.Error(t, err)

	source, err = NewDirectorySource(DirectorySourceConfig{
		RootPath: "/tmp",
	})
	assert.NotNil(t, source)
	assert.NoError(t, err)
}

func TestDirectorySource_GetAvailableDevices(t *testing.T) {
	configPath := t.TempDir()
	devicesPath := path.Join(configPath, DevicesPath)
	err := os.Mkdir(devicesPath, 0755)
	assert.NoError(t, err)

	source, err := NewDirectorySource(DirectorySourceConfig{
		RootPath: configPath,
	})
	assert.NoError(t, err)
	assert.NotNil(t, source)

	devices, err := source.GetAvailableDevices()
	assert.NoError(t, err)
	assert.Empty(t, devices)

	testSpecification := device.Specification{
		Kind: "test",
		Config: map[string]any{
			"foo": "bar",
		},
	}

	testSpecificationBuffer, err := yaml.Marshal(testSpecification)
	assert.NoError(t, err)
	err = os.WriteFile(path.Join(devicesPath, "foo.yml"), testSpecificationBuffer, 0644)
	assert.NoError(t, err)

	devices, err = source.GetAvailableDevices()
	assert.NoError(t, err)
	assert.Equal(t, map[device.Name]device.Specification{
		"foo": testSpecification,
	}, devices)

	err = os.WriteFile(path.Join(devicesPath, "invalid.xyz"), []byte("invalid"), 0644)
	assert.NoError(t, err)

	devices, err = source.GetAvailableDevices()
	assert.NoError(t, err)
	assert.Equal(t, map[device.Name]device.Specification{
		"foo": testSpecification,
	}, devices)

	err = os.WriteFile(path.Join(devicesPath, "invalid.yml"), []byte("invalid"), 0644)
	assert.NoError(t, err)

	devices, err = source.GetAvailableDevices()
	assert.Error(t, err)
	assert.Empty(t, devices)

}
