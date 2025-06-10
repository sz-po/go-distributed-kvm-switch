package loader

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api/utils"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/device"
	"sync"
	"testing"
	"time"
)

type deviceSourceMock struct {
	mock.Mock
}

func (source *deviceSourceMock) GetAvailableDevices() (map[device.Name]device.Specification, error) {
	args := source.Called()

	devices, _ := args.Get(0).(map[device.Name]device.Specification)
	err := args.Error(1)

	return devices, err
}

func TestNewDeviceLoader_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}

	deviceSpecification := device.Specification{
		Kind: "test",
		Config: map[string]any{
			"foo": "bar",
		},
	}

	deviceMock := &device.DeviceMock{}
	deviceMock.On("GetSpecification").Return(deviceSpecification)

	serviceMock := device.ServiceMock{}
	serviceMock.On("HasDevice", device.Name("foo")).Once().Return(false)
	serviceMock.On("HasDevice", device.Name("foo")).Return(true)
	serviceMock.On("CreateDevice", mock.Anything, device.Name("foo"), deviceSpecification).Return(deviceMock, nil)

	sourceMock := &deviceSourceMock{}
	sourceMock.On("GetAvailableDevices").Return(map[device.Name]device.Specification{
		"foo": deviceSpecification,
	}, nil)

	loader := NewDeviceLoader(DeviceLoaderConfig{
		ScanInterval: utils.Duration(100 * time.Millisecond),
	}, &serviceMock, sourceMock)

	err := loader.Start(ctx, wg)
	assert.NoError(t, err)

	serviceMock.AssertCalled(t, "CreateDevice", mock.Anything, device.Name("foo"), deviceSpecification)
}

func TestNewDeviceLoader_Start_SourceError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}

	sourceErr := errors.New("source error")

	serviceMock := device.ServiceMock{}

	sourceMock := &deviceSourceMock{}
	sourceMock.On("GetAvailableDevices").Return(nil, sourceErr)

	loader := NewDeviceLoader(DeviceLoaderConfig{
		ScanInterval: utils.Duration(100 * time.Millisecond),
	}, &serviceMock, sourceMock)

	err := loader.Start(ctx, wg)
	assert.ErrorIs(t, err, sourceErr)

	serviceMock.AssertNotCalled(t, "CreateDevice", mock.Anything, mock.Anything, mock.Anything)
}
