package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProcessHandler_Create(t *testing.T) {
	serviceMock := &process.ServiceMock{}
	handler := NewProcessHandler(serviceMock)

	processSpecification := process.Specification{
		ExecutablePath: "/bin/sleep",
		Arguments: []string{
			"5",
		},
	}

	processStatus := process.Status{
		Phase:   process.Running,
		Enabled: true,
	}

	processMock := &process.ProcessMock{}

	processMock.On("GetSpecification").Return(processSpecification)
	processMock.On("GetStatus").Return(processStatus)

	processName := process.Name("foo")

	serviceMock.On("CreateProcess", mock.Anything, processName, processSpecification).Return(processMock, nil)
	serviceMock.On("CreateProcess", mock.Anything, process.Name("invalid"), processSpecification).Return(nil, fmt.Errorf("invalid"))

	requestBody, err := json.Marshal(CreateRequest{
		Specification: processSpecification,
	})
	assert.NoError(t, err)

	responseBody, err := json.Marshal(CreateResponse{
		Specification: processSpecification,
		Status:        processStatus,
	})

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodPost, "/foo", bytes.NewReader(requestBody))
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)

	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, responseBody, response.Body.Bytes())

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/foo", bytes.NewBufferString("non-json"))
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Zero(t, response.Body.Len())

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/invalid", bytes.NewReader(requestBody))
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Zero(t, response.Body.Len())
}

func TestProcessHandler_GetByName(t *testing.T) {
	serviceMock := &process.ServiceMock{}
	handler := NewProcessHandler(serviceMock)

	processMock := &process.ProcessMock{}

	processSpecification := process.Specification{
		ExecutablePath: "/bin/sleep",
		Arguments: []string{
			"5",
		},
	}

	processStatus := process.Status{
		Phase:   process.Running,
		Enabled: true,
	}

	processMock.On("GetSpecification").Return(processSpecification)
	processMock.On("GetStatus").Return(processStatus)
	processName := process.Name("foo")

	serviceMock.On("GetProcessByName", processName).Return(processMock, nil)
	serviceMock.On("GetProcessByName", process.Name("invalid")).Return(nil, fmt.Errorf("other error"))
	serviceMock.On("GetProcessByName", mock.Anything).Return(nil, process.ErrProcessNotFound)

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/foo", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)

	responseBody, err := json.Marshal(GetByNameResponse{
		Name:          processName,
		Specification: processSpecification,
		Status:        processStatus,
	})

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, responseBody, response.Body.Bytes())

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodGet, "/bar", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodGet, "/invalid", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestProcessHandler_Delete(t *testing.T) {
	serviceMock := &process.ServiceMock{}
	handler := NewProcessHandler(serviceMock)

	processName := process.Name("foo")

	serviceMock.On("DeleteProcess", mock.Anything, processName).Return(nil)
	serviceMock.On("DeleteProcess", mock.Anything, process.Name("invalid")).Return(fmt.Errorf("other error"))
	serviceMock.On("DeleteProcess", mock.Anything, mock.Anything).Return(process.ErrProcessNotFound)

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodDelete, "/foo", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Zero(t, response.Body.Len())

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodDelete, "/bar", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodDelete, "/invalid", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestProcessHandler_Find(t *testing.T) {
	serviceMock := &process.ServiceMock{}
	handler := NewProcessHandler(serviceMock)

	processMock := &process.ProcessMock{}

	processSpecification := process.Specification{
		ExecutablePath: "/bin/sleep",
		Arguments: []string{
			"5",
		},
	}

	processStatus := process.Status{
		Phase:   process.Running,
		Enabled: true,
	}

	processMock.On("GetSpecification").Return(processSpecification)
	processMock.On("GetStatus").Return(processStatus)
	processName := process.Name("foo")

	serviceMock.On("FindProcess").Return(map[process.Name]process.Process{
		processName: processMock,
	})

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)

	responseBody, err := json.Marshal(FindResponse{
		Processes: []FindResponseProcess{
			{
				Name:          processName,
				Specification: processSpecification,
				Status:        processStatus,
			},
		},
		Count: 1,
	})

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, responseBody, response.Body.Bytes())
}

