package main

import "github.com/sz-po/go-distributed-kvm-switch/pkg/device"

func main() {
	device.Bootstrap[Config]()
}
