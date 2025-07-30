package protocol

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"go.openly.dev/pointy"
	"log/slog"
	"net"
	"sync"
	"testing"
)

func TestClient_Call_WithResult(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	client, serverEncoder, serverDecoder, cancel := setupClient(t)
	defer cancel()

	var callResult any
	var callErr error
	callDone := make(chan struct{})

	go func() {
		callResult, callErr = client.Call(context.Background(), "foo", "bar", "baz")
		close(callDone)
	}()

	var message MessageEnvelope
	err := serverDecoder.Decode(&message)
	assert.NoError(t, err)
	assert.NotNil(t, message.Call)
	assert.Equal(t, ServiceName("foo"), message.Call.ServiceName)
	assert.Equal(t, MethodName("bar"), message.Call.MethodName)
	assert.Equal(t, "baz", message.Call.Payload)

	callId := message.Call.CallId

	message = MessageEnvelope{
		Result: &MethodResultMessage{
			CallId:  callId,
			Payload: "qux",
		},
	}
	err = serverEncoder.Encode(message)
	assert.NoError(t, err)

	message = MessageEnvelope{
		Result: &MethodResultMessage{
			CallId:  callId,
			Payload: "qux-2",
		},
	}
	err = serverEncoder.Encode(message)
	assert.NoError(t, err)

	<-callDone
	assert.NoError(t, callErr)
	assert.Equal(t, "qux", callResult)
}

func TestClient_Call_WithError(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	client, serverEncoder, serverDecoder, cancel := setupClient(t)
	defer cancel()

	var callResult any
	var callErr error
	callDone := make(chan struct{})

	go func() {
		callResult, callErr = client.Call(context.Background(), "foo", "bar", "baz")
		close(callDone)
	}()

	var message MessageEnvelope
	err := serverDecoder.Decode(&message)
	assert.NoError(t, err)
	assert.NotNil(t, message.Call)
	assert.Equal(t, ServiceName("foo"), message.Call.ServiceName)
	assert.Equal(t, MethodName("bar"), message.Call.MethodName)
	assert.Equal(t, "baz", message.Call.Payload)

	callId := message.Call.CallId

	message = MessageEnvelope{
		Result: &MethodResultMessage{
			CallId: callId,
			Error:  pointy.String("some error"),
		},
	}
	err = serverEncoder.Encode(message)
	assert.NoError(t, err)

	<-callDone
	assert.Equal(t, "some error", callErr.Error())
	assert.Nil(t, callResult)
}

func TestClient_Call_WithContextCancel(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	client, serverEncoder, serverDecoder, cancel := setupClient(t)
	defer cancel()

	var callResult any
	var callErr error
	callDone := make(chan struct{})

	callCtx, callCancel := context.WithCancel(context.Background())

	go func() {
		callResult, callErr = client.Call(callCtx, "foo", "bar", "baz")
		close(callDone)
	}()

	var message MessageEnvelope
	err := serverDecoder.Decode(&message)
	assert.NoError(t, err)
	assert.NotNil(t, message.Call)
	assert.Equal(t, ServiceName("foo"), message.Call.ServiceName)
	assert.Equal(t, MethodName("bar"), message.Call.MethodName)
	assert.Equal(t, "baz", message.Call.Payload)

	callCancel()

	callId := message.Call.CallId

	message = MessageEnvelope{
		Result: &MethodResultMessage{
			CallId: callId,
			Error:  pointy.String("some error"),
		},
	}
	err = serverEncoder.Encode(message)
	assert.NoError(t, err)

	<-callDone
	assert.ErrorIs(t, context.Canceled, callErr)
	assert.Nil(t, callResult)
}

func setupClient(t *testing.T) (*Client, *json.Encoder, *json.Decoder, context.CancelFunc) {
	clientPipe, serverPipe := net.Pipe()
	serverEncoder := json.NewEncoder(serverPipe)
	serverDecoder := json.NewDecoder(serverPipe)
	client := NewClient(clientPipe)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	err := client.Start(ctx, wg)
	assert.NoError(t, err)

	return client, serverEncoder, serverDecoder, cancel
}
