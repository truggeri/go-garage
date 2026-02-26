package middleware

import (
	"context"
	"errors"
	"net/http"
)

// ErrResourceNotFound should be returned by a ResourceLookup when the requested resource does not exist.
var ErrResourceNotFound = errors.New("resource not found")

// LoadedResourceContextKey is the context key for the resource loaded by an ownership guard.
const LoadedResourceContextKey contextKey = "loadedResource"

// GetLoadedResourceFromContext retrieves the resource stored by a resource ownership guard.
func GetLoadedResourceFromContext(ctx context.Context) (interface{}, bool) {
	v := ctx.Value(LoadedResourceContextKey)
	return v, v != nil
}

// ResourceOwnershipChecker is a function that checks if a user owns a resource
// It returns true if the user has permission to access the resource
type ResourceOwnershipChecker func(accountID string, r *http.Request) (bool, error)

// ResourceLookup loads a resource from the request and returns it along with its owner's user ID.
// Return ErrResourceNotFound (or wrap it) when the resource does not exist.
type ResourceLookup func(ctx context.Context, r *http.Request) (resource interface{}, ownerID string, err error)

// ResourceAuthorizationGuard creates middleware that verifies resource ownership
// The checker function is called with the authenticated user's ID and the request
// to determine if they have permission to access the requested resource
func ResourceAuthorizationGuard(checker ResourceOwnershipChecker) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acctInfo, ok := GetAccountFromContext(r.Context())
			if !ok || acctInfo == nil {
				writeAuthzError(w, "authentication required", http.StatusUnauthorized)
				return
			}

			hasPermission, err := checker(acctInfo.ID, r)
			if err != nil {
				writeAuthzError(w, "authorization check failed", http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				writeAuthzError(w, "you do not have permission to access this resource", http.StatusForbidden)
				return
			}

			nextHandler.ServeHTTP(w, r)
		})
	}
}

// PageResourceOwnershipGuard creates middleware for page routes that loads a
// resource, verifies the authenticated user owns it, and stores it in the
// request context under LoadedResourceContextKey.
// On failure it returns plain-text HTTP errors appropriate for browser pages.
func PageResourceOwnershipGuard(lookup ResourceLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acctInfo, ok := GetAccountFromContext(r.Context())
			if !ok || acctInfo == nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			resource, ownerID, err := lookup(r.Context(), r)
			if err != nil {
				if errors.Is(err, ErrResourceNotFound) {
					http.Error(w, "Not Found", http.StatusNotFound)
				} else {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}

			if ownerID != acctInfo.ID {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), LoadedResourceContextKey, resource)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeAuthzError writes a JSON error response for authorization failures
func writeAuthzError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"success":false,"error":{"code":"AUTHORIZATION_ERROR","message":"` + message + `"}}`))
}
