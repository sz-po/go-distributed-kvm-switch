package main

import (
	"github.com/sz-po/go-distributed-kvm-switch/pkg/device"
	keyboard_v1_alpha "github.com/sz-po/go-distributed-kvm-switch/pkg/device/api/keyboard/v1_alpha"
)

func main() {
	keyboardSourceService := NewKeyboardSourceService()

	device.Bootstrap[Config](
		device.WithRuntimeOpt(device.WithServiceProvider(keyboard_v1_alpha.NewSourceServiceProvider(keyboardSourceService))),
	)
}
