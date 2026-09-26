package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/truggeri/go-garage/internal/auth"
)

// Authentication failure reasons surfaced to clients
var (
	errMissingAuthHeader  = errors.New("missing authorization header")
	errInvalidAuthFormat  = errors.New("invalid authorization format")
	errInvalidToken       = errors.New("invalid or expired token")
	errRefreshTokenUsed   = errors.New("access token required; refresh tokens cannot be used for API requests")
	errMissingCredentials = errors.New("missing authentication credentials")
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// AccountContextKey is the key used to store account info in request context
const AccountContextKey contextKey = "accountInfo"

// AuthMethodContextKey is the key used to store the authentication mechanism
// that succeeded for the request
const AuthMethodContextKey contextKey = "authMethod"

// AuthMethod identifies which credential type authenticated a request
type AuthMethod string

const (
	// AuthMethodBearer indicates the request carried an Authorization bearer token
	AuthMethodBearer AuthMethod = "bearer"
	// AuthMethodCookie indicates the request was authenticated with session cookies
	AuthMethodCookie AuthMethod = "cookie"
)

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

// GetAuthMethodFromContext retrieves the authentication mechanism used for the request
func GetAuthMethodFromContext(ctx context.Context) (AuthMethod, bool) {
	method, ok := ctx.Value(AuthMethodContextKey).(AuthMethod)
	return method, ok
}

// AuthenticationGuard creates middleware that validates JWT tokens
// It extracts the Bearer token from the Authorization header and validates it
func AuthenticationGuard(tokenMgr *auth.TokenManager) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acctInfo, err := authenticateBearer(tokenMgr, r)
			if err != nil {
				writeAuthError(w, err.Error())
				return
			}

			nextHandler.ServeHTTP(w, r.WithContext(authenticatedContext(r, acctInfo, AuthMethodBearer)))
		})
	}
}

// HybridAuthGuard creates middleware that authenticates API requests using either an
// Authorization bearer token or the access_token/refresh_token session cookies.
// Header credentials are preferred so programmatic API clients behave exactly as before;
// requests without an Authorization header fall back to the cookie flow (including refresh)
// used by browser page routes. Failures return the standard JSON error envelope, with an
// additional HX-Redirect header so htmx/browser requests navigate to the login page.
func HybridAuthGuard(tokenMgr *auth.TokenManager) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "" {
				acctInfo, err := authenticateBearer(tokenMgr, r)
				if err != nil {
					writeHybridAuthError(w, r, err.Error())
					return
				}

				nextHandler.ServeHTTP(w, r.WithContext(authenticatedContext(r, acctInfo, AuthMethodBearer)))
				return
			}

			acctInfo, err := authenticateCookies(tokenMgr, w, r)
			if err != nil {
				writeHybridAuthError(w, r, err.Error())
				return
			}

			nextHandler.ServeHTTP(w, r.WithContext(authenticatedContext(r, acctInfo, AuthMethodCookie)))
		})
	}
}

// authenticateBearer validates the Authorization header and returns the account it identifies
func authenticateBearer(tokenMgr *auth.TokenManager, r *http.Request) (*AccountInfo, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errMissingAuthHeader
	}

	headerParts := strings.SplitN(authHeader, " ", 2)
	if len(headerParts) != 2 || !strings.EqualFold(headerParts[0], "Bearer") {
		return nil, errInvalidAuthFormat
	}

	verified, err := tokenMgr.ValidateToken(headerParts[1])
	if err != nil {
		return nil, errInvalidToken
	}

	if verified.TokenKind != auth.AccessTokenKind {
		return nil, errRefreshTokenUsed
	}

	return &AccountInfo{ID: verified.AccountID, Name: verified.AccountName}, nil
}

// authenticateCookies validates the access_token cookie, refreshing it with the
// refresh_token cookie when needed. Refreshed cookies are written to the response.
func authenticateCookies(tokenMgr *auth.TokenManager, w http.ResponseWriter, r *http.Request) (*AccountInfo, error) {
	if cookie, err := r.Cookie("access_token"); err == nil {
		if verified, err := tokenMgr.ValidateToken(cookie.Value); err == nil && verified.TokenKind == auth.AccessTokenKind {
			return &AccountInfo{ID: verified.AccountID, Name: verified.AccountName}, nil
		}
	}

	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		return nil, errMissingCredentials
	}

	refreshVerified, err := tokenMgr.ValidateToken(refreshCookie.Value)
	if err != nil || refreshVerified.TokenKind != auth.RefreshTokenKind {
		clearCookie(w, "refresh_token", r.TLS != nil)
		return nil, errInvalidToken
	}

	bundle, err := tokenMgr.RefreshAccessToken(refreshCookie.Value)
	if err != nil {
		clearCookie(w, "refresh_token", r.TLS != nil)
		return nil, errInvalidToken
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

	return &AccountInfo{ID: refreshVerified.AccountID, Name: refreshVerified.AccountName}, nil
}

// authenticatedContext stores the account and the mechanism that authenticated it
func authenticatedContext(r *http.Request, acctInfo *AccountInfo, method AuthMethod) context.Context {
	ctx := context.WithValue(r.Context(), AccountContextKey, acctInfo)
	return context.WithValue(ctx, AuthMethodContextKey, method)
}

// CookieAuthGuard creates middleware that validates JWT tokens from the access_token cookie.
// When the access token is missing or expired, it attempts to refresh using the refresh_token cookie.
// On failure it redirects to the login page rather than returning a JSON error response.
// It is intended for browser-facing web page routes.
func CookieAuthGuard(tokenMgr *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acctInfo, err := authenticateCookies(tokenMgr, w, r)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r.WithContext(authenticatedContext(r, acctInfo, AuthMethodCookie)))
		})
	}
}

// writeHybridAuthError writes the JSON authentication error, adding an HX-Redirect
// header so htmx and other browser-originated requests navigate to the login page.
func writeHybridAuthError(w http.ResponseWriter, r *http.Request, message string) {
	if isBrowserRequest(r) {
		w.Header().Set("HX-Redirect", "/login")
	}
	writeAuthError(w, message)
}

// isBrowserRequest reports whether the request originated from htmx or a browser
// navigation rather than a programmatic API client
func isBrowserRequest(r *http.Request) bool {
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		return true
	}

	return strings.Contains(r.Header.Get("Accept"), "text/html")
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
