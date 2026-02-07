package applog

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildVehicleAppLog_JSONMode(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("info", "json", buffer)

	require.NotNil(t, appLogger)
	assert.True(t, appLogger.useJSONEncoding)
	assert.Equal(t, buffer, appLogger.outputWriter)
}

func TestBuildVehicleAppLog_TextMode(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("info", "text", buffer)

	require.NotNil(t, appLogger)
	assert.False(t, appLogger.useJSONEncoding)
}

func TestBuildVehicleAppLog_WithNilWriter(t *testing.T) {
	appLogger := BuildVehicleAppLog("info", "json", nil)

	require.NotNil(t, appLogger)
	assert.NotNil(t, appLogger.outputWriter)
}

func TestRecordInfo_CreatesLogEntry(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("info", "json", buffer)

	appLogger.RecordInfo("vehicle added", "vin", "ABC123", "make", "Toyota")

	logOutput := buffer.String()
	assert.Contains(t, logOutput, "vehicle added")
	assert.Contains(t, logOutput, "vin")
	assert.Contains(t, logOutput, "ABC123")

	var parsedLog map[string]interface{}
	err := json.Unmarshal([]byte(logOutput), &parsedLog)
	require.NoError(t, err)
	assert.Equal(t, "INFO", parsedLog["level"])
}

func TestRecordDebug_CreatesDebugEntry(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("debug", "json", buffer)

	appLogger.RecordDebug("checking maintenance", "vehicle_id", 42)

	logOutput := buffer.String()
	assert.Contains(t, logOutput, "checking maintenance")

	var parsedLog map[string]interface{}
	err := json.Unmarshal([]byte(logOutput), &parsedLog)
	require.NoError(t, err)
	assert.Equal(t, "DEBUG", parsedLog["level"])
}

func TestRecordWarning_CreatesWarnEntry(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("warn", "json", buffer)

	appLogger.RecordWarning("maintenance overdue", "days_late", 30)

	logOutput := buffer.String()
	assert.Contains(t, logOutput, "maintenance overdue")

	var parsedLog map[string]interface{}
	err := json.Unmarshal([]byte(logOutput), &parsedLog)
	require.NoError(t, err)
	assert.Equal(t, "WARN", parsedLog["level"])
}

func TestRecordError_CreatesErrorEntry(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("error", "json", buffer)

	appLogger.RecordError("database connection failed", "error", "timeout")

	logOutput := buffer.String()
	assert.Contains(t, logOutput, "database connection failed")

	var parsedLog map[string]interface{}
	err := json.Unmarshal([]byte(logOutput), &parsedLog)
	require.NoError(t, err)
	assert.Equal(t, "ERROR", parsedLog["level"])
}

func TestRecordHTTPActivity_LogsRequestDetails(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("info", "json", buffer)

	appLogger.RecordHTTPActivity("POST", "/api/vehicles", 201, 45, "192.168.1.100")

	logOutput := buffer.String()
	assert.Contains(t, logOutput, "go-garage web request")
	assert.Contains(t, logOutput, "POST")
	assert.Contains(t, logOutput, "/api/vehicles")
	assert.Contains(t, logOutput, "201")
	assert.Contains(t, logOutput, "192.168.1.100")

	var parsedLog map[string]interface{}
	err := json.Unmarshal([]byte(logOutput), &parsedLog)
	require.NoError(t, err)
	assert.Equal(t, "POST", parsedLog["http_method"])
	assert.Equal(t, "/api/vehicles", parsedLog["url_path"])
	assert.Equal(t, float64(201), parsedLog["response_status"])
}

func TestRecordPanicEvent_LogsPanicDetails(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("error", "json", buffer)

	appLogger.RecordPanicEvent("GET", "/api/crash", "null pointer", "stack trace here")

	logOutput := buffer.String()
	assert.Contains(t, logOutput, "go-garage panic recovered")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/api/crash")
	assert.Contains(t, logOutput, "null pointer")

	var parsedLog map[string]interface{}
	err := json.Unmarshal([]byte(logOutput), &parsedLog)
	require.NoError(t, err)
	assert.Equal(t, "ERROR", parsedLog["level"])
}

