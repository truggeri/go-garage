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

		wrapper := &statusCapture{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		elapsed := time.Since(startTime)

		log.Printf("[Go-Garage-Web] %s %s - Status: %d - Duration: %v",
			r.Method,
			r.URL.Path,
			wrapper.statusCode,
			elapsed,
		)
	})
}

type statusCapture struct {
	http.ResponseWriter
	statusCode int
}

func (sc *statusCapture) WriteHeader(code int) {
	sc.statusCode = code
	sc.ResponseWriter.WriteHeader(code)
}
