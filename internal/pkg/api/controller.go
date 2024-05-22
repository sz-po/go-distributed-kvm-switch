package api

import (
	"errors"
	"log/slog"
	"os"
	"time"
)

type Controller[TSpec Specification, TStatus Status, TInstance any] interface {
	InitInstance(object *Object[TSpec, TStatus]) (*TInstance, error)
	ReconcileInstance(object *Object[TSpec, TStatus], instance *TInstance) (*TStatus, error)
	ShutdownInstance(instance *TInstance) error
}

func WithController[TSpec Specification, TStatus Status, TInstance any](controller Controller[TSpec, TStatus, TInstance], tick <-chan time.Time) ServiceOpt[TSpec, TStatus] {
	return func(service *Service[TSpec, TStatus]) {
		service.attachHook(AfterCreate, func(oldObject *Object[TSpec, TStatus], newObject *Object[TSpec, TStatus]) error {
			objectName := newObject.Metadata.Name
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})).With(slog.String("module", "api.controller"), slog.String("objectName", string(objectName)))

			initFn := func(object *Object[TSpec, TStatus]) *TInstance {
				var instance *TInstance
				var err error

				for range tick {
					logger.Debug("Initiating controller.")
					instance, err = controller.InitInstance(object)
					if err != nil {
						logger.Error("Failed to initiate controller.", slog.String("error", err.Error()))
						continue
					}
					break
				}

				logger.Debug("Controller initiated.")
				return instance
			}

			shutdownFn := func(instance *TInstance) {
				logger.Debug("Trying to shutdown controller.")
				for range tick {
					err := controller.ShutdownInstance(instance)
					if err != nil {
						logger.Warn("Failed to shutdown controller. Retrying.", slog.String("error", err.Error()))
						continue
					}
					break
				}
				logger.Debug("Controller is shut down.")
			}

			pruneFn := func(objectName ObjectName) {
				for range tick {
					logger.Debug("Pruning object.")
					err := service.Prune(objectName)
					if err != nil && !errors.Is(err, ErrObjectNotFound) {
						logger.Warn("Failed to prune object. Retrying.", slog.String("error", err.Error()))
						continue
					}
					break
				}
			}

			go func() {
				logger.Debug("Starting controller.")
				defer logger.Debug("Controller finished.")

				instance := initFn(newObject)

				for range tick {
					object, err := service.Get(objectName, WithDeleted())

					if errors.Is(err, ErrObjectNotFound) || (object != nil && object.IsDeleted()) {
						logger.Debug("Object not found.")
						break
					}

					if err != nil {
						logger.Warn("Failed to get object. Retrying.", slog.String("error", err.Error()))
						continue
					}

					status, err := controller.ReconcileInstance(object, instance)
					if err != nil {
						logger.Warn("Failed to reconcile. Retrying.", slog.String("error", err.Error()))
						continue
					}

					if status != nil {
						_, err = service.UpdateStatus(objectName, *status)
						if err != nil {
							logger.Warn("Failed to update status. Retrying.", slog.String("error", err.Error()))
							continue
						}
					}
				}

				shutdownFn(instance)
				pruneFn(objectName)

			}()
			return nil
		})
	}
}
