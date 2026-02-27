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
// that the form field token was signed with the same session ID and secret.
func CSRFProtection(secret string) func(http.Handler) http.Handler {
	secretBytes := []byte(secret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID := sessionIDFromContext(r.Context())

			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
				formToken := r.FormValue(csrfFormField)
				if !validateCSRFToken(formToken, sessionID, secretBytes) {
					http.Error(w, "Forbidden - invalid CSRF token", http.StatusForbidden)
					return
				}
			}

			token := generateCSRFToken(sessionID, secretBytes)
			ctx := context.WithValue(r.Context(), csrfTokenContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
// The signature is HMAC-SHA256(sessionID + nonce, secret).
func generateCSRFToken(sessionID string, secret []byte) string {
	nonce := generateNonce()
	sig := computeHMAC(sessionID+nonce, secret)
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
	expectedSig := computeHMAC(sessionID+nonce, secret)

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
