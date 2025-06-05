package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	ListenPort    int    `kong:"default=8080"`
	ListenAddress string `kong:"default=127.0.0.1"`

	IdleTimeout     time.Duration `kong:"default=30s"`
	ReadTimeout     time.Duration `kong:"default=10s"`
	WriteTimeout    time.Duration `kong:"default=10s"`
	ShutdownTimeout time.Duration `kong:"default=30s"`
}

type ServerOpt func(server *Server)

type Server struct {
	router   *chi.Mux
	protocol *http.Server
	logger   *slog.Logger
	config   Config
}

func NewServer(config Config, opts ...ServerOpt) *Server {
	router := chi.NewRouter()

	protocol := &http.Server{
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	server := &Server{
		router:   router,
		protocol: protocol,
		logger:   slog.Default().WithGroup("api.http.server"),
		config:   config,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)

	s.logger.Info("Starting HTTP server",
		slog.Int("listenPort", s.config.ListenPort),
		slog.String("listenAddress", s.config.ListenAddress),
	)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.ListenAddress, s.config.ListenPort))
	if err != nil {
		wg.Done()
		return fmt.Errorf("failed to create listener: %w", err)
	}

	go func() {
		defer wg.Done()
		if err := s.protocol.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Failed to start HTTP server", slog.String("error", err.Error()))
		}
	}()

	go func() {
		<-ctx.Done()
		s.logger.Info("Gracefully shutting down HTTP server.")

		ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		if err := s.protocol.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to gracefully shutdown HTTP server", slog.String("error", err.Error()))
		}
	}()

	return nil
}

func WithHandler(path string, handler http.Handler) ServerOpt {
	return func(server *Server) {
		server.router.Mount(path, handler)
	}
}
