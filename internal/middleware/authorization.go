package middleware

import (
	"net/http"
)

// ResourceOwnershipChecker is a function that checks if a user owns a resource
// It returns true if the user has permission to access the resource
type ResourceOwnershipChecker func(accountID string, r *http.Request) (bool, error)

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

// writeAuthzError writes a JSON error response for authorization failures
func writeAuthzError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"success":false,"error":{"code":"AUTHORIZATION_ERROR","message":"` + message + `"}}`))
}
