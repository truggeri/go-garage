package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	healthCheckHandler(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected HTTP 200 status code")
	
	contentType := recorder.Header().Get("Content-Type")
	require.Equal(t, "application/json", contentType, "Expected JSON content type header")

	responseBody := recorder.Body.String()
	assert.Contains(t, responseBody, `"status":"healthy"`, "Response should indicate healthy status")
	assert.Contains(t, responseBody, `"timestamp"`, "Response should include timestamp field")
}

func TestGetEnvOrDefaultWithValue(t *testing.T) {
	t.Setenv("TEST_VAR_KEY", "custom_value")
	
	result := getEnvOrDefault("TEST_VAR_KEY", "fallback_value")
	
	assert.Equal(t, "custom_value", result, "Should return environment variable value when set")
}

func TestGetEnvOrDefaultWithoutValue(t *testing.T) {
	result := getEnvOrDefault("NONEXISTENT_VAR_KEY", "default_value")
	
	assert.Equal(t, "default_value", result, "Should return default value when environment variable not set")
}

func TestHealthCheckResponseFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	healthCheckHandler(recorder, req)

	body := recorder.Body.String()
	assert.True(t, strings.HasPrefix(body, "{"), "Response should start with opening brace")
	assert.True(t, strings.HasSuffix(body, "}"), "Response should end with closing brace")
}
