package source

import (
	"fmt"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/device"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

const DevicesPath = "devices"

type DirectorySourceConfig struct {
	RootPath string `kong:"required"`
}

type DirectorySource struct {
	config DirectorySourceConfig
	logger *slog.Logger
}

func NewDirectorySource(config DirectorySourceConfig) (*DirectorySource, error) {
	if rootPathInfo, err := os.Stat(config.RootPath); err != nil {
		return nil, fmt.Errorf("failed to stat root path: %w", err)
	} else if !rootPathInfo.IsDir() {
		return nil, fmt.Errorf("root path is not a directory")
	}

	return &DirectorySource{
		config: config,
		logger: slog.Default().With(slog.String("componentName", "loader.DirectorySource")),
	}, nil
}

func (source *DirectorySource) GetAvailableDevices() (map[device.Name]device.Specification, error) {
	devicesPath := path.Join(source.config.RootPath, DevicesPath)

	source.logger.Debug("Scanning for available devices config.", slog.String("devicesPath", devicesPath))

	devicesFilesGlob := path.Join(devicesPath, "*.yml")

	devicesFiles, err := filepath.Glob(devicesFilesGlob)
	if err != nil {
		return nil, fmt.Errorf("failed to glob devices files: %w", err)
	}

	devices := map[device.Name]device.Specification{}

	for _, deviceFile := range devicesFiles {
		deviceName := device.Name(strings.TrimSuffix(path.Base(deviceFile), ".yml"))

		deviceSpecificationBuffer, err := os.ReadFile(deviceFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read device config file: %w", err)
		}

		var deviceSpecification device.Specification

		if err = yaml.Unmarshal(deviceSpecificationBuffer, &deviceSpecification); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device config: %w", err)
		}

		devices[deviceName] = deviceSpecification
	}

	return devices, nil
}
