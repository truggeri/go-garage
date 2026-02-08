package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/truggeri/go-garage/pkg/applog"
)

// RecoverFromPanic creates middleware that recovers from panics in Go-Garage handlers
// It logs the panic details and returns a 500 error response
func RecoverFromPanic(vehicleLog *applog.VehicleAppLog) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if panicErr := recover(); panicErr != nil {
					vehicleLog.RecordPanicEvent(
						r.Method,
						r.URL.Path,
						panicErr,
						string(debug.Stack()),
					)
					jsonPayload := []byte(`{"error":"Internal server error","message":"An unexpected error occurred"}`)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					//nolint:errcheck
					w.Write(jsonPayload)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