func TestProcessHandler_Disable(t *testing.T) {
	serviceMock := &process.ServiceMock{}
	handler := NewProcessHandler(serviceMock)

	validProcessMock := &process.ProcessMock{}
	validProcessMock.On("Disable", mock.Anything).Return(nil)

	disabledProcessMock := &process.ProcessMock{}
	disabledProcessMock.On("Disable", mock.Anything).Return(process.ErrProcessAlreadyDisabled)

	invalidProcessMock := &process.ProcessMock{}
	invalidProcessMock.On("Disable", mock.Anything).Return(fmt.Errorf("other error"))

	serviceMock.On("GetProcessByName", process.Name("valid")).Return(validProcessMock, nil)
	serviceMock.On("GetProcessByName", process.Name("disabled")).Return(disabledProcessMock, nil)
	serviceMock.On("GetProcessByName", process.Name("invalid")).Return(invalidProcessMock, nil)
	serviceMock.On("GetProcessByName", process.Name("service-malfunction")).Return(nil, fmt.Errorf("service-malfunction error"))
	serviceMock.On("GetProcessByName", mock.Anything).Return(invalidProcessMock, process.ErrProcessNotFound)

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodPost, "/valid/disable", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/disabled/disable", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusConflict, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/invalid/disable", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/service-malfunction/disable", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/other/disable", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestProcessHandler_Enable(t *testing.T) {
	serviceMock := &process.ServiceMock{}
	handler := NewProcessHandler(serviceMock)

	validProcessMock := &process.ProcessMock{}
	validProcessMock.On("Enable", mock.Anything).Return(nil)

	enabledProcessMock := &process.ProcessMock{}
	enabledProcessMock.On("Enable", mock.Anything).Return(process.ErrProcessAlreadyEnabled)

	invalidProcessMock := &process.ProcessMock{}
	invalidProcessMock.On("Enable", mock.Anything).Return(fmt.Errorf("other error"))

	serviceMock.On("GetProcessByName", process.Name("valid")).Return(validProcessMock, nil)
	serviceMock.On("GetProcessByName", process.Name("enabled")).Return(enabledProcessMock, nil)
	serviceMock.On("GetProcessByName", process.Name("invalid")).Return(invalidProcessMock, nil)
	serviceMock.On("GetProcessByName", process.Name("service-malfunction")).Return(nil, fmt.Errorf("service-malfunction error"))
	serviceMock.On("GetProcessByName", mock.Anything).Return(invalidProcessMock, process.ErrProcessNotFound)

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodPost, "/valid/enable", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/enabled/enable", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusConflict, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/invalid/enable", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/service-malfunction/enable", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/other/enable", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestProcessHandler_Restart(t *testing.T) {
	serviceMock := &process.ServiceMock{}
	handler := NewProcessHandler(serviceMock)

	validProcessMock := &process.ProcessMock{}
	validProcessMock.On("Restart", mock.Anything).Return(nil)

	invalidProcessMock := &process.ProcessMock{}
	invalidProcessMock.On("Restart", mock.Anything).Return(fmt.Errorf("other error"))

	serviceMock.On("GetProcessByName", process.Name("valid")).Return(validProcessMock, nil)
	serviceMock.On("GetProcessByName", process.Name("invalid")).Return(invalidProcessMock, nil)
	serviceMock.On("GetProcessByName", process.Name("service-malfunction")).Return(nil, fmt.Errorf("service-malfunction error"))
	serviceMock.On("GetProcessByName", mock.Anything).Return(invalidProcessMock, process.ErrProcessNotFound)

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodPost, "/valid/restart", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/invalid/restart", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/service-malfunction/restart", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	response = httptest.NewRecorder()
	request, err = http.NewRequest(http.MethodPost, "/other/restart", nil)
	assert.NoError(t, err)

	handler.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
}
