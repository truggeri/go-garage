package middleware

import (
	"log"
	"net/http"
	"time"
)

// RequestLogger creates middleware that logs HTTP requests for Go-Garage
// It captures method, path, duration, and status code for each request
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Create custom response writer to capture status code
		// Default is 200 OK per HTTP spec when handlers write without calling WriteHeader
		wrapper := &statusCapture{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process the request
		next.ServeHTTP(wrapper, r)

		// Calculate request duration
		elapsed := time.Since(startTime)

		// Log request details
		log.Printf("[Go-Garage] %s %s - Status: %d - Duration: %v",
			r.Method,
			r.URL.Path,
			wrapper.statusCode,
			elapsed,
		)
	})
}

// statusCapture wraps http.ResponseWriter to capture the status code
type statusCapture struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing it
func (sc *statusCapture) WriteHeader(code int) {
	sc.statusCode = code
	sc.ResponseWriter.WriteHeader(code)
}
