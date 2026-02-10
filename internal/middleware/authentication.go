package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/truggeri/go-garage/internal/auth"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// AccountContextKey is the key used to store account info in request context
const AccountContextKey contextKey = "accountInfo"

// AccountInfo holds authenticated user information extracted from JWT
type AccountInfo struct {
	ID   string
	Name string
}

// GetAccountFromContext retrieves the authenticated account from request context
func GetAccountFromContext(ctx context.Context) (*AccountInfo, bool) {
	acct, ok := ctx.Value(AccountContextKey).(*AccountInfo)
	return acct, ok
}

// AuthenticationGuard creates middleware that validates JWT tokens
// It extracts the Bearer token from the Authorization header and validates it
func AuthenticationGuard(tokenMgr *auth.TokenManager) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, "missing authorization header")
				return
			}

			headerParts := strings.SplitN(authHeader, " ", 2)
			if len(headerParts) != 2 || !strings.EqualFold(headerParts[0], "Bearer") {
				writeAuthError(w, "invalid authorization format")
				return
			}

			tokenString := headerParts[1]
			verified, err := tokenMgr.ValidateToken(tokenString)
			if err != nil {
				writeAuthError(w, "invalid or expired token")
				return
			}

			if verified.TokenKind != auth.AccessTokenKind {
				writeAuthError(w, "access token required; refresh tokens cannot be used for API requests")
				return
			}

			acctInfo := &AccountInfo{
				ID:   verified.AccountID,
				Name: verified.AccountName,
			}

			enrichedCtx := context.WithValue(r.Context(), AccountContextKey, acctInfo)
			nextHandler.ServeHTTP(w, r.WithContext(enrichedCtx))
		})
	}
}

// writeAuthError writes a JSON error response for authentication failures
func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"success":false,"error":{"code":"AUTHENTICATION_ERROR","message":"` + message + `"}}`))
}
