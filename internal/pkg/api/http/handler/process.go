package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
	"log/slog"
	"net/http"
)

type ProcessHandler struct {
	service *process.Service
	router  *chi.Mux
	logger  *slog.Logger
}

func NewProcessHandler(service *process.Service) *ProcessHandler {
	router := chi.NewRouter()

	handler := &ProcessHandler{
		service: service,
		router:  router,
		logger: slog.Default().With(
			slog.String("componentName", "api.http.handler.ProcessHandler"),
		),
	}

	router.Post("/{name}", handler.Create)
	router.Delete("/{name}", handler.Delete)
	router.Get("/{name}", handler.Get)
	router.Get("/", handler.Find)
	router.Post("/{name}/restart", handler.Restart)
	router.Post("/{name}/enable", handler.Enable)
	router.Post("/{name}/disable", handler.Disable)

	return handler
}

func (handler *ProcessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.router.ServeHTTP(w, r)
}

func (handler *ProcessHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger := handler.logger.With(slog.String("handlerName", "Create"))

	var request CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Error("Failed to decode request body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	processName := process.Name(chi.URLParam(r, "name"))

	process, err := handler.service.Create(r.Context(), processName, request.Specification)
	if err != nil {
		logger.Error("Failed to create process", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := CreateResponse{
		Specification: process.GetSpecification(),
		Status:        process.GetStatus(),
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", slog.String("error", err.Error()))
	}
}

func (handler *ProcessHandler) Delete(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (handler *ProcessHandler) Get(w http.ResponseWriter, r *http.Request) {
	logger := handler.logger.With(slog.String("handlerName", "GetStatus"))

	processName := process.Name(chi.URLParam(r, "name"))

	processInstance, err := handler.service.Get(processName)
	if err != nil {
		logger.Error("Failed to get process", slog.String("error", err.Error()))
		if errors.Is(err, process.ErrProcessNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	response := GetResponse{
		Name:          processName,
		Specification: processInstance.GetSpecification(),
		Status:        processInstance.GetStatus(),
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", slog.String("error", err.Error()))
	}
}

func (handler *ProcessHandler) Find(w http.ResponseWriter, r *http.Request) {
	logger := handler.logger.With(slog.String("handlerName", "Find"))

	processes := handler.service.Find()

	response := FindResponse{
		Processes: []FindResponseProcess{},
		Count:     len(processes),
	}

	for processName, process := range processes {
		response.Processes = append(response.Processes, FindResponseProcess{
			Name:          processName,
			Specification: process.GetSpecification(),
			Status:        process.GetStatus(),
		})
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", slog.String("error", err.Error()))
	}
}
func (handler *ProcessHandler) Restart(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (handler *ProcessHandler) Enable(w http.ResponseWriter, r *http.Request) {
	logger := handler.logger.With(slog.String("handlerName", "Enable"))

	processName := process.Name(chi.URLParam(r, "name"))

	processInstance, err := handler.service.Get(processName)
	if err != nil {
		logger.Error("Failed to get process", slog.String("error", err.Error()))
		if errors.Is(err, process.ErrProcessNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	err = processInstance.Enable(r.Context())
	if err != nil {
		logger.Error("Failed to enable process", slog.String("error", err.Error()))
		if errors.Is(err, process.ErrProcessAlreadyEnabled) {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}

func (handler *ProcessHandler) Disable(w http.ResponseWriter, r *http.Request) {
	logger := handler.logger.With(slog.String("handlerName", "Disable"))

	processName := process.Name(chi.URLParam(r, "name"))

	processInstance, err := handler.service.Get(processName)
	if err != nil {
		logger.Error("Failed to get process", slog.String("error", err.Error()))
		if errors.Is(err, process.ErrProcessNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	err = processInstance.Disable(r.Context())
	if err != nil {
		logger.Error("Failed to disable process", slog.String("error", err.Error()))
		if errors.Is(err, process.ErrProcessAlreadyDisabled) {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}
