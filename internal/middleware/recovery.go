package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// RecoverFromPanic creates middleware that recovers from panics in Go-Garage handlers
// It logs the panic details and returns a 500 error response
func RecoverFromPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Go-Garage PANIC] Request: %s %s - Error: %v\nStack Trace:\n%s",
					r.Method,
					r.URL.Path,
					err,
					string(debug.Stack()),
				)

				errorResponse := map[string]string{
					"error":   "Internal server error",
					"message": "An unexpected error occurred",
				}
				responseJSON, _ := json.Marshal(errorResponse)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(responseJSON)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
