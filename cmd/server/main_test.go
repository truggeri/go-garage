package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestHealthCheckResponseFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	healthCheckHandler(recorder, req)

	body := recorder.Body.Bytes()

	// Verify response is valid JSON by unmarshaling it
	var result map[string]string
	err := json.Unmarshal(body, &result)
	require.NoError(t, err, "Response body should be valid JSON")

	// Verify expected fields exist in the JSON
	_, hasStatus := result["status"]
	_, hasTimestamp := result["timestamp"]
	assert.True(t, hasStatus, "JSON should contain status field")
	assert.True(t, hasTimestamp, "JSON should contain timestamp field")
}
