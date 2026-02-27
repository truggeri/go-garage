package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// csrfTokenContextKey is the context key for the CSRF token.
const csrfTokenContextKey contextKey = "csrfToken"

// csrfCookieName is the name of the cookie that stores the CSRF token.
const csrfCookieName = "csrf_token"

// csrfFormField is the name of the hidden form field that carries the CSRF token.
const csrfFormField = "csrf_token"

// csrfTokenLength is the number of random bytes used to generate a token (32 bytes = 64 hex chars).
const csrfTokenLength = 32

// csrfCookieMaxAge is the lifetime of the CSRF cookie in seconds (5 minutes).
const csrfCookieMaxAge = 5 * 60

// GetCSRFToken retrieves the CSRF token stored in the request context.
// Returns an empty string when no token is present.
func GetCSRFToken(ctx context.Context) string {
	v, _ := ctx.Value(csrfTokenContextKey).(string)
	return v
}

// CSRFProtection creates middleware that implements the double-submit cookie
// pattern for CSRF protection. On every request it ensures a csrf_token cookie
// exists and stores the token in the request context so that handlers can pass
// it to templates. On state-changing requests (POST, PUT, DELETE) it validates
// that the form field value matches the cookie value.
func CSRFProtection() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := readOrCreateToken(w, r)

			ctx := context.WithValue(r.Context(), csrfTokenContextKey, token)
			r = r.WithContext(ctx)

			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
				formToken := r.FormValue(csrfFormField)
				if formToken == "" || formToken != token {
					http.Error(w, "Forbidden - invalid CSRF token", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// readOrCreateToken reads the CSRF token from the cookie or generates a new
// one, setting it as a cookie on the response.
func readOrCreateToken(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(csrfCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	token := generateToken()
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   csrfCookieMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.TLS != nil,
	})
	return token
}

// generateToken creates a cryptographically random hex-encoded token.
func generateToken() string {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		// Fallback should never happen in practice; crypto/rand reads from the OS.
		panic("csrf: failed to generate random token: " + err.Error())
	}
	return hex.EncodeToString(b)
}
