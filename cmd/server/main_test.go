package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/database"
)

func setupTestGarage(t *testing.T) *database.SQLiteGarage {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	garage, err := database.InitializeGarage(dbPath, database.StandardWorkerPoolSettings())
	require.NoError(t, err)

	// Run migrations
	migrationsPath := "../../migrations"
	err = database.BootstrapSchema(context.Background(), garage, migrationsPath)
	require.NoError(t, err)

	return garage
}

func TestHealthCheckEndpoint(t *testing.T) {
	garage := setupTestGarage(t)
	defer garage.Terminate()

	handler := createHealthCheckHandler(garage)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected HTTP 200 status code")

	contentType := recorder.Header().Get("Content-Type")
	require.Equal(t, "application/json", contentType, "Expected JSON content type header")

	responseBody := recorder.Body.String()
	assert.Contains(t, responseBody, `"status":"healthy"`, "Response should indicate healthy status")
	assert.Contains(t, responseBody, `"database":"healthy"`, "Response should indicate database is healthy")
	assert.Contains(t, responseBody, `"timestamp"`, "Response should include timestamp field")
}

func TestHealthCheckResponseFormat(t *testing.T) {
	garage := setupTestGarage(t)
	defer garage.Terminate()

	handler := createHealthCheckHandler(garage)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler(recorder, req)

	body := recorder.Body.Bytes()

	// Verify response is valid JSON by unmarshaling it
	var result map[string]string
	err := json.Unmarshal(body, &result)
	require.NoError(t, err, "Response body should be valid JSON")

	// Verify expected fields exist in the JSON
	_, hasStatus := result["status"]
	_, hasDatabase := result["database"]
	_, hasTimestamp := result["timestamp"]
	assert.True(t, hasStatus, "JSON should contain status field")
	assert.True(t, hasDatabase, "JSON should contain database field")
	assert.True(t, hasTimestamp, "JSON should contain timestamp field")
}

func TestHealthCheckUnhealthyDatabase(t *testing.T) {
	garage := setupTestGarage(t)
	garage.Terminate() // Close the database to make it unhealthy

	handler := createHealthCheckHandler(garage)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler(recorder, req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "Expected HTTP 503 when database is unhealthy")

	responseBody := recorder.Body.String()
	assert.Contains(t, responseBody, `"database":"unhealthy"`, "Response should indicate database is unhealthy")
}
