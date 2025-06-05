package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api/http/handler"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
	"net/http"
)

func NewRouter(processService *process.Service) http.Handler {
	router := chi.NewRouter()

	router.Mount("/process", handler.NewProcessHandler(processService))

	return router
}
