package applog

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

// VehicleAppLog provides logging capabilities specifically designed for the go-garage vehicle management system
type VehicleAppLog struct {
	baseLogger      *slog.Logger
	outputWriter    io.Writer
	minimumSeverity slog.Level
	useJSONEncoding bool
}

// BuildVehicleAppLog constructs a new logger for the go-garage application
func BuildVehicleAppLog(severity, encoding string, writer io.Writer) *VehicleAppLog {
	if writer == nil {
		writer = os.Stdout
	}

	severityLevel := convertSeverityStringToLevel(severity)
	jsonMode := (encoding == "json")

	handlerOpts := &slog.HandlerOptions{
		Level: severityLevel,
	}

	var baseHandler slog.Handler
	if jsonMode {
		baseHandler = slog.NewJSONHandler(writer, handlerOpts)
	} else {
		baseHandler = slog.NewTextHandler(writer, handlerOpts)
	}

	return &VehicleAppLog{
		baseLogger:      slog.New(baseHandler),
		outputWriter:    writer,
		minimumSeverity: severityLevel,
		useJSONEncoding: jsonMode,
	}
}

// RecordInfo writes an informational log entry
func (v *VehicleAppLog) RecordInfo(message string, keyValuePairs ...any) {
	v.baseLogger.Info(message, keyValuePairs...)
}

// RecordDebug writes a debug log entry
func (v *VehicleAppLog) RecordDebug(message string, keyValuePairs ...any) {
	v.baseLogger.Debug(message, keyValuePairs...)
}

// RecordWarning writes a warning log entry
func (v *VehicleAppLog) RecordWarning(message string, keyValuePairs ...any) {
	v.baseLogger.Warn(message, keyValuePairs...)
}

// RecordError writes an error log entry
func (v *VehicleAppLog) RecordError(message string, keyValuePairs ...any) {
	v.baseLogger.Error(message, keyValuePairs...)
}

// RecordHTTPActivity logs HTTP request/response activity with go-garage specific context
func (v *VehicleAppLog) RecordHTTPActivity(verb, urlPath string, statusCode int, durationMS int64, clientIP string) {
	v.baseLogger.Info("go-garage web request",
		"http_method", verb,
		"url_path", urlPath,
		"response_status", statusCode,
		"processing_time_ms", durationMS,
		"client_ip", clientIP,
		"timestamp", time.Now().Unix(),
	)
}

// RecordPanicEvent logs panic recovery events with full details
func (v *VehicleAppLog) RecordPanicEvent(verb, urlPath string, panicValue any, stackTrace string) {
	v.baseLogger.Error("go-garage panic recovered",
		"http_method", verb,
		"url_path", urlPath,
		"panic_value", fmt.Sprintf("%v", panicValue),
		"stack_trace", stackTrace,
		"timestamp", time.Now().Unix(),
	)
}

// RecordAppStartup logs application startup information
func (v *VehicleAppLog) RecordAppStartup(environment, host string, port int) {
	v.baseLogger.Info("go-garage server starting",
		"environment", environment,
		"host", host,
		"port", port,
		"timestamp", time.Now().Unix(),
	)
}

// RecordAppShutdown logs application shutdown information
func (v *VehicleAppLog) RecordAppShutdown(reason string) {
	v.baseLogger.Info("go-garage server shutting down",
		"reason", reason,
		"timestamp", time.Now().Unix(),
	)
}

// convertSeverityStringToLevel maps severity names to slog levels
func convertSeverityStringToLevel(severity string) slog.Level {
	levelMap := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	if level, exists := levelMap[severity]; exists {
		return level
	}
	return slog.LevelInfo
}
