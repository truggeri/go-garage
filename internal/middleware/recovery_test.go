package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/pkg/applog"
)

func TestRecoverFromPanic_CatchesPanic(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	vehicleLog := applog.BuildVehicleAppLog("error", "json", logBuffer)

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic message")
	})

	recoveredHandler := RecoverFromPanic(vehicleLog)(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/panic-endpoint", nil)
	recorder := httptest.NewRecorder()

	require.NotPanics(t, func() {
		recoveredHandler.ServeHTTP(recorder, req)
	}, "Middleware should catch panic")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Internal server error")
	assert.Contains(t, recorder.Body.String(), "error")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "go-garage panic recovered")
	assert.Contains(t, logOutput, "test panic message")
	assert.Contains(t, logOutput, "stack_trace")
}

func TestRecoverFromPanic_NormalRequestPassesThrough(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	vehicleLog := applog.BuildVehicleAppLog("error", "json", logBuffer)

	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	recoveredHandler := RecoverFromPanic(vehicleLog)(normalHandler)

	req := httptest.NewRequest(http.MethodGet, "/normal", nil)
	recorder := httptest.NewRecorder()
	recoveredHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "success", recorder.Body.String())

	logOutput := logBuffer.String()
	assert.NotContains(t, logOutput, "panic")
}

func TestRecoverFromPanic_LogsRequestDetails(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	vehicleLog := applog.BuildVehicleAppLog("error", "json", logBuffer)

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("error in handler")
	})

	recoveredHandler := RecoverFromPanic(vehicleLog)(panicHandler)
	req := httptest.NewRequest(http.MethodPost, "/api/vehicles/123", nil)
	recorder := httptest.NewRecorder()
	recoveredHandler.ServeHTTP(recorder, req)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "POST", "Log should contain HTTP method")
	assert.Contains(t, logOutput, "/api/vehicles/123", "Log should contain request path")
	assert.Contains(t, logOutput, "error in handler", "Log should contain panic message")
}

func TestRecoverFromPanic_ReturnsJSONError(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	vehicleLog := applog.BuildVehicleAppLog("error", "json", logBuffer)

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	recoveredHandler := RecoverFromPanic(vehicleLog)(panicHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	recoveredHandler.ServeHTTP(recorder, req)

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
			logBuffer := &bytes.Buffer{}
			vehicleLog := applog.BuildVehicleAppLog("error", "json", logBuffer)

			panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(tt.panicValue)
			})

			recoveredHandler := RecoverFromPanic(vehicleLog)(panicHandler)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			recorder := httptest.NewRecorder()

			require.NotPanics(t, func() {
				recoveredHandler.ServeHTTP(recorder, req)
			})

			assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		})
	}
}

