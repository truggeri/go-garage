# Go-Garage Logging

This document describes the logging system used in the Go-Garage vehicle management application.

## Overview

Go-Garage uses Go's native `log/slog` package wrapped in a custom `applog` package designed specifically for vehicle management logging needs. The logging system provides structured logging with configurable formats and severity levels.

## Configuration

Logging is configured through environment variables:

### LOG_LEVEL

Controls the minimum severity level for log output.

**Valid values:**
- `debug` - All log messages (debug, info, warn, error)
- `info` - Informational and higher (info, warn, error) [DEFAULT]
- `warn` - Warnings and errors only (warn, error)
- `error` - Errors only

**Example:**
```bash
export LOG_LEVEL=debug
```

### LOG_FORMAT

Controls the output format of log messages.

**Valid values:**
- `json` - Structured JSON format [DEFAULT]
- `text` - Human-readable text format

**Example:**
```bash
export LOG_FORMAT=text
```

## Log Message Types

### Application Lifecycle

**Startup:**
```json
{
  "time": "2026-02-07T20:24:02.709Z",
  "level": "INFO",
  "msg": "go-garage server starting",
  "environment": "development",
  "host": "0.0.0.0",
  "port": 8080,
  "timestamp": 1770495842
}
```

**Shutdown:**
```json
{
  "time": "2026-02-07T20:24:07.451Z",
  "level": "INFO",
  "msg": "go-garage server shutting down",
  "reason": "Signal terminated received",
  "timestamp": 1770495847
}
```

### HTTP Requests

All HTTP requests are automatically logged with:
- HTTP method
- URL path
- Response status code
- Processing time in milliseconds
- Client IP address

**Example:**
```json
{
  "time": "2026-02-07T20:25:15.664Z",
  "level": "INFO",
  "msg": "go-garage web request",
  "http_method": "GET",
  "url_path": "/health",
  "response_status": 200,
  "processing_time_ms": 0,
  "client_ip": "[::1]:51360",
  "timestamp": 1770495915
}
```

### Panic Recovery

When a panic occurs in a request handler, it's automatically caught and logged:

```json
{
  "time": "2026-02-07T20:25:20.123Z",
  "level": "ERROR",
  "msg": "go-garage panic recovered",
  "http_method": "POST",
  "url_path": "/api/vehicles",
  "panic_value": "null pointer dereference",
  "stack_trace": "goroutine 1 [running]:\n...",
  "timestamp": 1770495920
}
```

## Usage in Code

### Using the Logger

The logger is initialized in `main.go` and passed to middleware and handlers:

```go
import "github.com/truggeri/go-garage/pkg/applog"

// Create logger from configuration
vehicleLog := applog.BuildVehicleAppLog(cfg.Logging.Level, cfg.Logging.Format, nil)

// Log informational messages
vehicleLog.RecordInfo("vehicle created", "vin", "ABC123", "make", "Toyota")

// Log debug messages
vehicleLog.RecordDebug("processing request", "step", "validation")

// Log warnings
vehicleLog.RecordWarning("slow query detected", "duration_ms", 1500)

// Log errors
vehicleLog.RecordError("database connection failed", "error", err.Error())
```

### Log Methods

- `RecordInfo(message, keyValuePairs...)` - Informational messages
- `RecordDebug(message, keyValuePairs...)` - Debug messages
- `RecordWarning(message, keyValuePairs...)` - Warning messages
- `RecordError(message, keyValuePairs...)` - Error messages

### Specialized Methods

The `VehicleAppLog` provides specialized methods for common go-garage logging scenarios:

- `RecordHTTPActivity(verb, path, statusCode, durationMS, clientIP)` - HTTP request logging
- `RecordPanicEvent(verb, path, panicValue, stackTrace)` - Panic recovery logging
- `RecordAppStartup(environment, host, port)` - Application startup logging
- `RecordAppShutdown(reason)` - Application shutdown logging

## Text Format Example

When `LOG_FORMAT=text`, logs appear in human-readable format:

```
time=2026-02-07T20:25:06.610Z level=INFO msg="go-garage server starting" environment=development host=0.0.0.0 port=8080 timestamp=1770495906
time=2026-02-07T20:25:06.610Z level=INFO msg="Server listening" address=0.0.0.0:8080
time=2026-02-07T20:25:08.536Z level=INFO msg="go-garage server shutting down" reason="Signal terminated received" timestamp=1770495908
```

## Log Rotation

Log rotation is handled externally by the deployment environment. Common options:

### Docker/Docker Compose

Configure logging driver in `docker-compose.yml`:

```yaml
services:
  go-garage:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### Systemd

Use journald for log rotation:

```ini
[Service]
StandardOutput=journal
StandardError=journal
```

### Kubernetes

Use container log rotation or log aggregation solutions like:
- Fluentd
- Logstash
- CloudWatch Logs
- Stackdriver

### Manual Log Rotation

If running directly, use `logrotate`:

```
/var/log/go-garage/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 go-garage go-garage
}
```

## Best Practices

1. **Use appropriate log levels:**
   - `debug` - Detailed information for diagnosing problems
   - `info` - General informational messages about application flow
   - `warn` - Warning messages for potentially harmful situations
   - `error` - Error messages for error events

2. **Include context:**
   Always include relevant context as key-value pairs:
   ```go
   vehicleLog.RecordError("vehicle not found", "vin", vin, "user_id", userID)
   ```

3. **Don't log sensitive data:**
   Never log passwords, tokens, credit card numbers, or PII

4. **Use structured logging:**
   Always use key-value pairs instead of string interpolation for better searchability

5. **Production settings:**
   - Use `LOG_LEVEL=info` in production (default)
   - Use `LOG_FORMAT=json` in production (default) for easier parsing

## Performance Considerations

- The logging system is optimized for performance
- Log statements below the configured level are filtered efficiently
- JSON encoding is performed only for messages that will be output
- HTTP request logging has minimal overhead (<1ms per request)

## Troubleshooting

### Logs not appearing

Check the `LOG_LEVEL` setting. If set to `error`, only error messages will appear.

### Incorrect format

Verify `LOG_FORMAT` is set to either `json` or `text`. Invalid values default to `json`.

### Missing HTTP request logs

Ensure the `RequestLogger` middleware is properly configured in the HTTP handler chain.

### Log file growing too large

Configure log rotation using your deployment platform's logging system or external tools like `logrotate`.
