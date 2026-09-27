package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

// csrfTokenContextKey is the context key for the CSRF token.
const csrfTokenContextKey contextKey = "csrfToken"

// csrfFormField is the name of the hidden form field that carries the CSRF token.
const csrfFormField = "csrf_token"

// csrfHeaderName is the name of the request header that carries the CSRF token
// for requests without a form-encoded body (e.g. htmx JSON requests).
const csrfHeaderName = "X-CSRF-Token"

// csrfNonceLength is the number of random bytes used to generate a nonce (32 bytes = 64 hex chars).
const csrfNonceLength = 32

// csrfTokenSeparator separates the nonce from the HMAC signature in the token string.
const csrfTokenSeparator = "."

// GetCSRFToken retrieves the CSRF token stored in the request context.
// Returns an empty string when no token is present.
func GetCSRFToken(ctx context.Context) string {
	v, _ := ctx.Value(csrfTokenContextKey).(string)
	return v
}

// CSRFProtection creates middleware that implements HMAC-based CSRF protection.
// On every request it generates a new CSRF token using the user's session ID
// (from AccountInfo if present) and a cryptographic nonce, signed with the
// provided secret. On state-changing requests (POST, PUT, DELETE) it validates
// that the token supplied in the X-CSRF-Token header or the form field was
// signed with the same session ID and secret.
func CSRFProtection(secret string) func(http.Handler) http.Handler {
	return csrfMiddleware(secret, func(*http.Request) bool { return true }, writePageCSRFError)
}

// APICSRFProtection creates middleware that applies CSRF protection to REST API
// requests, but only when the request was authenticated with session cookies.
// Bearer-authenticated requests are not vulnerable to classic CSRF because
// browsers do not automatically attach an Authorization header, so they are
// left unaffected. Failures use the standard JSON error envelope.
// It must be mounted after the authentication middleware so the authentication
// mechanism and account are available in the request context.
func APICSRFProtection(secret string) func(http.Handler) http.Handler {
	return csrfMiddleware(secret, isCookieAuthenticated, writeAPICSRFError)
}

// csrfMiddleware builds CSRF middleware that validates state-changing requests
// selected by shouldEnforce and reports failures with writeError.
func csrfMiddleware(
	secret string,
	shouldEnforce func(*http.Request) bool,
	writeError func(http.ResponseWriter),
) func(http.Handler) http.Handler {
	secretBytes := []byte(secret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID := sessionIDFromContext(r.Context())

			if isStateChanging(r.Method) && shouldEnforce(r) {
				if !validateCSRFToken(csrfTokenFromRequest(r), sessionID, secretBytes) {
					writeError(w)
					return
				}
			}

			token := generateCSRFToken(sessionID, secretBytes)
			ctx := context.WithValue(r.Context(), csrfTokenContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isStateChanging reports whether an HTTP method mutates server state and
// therefore requires CSRF validation.
func isStateChanging(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
}

// isCookieAuthenticated reports whether the request was authenticated with
// session cookies rather than an Authorization bearer token.
func isCookieAuthenticated(r *http.Request) bool {
	method, ok := GetAuthMethodFromContext(r.Context())
	return ok && method == AuthMethodCookie
}

// csrfTokenFromRequest extracts the CSRF token supplied by the client. The
// X-CSRF-Token header is preferred so JSON requests are supported; full-page
// form submissions fall back to the hidden csrf_token form field.
func csrfTokenFromRequest(r *http.Request) string {
	if token := r.Header.Get(csrfHeaderName); token != "" {
		return token
	}
	return r.FormValue(csrfFormField)
}

// writePageCSRFError reports a CSRF failure for browser page routes.
func writePageCSRFError(w http.ResponseWriter) {
	http.Error(w, "Forbidden - invalid CSRF token", http.StatusForbidden)
}

// writeAPICSRFError reports a CSRF failure using the standard JSON error envelope.
func writeAPICSRFError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"success":false,"error":{"code":"CSRF_ERROR","message":"invalid CSRF token"}}`))
}

// sessionIDFromContext returns the authenticated user's ID when available,
// or an empty string for unauthenticated requests (e.g. login/register pages).
func sessionIDFromContext(ctx context.Context) string {
	acct, ok := GetAccountFromContext(ctx)
	if ok && acct != nil {
		return acct.ID
	}
	return ""
}

// generateCSRFToken creates an HMAC-based CSRF token in the format "nonce.signature".
// The signature is HMAC-SHA256(sessionID | nonce, secret) where | is a delimiter
// to prevent concatenation ambiguity.
func generateCSRFToken(sessionID string, secret []byte) string {
	nonce := generateNonce()
	sig := computeHMAC(sessionID+"|"+nonce, secret)
	return nonce + csrfTokenSeparator + sig
}

// validateCSRFToken checks that the provided token has a valid HMAC signature
// for the given session ID and secret.
func validateCSRFToken(token, sessionID string, secret []byte) bool {
	parts := strings.SplitN(token, csrfTokenSeparator, 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}

	nonce, providedSig := parts[0], parts[1]
	expectedSig := computeHMAC(sessionID+"|"+nonce, secret)

	return hmac.Equal([]byte(providedSig), []byte(expectedSig))
}

// computeHMAC returns a hex-encoded HMAC-SHA256 of the given message.
func computeHMAC(message string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// generateNonce creates a cryptographically random hex-encoded nonce.
func generateNonce() string {
	b := make([]byte, csrfNonceLength)
	if _, err := rand.Read(b); err != nil {
		panic("csrf: failed to generate random nonce: " + err.Error())
	}
	return hex.EncodeToString(b)
}
