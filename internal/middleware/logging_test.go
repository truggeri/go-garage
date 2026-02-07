package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/pkg/applog"
)

func TestRequestLogger_LogsRequestDetails(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	vehicleLog := applog.BuildVehicleAppLog("info", "json", logBuffer)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := RequestLogger(vehicleLog)(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	recorder := httptest.NewRecorder()
	loggedHandler.ServeHTTP(recorder, req)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "go-garage web request", "Log should contain go-garage web request tag")
	assert.Contains(t, logOutput, "GET", "Log should contain HTTP method")
	assert.Contains(t, logOutput, "/test-path", "Log should contain request path")
	assert.Contains(t, logOutput, "200", "Log should contain status code")
}

func TestRequestLogger_CapturesStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"BadRequest", http.StatusBadRequest},
		{"NotFound", http.StatusNotFound},
		{"InternalError", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer := &bytes.Buffer{}
			vehicleLog := applog.BuildVehicleAppLog("info", "json", logBuffer)

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			loggedHandler := RequestLogger(vehicleLog)(testHandler)
			req := httptest.NewRequest(http.MethodPost, "/endpoint", nil)
			recorder := httptest.NewRecorder()
			loggedHandler.ServeHTTP(recorder, req)

			logOutput := logBuffer.String()
			var logEntry map[string]interface{}
			err := json.Unmarshal([]byte(logOutput), &logEntry)
			require.NoError(t, err)
			assert.Equal(t, float64(tt.statusCode), logEntry["response_status"])
		})
	}
}

func TestRequestLogger_PassesRequestThrough(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	vehicleLog := applog.BuildVehicleAppLog("info", "json", logBuffer)

	responseBody := "test response body"
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(responseBody))
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	loggedHandler := RequestLogger(vehicleLog)(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	loggedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, responseBody, recorder.Body.String())
}

func TestRequestLogger_DifferentHTTPMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			logBuffer := &bytes.Buffer{}
			vehicleLog := applog.BuildVehicleAppLog("info", "json", logBuffer)

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			loggedHandler := RequestLogger(vehicleLog)(testHandler)
			req := httptest.NewRequest(method, "/api/vehicles", nil)
			recorder := httptest.NewRecorder()
			loggedHandler.ServeHTTP(recorder, req)

			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, method, "Log should contain HTTP method")
			assert.Contains(t, logOutput, "/api/vehicles", "Log should contain path")
		})
	}
}
