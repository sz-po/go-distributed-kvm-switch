package protocol

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sync"
)

type ClientOpt func(*Client)

type Client struct {
	pipe                  io.ReadWriter
	messageHandlerFn      map[CallId]chan MethodResultMessage
	messageHandlerFnMutex *sync.Mutex

	callQueue   chan MethodCallMessage
	resultQueue chan MethodResultMessage

	logger *slog.Logger
}

func NewClient(pipe io.ReadWriter, opts ...ClientOpt) *Client {
	client := &Client{
		pipe:                  pipe,
		messageHandlerFn:      make(map[CallId]chan MethodResultMessage),
		messageHandlerFnMutex: &sync.Mutex{},

		callQueue:   make(chan MethodCallMessage, 16),
		resultQueue: make(chan MethodResultMessage, 16),

		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(client)
	}

	client.logger = client.logger.With(
		slog.String("componentName", "ProtocolClient"),
	)

	return client
}

func (client *Client) Start(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	go client.readLoop(ctx, wg, client.pipe)

	wg.Add(1)
	go client.writeLoop(ctx, wg, client.pipe)

	wg.Add(1)
	go client.handleResultLoop(ctx, wg)

	wg.Add(1)
	go func() {
		<-ctx.Done()

		close(client.callQueue)
		close(client.resultQueue)
		wg.Done()
	}()

	return nil
}

func (client *Client) Call(ctx context.Context, serviceName ServiceName, methodName MethodName, payload any) (any, error) {
	callId := CreateRandomCallId()

	defer func() {
		client.messageHandlerFnMutex.Lock()
		delete(client.messageHandlerFn, callId)
		client.messageHandlerFnMutex.Unlock()
	}()

	resultCh := make(chan MethodResultMessage)

	client.messageHandlerFnMutex.Lock()
	client.messageHandlerFn[callId] = resultCh
	client.messageHandlerFnMutex.Unlock()

	client.callQueue <- MethodCallMessage{
		CallId:      callId,
		ServiceName: serviceName,
		MethodName:  methodName,
		Payload:     payload,
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultCh:
		if result.Error != nil {
			return nil, errors.New(*result.Error)
		} else {
			return result.Payload, nil
		}
	}
}

// readLoop reads messages from the reader and sends them to the resultQueue.
func (client *Client) readLoop(ctx context.Context, wg *sync.WaitGroup, reader io.Reader) {
	defer wg.Done()

	client.logger.Debug("Starting read loop.")

	decoder := json.NewDecoder(reader)

	for decoder.More() {
		message := MessageEnvelope{}

		if err := decoder.Decode(&message); errors.Is(err, io.EOF) {
			client.logger.Debug("Reached end of stream while reading.")
			return
		} else if err != nil {
			client.logger.Warn("Failed to decode message.", slog.String("error", err.Error()))
			continue
		}

		if message.Result == nil {
			continue
		}

		client.resultQueue <- *message.Result
	}
}

// writeLoop writes messages from the callQueue to the writer.
func (client *Client) writeLoop(ctx context.Context, wg *sync.WaitGroup, writer io.Writer) {
	defer wg.Done()

	client.logger.Debug("Starting write loop.")

	for {
		select {
		case methodCall := <-client.callQueue:
			if err := json.NewEncoder(writer).Encode(MessageEnvelope{
				Call: &methodCall,
			}); err != nil {
				client.logger.Warn("Failed to encode method call.", slog.String("error", err.Error()))
			}
			if bufioReadWriter, ok := client.pipe.(*bufio.ReadWriter); ok {
				bufioReadWriter.Flush()
			}
		case <-ctx.Done():
			return
		}
	}
}

// handleResultLoop handles messages from the resultQueue and sends them to the messageHandlerFns of appropriate call.
func (client *Client) handleResultLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case methodResult := <-client.resultQueue:
			client.messageHandlerFnMutex.Lock()
			if resultCh, ok := client.messageHandlerFn[methodResult.CallId]; ok {
				resultCh <- methodResult
			} else {
				client.logger.Warn("Received result for unknown call.", slog.String("callId", string(methodResult.CallId)))
			}
			client.messageHandlerFnMutex.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