func TestRecordAppStartup_LogsStartupInfo(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("info", "json", buffer)

	appLogger.RecordAppStartup("production", "0.0.0.0", 8080)

	logOutput := buffer.String()
	assert.Contains(t, logOutput, "go-garage server starting")
	assert.Contains(t, logOutput, "production")
	assert.Contains(t, logOutput, "8080")

	var parsedLog map[string]interface{}
	err := json.Unmarshal([]byte(logOutput), &parsedLog)
	require.NoError(t, err)
	assert.Equal(t, "production", parsedLog["environment"])
	assert.Equal(t, float64(8080), parsedLog["port"])
}

func TestRecordAppShutdown_LogsShutdownInfo(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("info", "json", buffer)

	appLogger.RecordAppShutdown("SIGTERM received")

	logOutput := buffer.String()
	assert.Contains(t, logOutput, "go-garage server shutting down")
	assert.Contains(t, logOutput, "SIGTERM received")

	var parsedLog map[string]interface{}
	err := json.Unmarshal([]byte(logOutput), &parsedLog)
	require.NoError(t, err)
	assert.Equal(t, "SIGTERM received", parsedLog["reason"])
}

func TestSeverityFiltering_DebugHidden(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("info", "json", buffer)

	appLogger.RecordDebug("this is hidden")
	appLogger.RecordInfo("this is visible")

	logOutput := buffer.String()
	assert.NotContains(t, logOutput, "this is hidden")
	assert.Contains(t, logOutput, "this is visible")
}

func TestSeverityFiltering_ErrorLevelOnly(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("error", "json", buffer)

	appLogger.RecordDebug("debug hidden")
	appLogger.RecordInfo("info hidden")
	appLogger.RecordWarning("warning hidden")
	appLogger.RecordError("error visible")

	logOutput := buffer.String()
	logLines := strings.Split(strings.TrimSpace(logOutput), "\n")
	assert.Len(t, logLines, 1)
	assert.Contains(t, logOutput, "error visible")
}

func TestTextFormatOutput(t *testing.T) {
	buffer := &bytes.Buffer{}
	appLogger := BuildVehicleAppLog("info", "text", buffer)

	appLogger.RecordInfo("test message", "key1", "value1")

	logOutput := buffer.String()
	assert.Contains(t, logOutput, "test message")
	assert.Contains(t, logOutput, "key1")
	assert.Contains(t, logOutput, "value1")
	assert.Contains(t, logOutput, "level=INFO")
}

func TestConvertSeverityStringToLevel_AllLevels(t *testing.T) {
	testCases := []struct {
		inputSeverity string
		expectsDebug  bool
		expectsInfo   bool
		expectsWarn   bool
		expectsError  bool
	}{
		{"debug", true, true, true, true},
		{"info", false, true, true, true},
		{"warn", false, false, true, true},
		{"error", false, false, false, true},
		{"unknown", false, true, true, true}, // defaults to info
	}

	for _, tc := range testCases {
		t.Run(tc.inputSeverity, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			appLogger := BuildVehicleAppLog(tc.inputSeverity, "json", buffer)

			appLogger.RecordDebug("debug msg")
			hasDebug := strings.Contains(buffer.String(), "debug msg")
			assert.Equal(t, tc.expectsDebug, hasDebug)

			buffer.Reset()
			appLogger.RecordInfo("info msg")
			hasInfo := strings.Contains(buffer.String(), "info msg")
			assert.Equal(t, tc.expectsInfo, hasInfo)

			buffer.Reset()
			appLogger.RecordWarning("warn msg")
			hasWarn := strings.Contains(buffer.String(), "warn msg")
			assert.Equal(t, tc.expectsWarn, hasWarn)

			buffer.Reset()
			appLogger.RecordError("error msg")
			hasError := strings.Contains(buffer.String(), "error msg")
			assert.Equal(t, tc.expectsError, hasError)
		})
	}
}
