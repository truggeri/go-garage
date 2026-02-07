package middleware

import (
	"net/http"
	"time"

	"github.com/truggeri/go-garage/pkg/applog"
)

// RequestLogger creates middleware that logs HTTP requests for Go-Garage
// It captures method, path, duration, and status code for each request
func RequestLogger(vehicleLog *applog.VehicleAppLog) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestStart := time.Now()

			respRecorder := &statusCapture{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(respRecorder, r)

			processingDuration := time.Since(requestStart)

			vehicleLog.RecordHTTPActivity(
				r.Method,
				r.URL.Path,
				respRecorder.statusCode,
				processingDuration.Milliseconds(),
				r.RemoteAddr,
			)
		})
	}
}

type statusCapture struct {
	http.ResponseWriter
	statusCode int
}

func (sc *statusCapture) WriteHeader(code int) {
	sc.statusCode = code
	sc.ResponseWriter.WriteHeader(code)
}
