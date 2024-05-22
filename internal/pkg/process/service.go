package process

import (
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api"
	"log/slog"
	"os"
	"time"
)

type Service struct {
	*api.Service[Specification, Status]

	logger *slog.Logger
}

func NewService() *Service {
	store := api.NewMemoryObjectStore[Specification, Status]()

	service := &Service{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(slog.String("module", "process.service")),
	}

	service.Service = api.NewService[Specification, Status](store,
		api.WithDefaults[Specification, Status](),
		api.WithImmutableSpecification[Specification, Status](),
		api.WithServiceHook[Specification, Status](api.BeforeCreate, service.beforeObjectCreated),
		api.WithController[Specification, Status, Runner](NewController(), time.NewTicker(200*time.Millisecond).C),
	)

	return service
}

func (service *Service) beforeObjectCreated(oldObject *api.Object[Specification, Status], newObject *api.Object[Specification, Status]) error {
	_, err := os.Stat(newObject.Specification.Execution.ExecutablePath)
	if err != nil {
		return err
	}

	return nil
}
