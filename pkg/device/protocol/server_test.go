package protocol

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	serviceProvider := &mockedServiceProvider{}
	serviceProvider.On("GetServiceName").Return(ServiceName("foo"))
	serviceProvider.On("GetMethods").Return(map[MethodName]ServiceMethod{
		"echo": func(ctx context.Context, payload any) (any, error) {
			return payload, nil
		},
		"multiply": func(ctx context.Context, payload any) (any, error) {
			return payload.(float64) * 2, nil
		},
		"error": func(ctx context.Context, payload any) (any, error) {
			return nil, errors.New("some error")
		},
	})

	serviceRegistry := NewServiceRegistry(WithServiceProvider(serviceProvider))

	clientPipe, serverPipe := net.Pipe()
	clientEncoder := json.NewEncoder(clientPipe)
	clientDecoder := json.NewDecoder(clientPipe)
	jsonServer := NewServer(serverPipe, serviceRegistry)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	err := jsonServer.Start(ctx, wg)
	assert.NoError(t, err)

	/*
	 * Call non-existing service
	 */

	err = clientEncoder.Encode(MethodCallMessage{
		CallId:      CallId("call-1"),
		ServiceName: ServiceName("non-existing"),
		MethodName:  MethodName("non-existing"),
		Payload:     nil,
	})
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	methodResult := MethodResultMessage{}
	err = clientDecoder.Decode(&methodResult)
	assert.NoError(t, err)
	assert.Equal(t, CallId("call-1"), methodResult.CallId)
	assert.Equal(t, ErrServiceNotFound.Error(), *methodResult.Error)
	assert.Nil(t, methodResult.Payload)

	/*
	 * Call existing service but non-existing method
	 */

	err = clientEncoder.Encode(MethodCallMessage{
		CallId:      CallId("call-2"),
		ServiceName: ServiceName("foo"),
		MethodName:  MethodName("non-existing"),
		Payload:     nil,
	})
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	methodResult = MethodResultMessage{}
	err = clientDecoder.Decode(&methodResult)
	assert.NoError(t, err)
	assert.Equal(t, CallId("call-2"), methodResult.CallId)
	assert.Equal(t, ErrServiceMethodNotFound.Error(), *methodResult.Error)
	assert.Nil(t, methodResult.Payload)

	/*
	 * Call existing service and "echo" method
	 */

	err = clientEncoder.Encode(MethodCallMessage{
		CallId:      CallId("call-3"),
		ServiceName: ServiceName("foo"),
		MethodName:  MethodName("echo"),
		Payload:     "hello",
	})
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	methodResult = MethodResultMessage{}
	err = clientDecoder.Decode(&methodResult)
	assert.NoError(t, err)
	assert.Equal(t, CallId("call-3"), methodResult.CallId)
	assert.Nil(t, methodResult.Error)
	assert.Equal(t, "hello", methodResult.Payload)

	/*
	 * Call existing service and "multiply" method
	 */

	err = clientEncoder.Encode(MethodCallMessage{
		CallId:      CallId("call-4"),
		ServiceName: ServiceName("foo"),
		MethodName:  MethodName("multiply"),
		Payload:     float64(5),
	})
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	methodResult = MethodResultMessage{}
	err = clientDecoder.Decode(&methodResult)
	assert.NoError(t, err)
	assert.Equal(t, CallId("call-4"), methodResult.CallId)
	assert.Nil(t, methodResult.Error)
	assert.Equal(t, float64(10), methodResult.Payload)

	/*
	 * Call existing service and "error" method
	 */

	err = clientEncoder.Encode(MethodCallMessage{
		CallId:      CallId("call-5"),
		ServiceName: ServiceName("foo"),
		MethodName:  MethodName("error"),
		Payload:     nil,
	})
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	methodResult = MethodResultMessage{}
	err = clientDecoder.Decode(&methodResult)
	assert.NoError(t, err)
	assert.Equal(t, CallId("call-5"), methodResult.CallId)
	assert.Equal(t, errors.New("some error").Error(), *methodResult.Error)
	assert.Nil(t, methodResult.Payload)

	/*
	 * Closing the server pipe
	 */
	serverPipe.Close()
	time.Sleep(50 * time.Millisecond)

	cancel()
	wg.Wait()
}
