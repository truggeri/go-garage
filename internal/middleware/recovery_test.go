package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverFromPanic_CatchesPanic(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	// Create handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic message")
	})

	// Wrap with recovery middleware
	recoveredHandler := RecoverFromPanic(panicHandler)

	// Make request - should not panic
	req := httptest.NewRequest(http.MethodGet, "/panic-endpoint", nil)
	recorder := httptest.NewRecorder()

	// This should not cause the test to fail
	require.NotPanics(t, func() {
		recoveredHandler.ServeHTTP(recorder, req)
	}, "Middleware should catch panic")

	// Verify error response
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Internal server error")
	assert.Contains(t, recorder.Body.String(), "error")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	// Verify panic was logged
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "[Go-Garage PANIC]")
	assert.Contains(t, logOutput, "test panic message")
	assert.Contains(t, logOutput, "Stack Trace:")
}

func TestRecoverFromPanic_NormalRequestPassesThrough(t *testing.T) {
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	// Create normal handler
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with recovery middleware
	recoveredHandler := RecoverFromPanic(normalHandler)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/normal", nil)
	recorder := httptest.NewRecorder()
	recoveredHandler.ServeHTTP(recorder, req)

	// Verify normal response
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "success", recorder.Body.String())

	// Verify no panic was logged
	logOutput := logBuffer.String()
	assert.NotContains(t, logOutput, "PANIC")
}

func TestRecoverFromPanic_LogsRequestDetails(t *testing.T) {
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("error in handler")
	})

	recoveredHandler := RecoverFromPanic(panicHandler)
	req := httptest.NewRequest(http.MethodPost, "/api/vehicles/123", nil)
	recorder := httptest.NewRecorder()
	recoveredHandler.ServeHTTP(recorder, req)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "POST", "Log should contain HTTP method")
	assert.Contains(t, logOutput, "/api/vehicles/123", "Log should contain request path")
	assert.Contains(t, logOutput, "error in handler", "Log should contain panic message")
}

func TestRecoverFromPanic_ReturnsJSONError(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	recoveredHandler := RecoverFromPanic(panicHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	recoveredHandler.ServeHTTP(recorder, req)

	// Verify JSON response
	body := recorder.Body.String()
	assert.Contains(t, body, `"error"`, "Response should contain error field")
	assert.Contains(t, body, `"message"`, "Response should contain message field")
	assert.Contains(t, body, "Internal server error", "Response should indicate server error")
}

func TestRecoverFromPanic_DifferentPanicTypes(t *testing.T) {
	tests := []struct {
		name       string
		panicValue interface{}
	}{
		{"StringPanic", "string panic"},
		{"IntPanic", 42},
		{"ErrorPanic", assert.AnError},
		{"NilPanic", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			originalOutput := log.Writer()
			log.SetOutput(&logBuffer)
			defer log.SetOutput(originalOutput)

			panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(tt.panicValue)
			})

			recoveredHandler := RecoverFromPanic(panicHandler)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			recorder := httptest.NewRecorder()

			require.NotPanics(t, func() {
				recoveredHandler.ServeHTTP(recorder, req)
			})

			assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		})
	}
}
