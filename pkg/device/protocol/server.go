package protocol

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"go.openly.dev/pointy"
	"io"
	"log/slog"
	"sync"
)

type ServerOpt func(*Server)

type Server struct {
	pipe            io.ReadWriter
	serviceRegistry *ServiceRegistry

	callQueue   chan MethodCallMessage
	resultQueue chan MethodResultMessage

	logger *slog.Logger
}

func NewServer(pipe io.ReadWriter, serviceRegistry *ServiceRegistry, opts ...ServerOpt) *Server {
	server := &Server{
		pipe:            pipe,
		serviceRegistry: serviceRegistry,

		callQueue:   make(chan MethodCallMessage, 16),
		resultQueue: make(chan MethodResultMessage, 16),

		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(server)
	}

	server.logger = server.logger.With(
		slog.String("componentName", "ProtocolServer"),
	)

	return server
}

func (server *Server) Start(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	go server.readLoop(ctx, wg, server.pipe)

	wg.Add(1)
	go server.writeLoop(ctx, wg, server.pipe)

	wg.Add(1)
	go server.executionLoop(ctx, wg)

	wg.Add(1)
	go func() {
		<-ctx.Done()

		close(server.callQueue)
		close(server.resultQueue)
		wg.Done()
	}()

	return nil
}

func (server *Server) readLoop(ctx context.Context, wg *sync.WaitGroup, reader io.Reader) {
	defer wg.Done()

	server.logger.Debug("Starting read loop.")

	decoder := json.NewDecoder(reader)

	for decoder.More() {
		message := MessageEnvelope{}

		if err := decoder.Decode(&message); errors.Is(err, io.EOF) {
			server.logger.Debug("Reached end of stream while reading.")
			return
		} else if err != nil {
			server.logger.Warn("Failed to decode message.", slog.String("error", err.Error()))
			continue
		}

		if message.Call == nil {
			continue
		}

		server.callQueue <- *message.Call
	}
}

func (server *Server) writeLoop(ctx context.Context, wg *sync.WaitGroup, writer io.Writer) {
	defer wg.Done()

	server.logger.Debug("Starting write loop.")

	for {
		select {
		case methodResult := <-server.resultQueue:
			if err := json.NewEncoder(writer).Encode(MessageEnvelope{
				Result: &methodResult,
			}); err != nil {
				server.logger.Warn("Failed to encode method result.", slog.String("error", err.Error()))
			}
			if bufioReadWriter, ok := server.pipe.(*bufio.ReadWriter); ok {
				bufioReadWriter.Flush()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (server *Server) executionLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	server.logger.Debug("Starting execution loop.")

	for {
		select {
		case methodCall := <-server.callQueue:
			logger := server.logger.With(
				slog.String("serviceName", string(methodCall.ServiceName)),
				slog.String("methodName", string(methodCall.MethodName)),
			)

			payload, err := server.serviceRegistry.CallServiceMethod(ctx, methodCall.ServiceName, methodCall.MethodName, methodCall.Payload)
			if err != nil {
				server.resultQueue <- MethodResultMessage{
					CallId: methodCall.CallId,
					Error:  pointy.String(err.Error()),
				}
				logger.Warn("Failed to call service method.", slog.String("error", err.Error()))
			} else {
				server.resultQueue <- MethodResultMessage{
					CallId:  methodCall.CallId,
					Payload: payload,
				}
				logger.Debug("Successfully called service method.")
			}

		case <-ctx.Done():
			return
		}
	}
}
