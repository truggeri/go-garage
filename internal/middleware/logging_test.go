package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestLogger_LogsRequestDetails(t *testing.T) {
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := RequestLogger(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	recorder := httptest.NewRecorder()
	loggedHandler.ServeHTTP(recorder, req)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "[Go-Garage-Web]", "Log should contain Go-Garage-Web tag")
	assert.Contains(t, logOutput, "GET", "Log should contain HTTP method")
	assert.Contains(t, logOutput, "/test-path", "Log should contain request path")
	assert.Contains(t, logOutput, "Status: 200", "Log should contain status code")
	assert.Contains(t, logOutput, "Duration:", "Log should contain duration")
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
			var logBuffer bytes.Buffer
			originalOutput := log.Writer()
			log.SetOutput(&logBuffer)
			defer log.SetOutput(originalOutput)

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			loggedHandler := RequestLogger(testHandler)
			req := httptest.NewRequest(http.MethodPost, "/endpoint", nil)
			recorder := httptest.NewRecorder()
			loggedHandler.ServeHTTP(recorder, req)

			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, "Status:", "Log should contain status label")
		})
	}
}

func TestRequestLogger_PassesRequestThrough(t *testing.T) {
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	responseBody := "test response body"
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(responseBody))
	})

	loggedHandler := RequestLogger(testHandler)
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
			var logBuffer bytes.Buffer
			originalOutput := log.Writer()
			log.SetOutput(&logBuffer)
			defer log.SetOutput(originalOutput)

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			loggedHandler := RequestLogger(testHandler)
			req := httptest.NewRequest(method, "/api/vehicles", nil)
			recorder := httptest.NewRecorder()
			loggedHandler.ServeHTTP(recorder, req)

			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, method, "Log should contain HTTP method")
			assert.Contains(t, logOutput, "/api/vehicles", "Log should contain path")
		})
	}
}
