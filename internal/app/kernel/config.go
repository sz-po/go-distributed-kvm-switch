package kernel

import (
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api/http"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/device"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
)

type ServiceConfig struct {
	Process process.LocalServiceConfig `kong:"embed,prefix=process-"`
	Device  device.ServiceConfig       `kong:"embed,prefix=device-"`
}

type ApiConfig struct {
	Http http.Config `kong:"embed,prefix=http-"`
}

type LogConfig struct {
	Level  string
	Format string
}

type Config struct {
	Api     ApiConfig     `kong:"embed,prefix=api-"`
	Service ServiceConfig `kong:"embed,prefix=service-"`
}
