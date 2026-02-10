package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/auth"
)

func setupTokenManager(t *testing.T) *auth.TokenManager {
	mgr, err := auth.BuildTokenManager("test-secret-key-12345", auth.StandardTokenDurations())
	require.NoError(t, err)
	return mgr
}

func TestAuthenticationGuard(t *testing.T) {
	tokenMgr := setupTokenManager(t)

	payload := auth.TokenPayload{
		AccountID:   "user-test-123",
		AccountName: "testuser",
	}
	bundle, err := tokenMgr.GenerateTokenBundle(payload)
	require.NoError(t, err)

	t.Run("allows request with valid access token", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acct, ok := GetAccountFromContext(r.Context())
			if ok {
				capturedAcctInfo = acct
			}
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := AuthenticationGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+bundle.AccessToken)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedAcctInfo)
		assert.Equal(t, "user-test-123", capturedAcctInfo.ID)
		assert.Equal(t, "testuser", capturedAcctInfo.Name)
	})

	t.Run("rejects request without authorization header", func(t *testing.T) {
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := AuthenticationGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "missing authorization header")
	})

	t.Run("rejects request with invalid authorization format", func(t *testing.T) {
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := AuthenticationGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Basic somecredentials")
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid authorization format")
	})

	t.Run("rejects request with invalid token", func(t *testing.T) {
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := AuthenticationGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid or expired token")
	})

	t.Run("rejects request with refresh token instead of access token", func(t *testing.T) {
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := AuthenticationGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+bundle.RefreshToken)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid token type")
	})

	t.Run("accepts lowercase bearer prefix", func(t *testing.T) {
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := AuthenticationGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "bearer "+bundle.AccessToken)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestGetAccountFromContext(t *testing.T) {
	t.Run("returns nil when no account in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		acct, ok := GetAccountFromContext(req.Context())
		assert.False(t, ok)
		assert.Nil(t, acct)
	})
}
