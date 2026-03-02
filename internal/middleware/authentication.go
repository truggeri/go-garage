package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

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

// CookieAuthGuard creates middleware that validates JWT tokens from the access_token cookie.
// When the access token is missing or expired, it attempts to refresh using the refresh_token cookie.
// On failure it redirects to the login page rather than returning a JSON error response.
// It is intended for browser-facing web page routes.
func CookieAuthGuard(tokenMgr *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var acctInfo *AccountInfo

			// Try access token first
			if cookie, err := r.Cookie("access_token"); err == nil {
				if verified, err := tokenMgr.ValidateToken(cookie.Value); err == nil && verified.TokenKind == auth.AccessTokenKind {
					acctInfo = &AccountInfo{
						ID:   verified.AccountID,
						Name: verified.AccountName,
					}
				}
			}

			// If access token is missing or invalid, try to refresh using the refresh_token cookie
			if acctInfo == nil {
				refreshCookie, err := r.Cookie("refresh_token")
				if err != nil {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}

				refreshVerified, err := tokenMgr.ValidateToken(refreshCookie.Value)
				if err != nil || refreshVerified.TokenKind != auth.RefreshTokenKind {
					clearCookie(w, "refresh_token", r.TLS != nil)
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}

				bundle, err := tokenMgr.RefreshAccessToken(refreshCookie.Value)
				if err != nil {
					clearCookie(w, "refresh_token", r.TLS != nil)
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:     "access_token",
					Value:    bundle.AccessToken,
					Path:     "/",
					MaxAge:   int(time.Until(bundle.AccessExpiresAt).Seconds()),
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
					Secure:   r.TLS != nil,
				})
				http.SetCookie(w, &http.Cookie{
					Name:     "refresh_token",
					Value:    bundle.RefreshToken,
					Path:     "/",
					MaxAge:   int(time.Until(bundle.RefreshExpiresAt).Seconds()),
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
					Secure:   r.TLS != nil,
				})

				acctInfo = &AccountInfo{
					ID:   refreshVerified.AccountID,
					Name: refreshVerified.AccountName,
				}
			}

			enrichedCtx := context.WithValue(r.Context(), AccountContextKey, acctInfo)
			next.ServeHTTP(w, r.WithContext(enrichedCtx))
		})
	}
}

// writeAuthError writes a JSON error response for authentication failures
func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"success":false,"error":{"code":"AUTHENTICATION_ERROR","message":"` + message + `"}}`))
}

// clearCookie expires a named cookie immediately by setting MaxAge to -1.
func clearCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secure,
	})
}
