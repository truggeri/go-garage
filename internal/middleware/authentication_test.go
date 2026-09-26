package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		assert.Contains(t, rec.Body.String(), "access token required")
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

func TestCookieAuthGuard(t *testing.T) {
	tokenMgr := setupTokenManager(t)

	payload := auth.TokenPayload{
		AccountID:   "user-test-123",
		AccountName: "testuser",
	}
	bundle, err := tokenMgr.GenerateTokenBundle(payload)
	require.NoError(t, err)

	t.Run("allows request with valid access token cookie", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acct, ok := GetAccountFromContext(r.Context())
			if ok {
				capturedAcctInfo = acct
			}
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := CookieAuthGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: bundle.AccessToken})
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedAcctInfo)
		assert.Equal(t, "user-test-123", capturedAcctInfo.ID)
		assert.Equal(t, "testuser", capturedAcctInfo.Name)
	})

	t.Run("redirects to login when no cookie present", func(t *testing.T) {
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := CookieAuthGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("Location"))
	})

	t.Run("redirects to login when cookie has invalid token", func(t *testing.T) {
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := CookieAuthGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid.token.here"})
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("Location"))
	})

	t.Run("redirects to login when refresh token is used instead of access token", func(t *testing.T) {
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := CookieAuthGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: bundle.RefreshToken})
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("Location"))
	})

	t.Run("refreshes access token when expired and refresh token is valid", func(t *testing.T) {
		expiredAccessMgr, err := auth.BuildTokenManager("test-secret-key-12345", auth.TokenDurations{
			AccessValidity:  -1 * time.Hour,
			RefreshValidity: 7 * 24 * time.Hour,
		})
		require.NoError(t, err)

		expiredBundle, err := expiredAccessMgr.GenerateTokenBundle(payload)
		require.NoError(t, err)

		var capturedAcctInfo *AccountInfo
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acct, ok := GetAccountFromContext(r.Context())
			if ok {
				capturedAcctInfo = acct
			}
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := CookieAuthGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: expiredBundle.AccessToken})
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: expiredBundle.RefreshToken})
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedAcctInfo)
		assert.Equal(t, "user-test-123", capturedAcctInfo.ID)
		assert.Equal(t, "testuser", capturedAcctInfo.Name)

		var newAccessToken, newRefreshToken string
		for _, c := range rec.Result().Cookies() {
			switch c.Name {
			case "access_token":
				newAccessToken = c.Value
			case "refresh_token":
				newRefreshToken = c.Value
			}
		}
		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
		assert.NotEqual(t, expiredBundle.AccessToken, newAccessToken)
	})

	t.Run("redirects to login when access token is expired and refresh token is also expired", func(t *testing.T) {
		expiredMgr, err := auth.BuildTokenManager("test-secret-key-12345", auth.TokenDurations{
			AccessValidity:  -1 * time.Hour,
			RefreshValidity: -1 * time.Hour,
		})
		require.NoError(t, err)

		expiredBundle, err := expiredMgr.GenerateTokenBundle(payload)
		require.NoError(t, err)

		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := CookieAuthGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: expiredBundle.AccessToken})
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: expiredBundle.RefreshToken})
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("Location"))

		var clearedRefreshToken *http.Cookie
		for _, c := range rec.Result().Cookies() {
			if c.Name == "refresh_token" {
				clearedRefreshToken = c
			}
		}
		require.NotNil(t, clearedRefreshToken)
		assert.Equal(t, -1, clearedRefreshToken.MaxAge)
	})

	t.Run("refreshes access token when access token cookie is absent and refresh token is valid", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acct, ok := GetAccountFromContext(r.Context())
			if ok {
				capturedAcctInfo = acct
			}
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := CookieAuthGuard(tokenMgr)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: bundle.RefreshToken})
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedAcctInfo)
		assert.Equal(t, "user-test-123", capturedAcctInfo.ID)
	})
}

func TestHybridAuthGuard(t *testing.T) {
	tokenMgr := setupTokenManager(t)

	payload := auth.TokenPayload{
		AccountID:   "user-test-123",
		AccountName: "testuser",
	}
	bundle, err := tokenMgr.GenerateTokenBundle(payload)
	require.NoError(t, err)

	okHandler := func(acct **AccountInfo, method *AuthMethod) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if info, ok := GetAccountFromContext(r.Context()); ok {
				*acct = info
			}
			if used, ok := GetAuthMethodFromContext(r.Context()); ok {
				*method = used
			}
			w.WriteHeader(http.StatusOK)
		})
	}

	t.Run("allows request with valid bearer token", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		var capturedMethod AuthMethod

		guardedHandler := HybridAuthGuard(tokenMgr)(okHandler(&capturedAcctInfo, &capturedMethod))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		req.Header.Set("Authorization", "Bearer "+bundle.AccessToken)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedAcctInfo)
		assert.Equal(t, "user-test-123", capturedAcctInfo.ID)
		assert.Equal(t, AuthMethodBearer, capturedMethod)
	})

	t.Run("allows request with valid access token cookie", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		var capturedMethod AuthMethod

		guardedHandler := HybridAuthGuard(tokenMgr)(okHandler(&capturedAcctInfo, &capturedMethod))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: bundle.AccessToken})
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedAcctInfo)
		assert.Equal(t, "user-test-123", capturedAcctInfo.ID)
		assert.Equal(t, AuthMethodCookie, capturedMethod)
	})

	t.Run("refreshes expired access token cookie using refresh token", func(t *testing.T) {
		expiredAccessMgr, err := auth.BuildTokenManager("test-secret-key-12345", auth.TokenDurations{
			AccessValidity:  -1 * time.Hour,
			RefreshValidity: 7 * 24 * time.Hour,
		})
		require.NoError(t, err)

		expiredBundle, err := expiredAccessMgr.GenerateTokenBundle(payload)
		require.NoError(t, err)

		var capturedAcctInfo *AccountInfo
		var capturedMethod AuthMethod

		guardedHandler := HybridAuthGuard(tokenMgr)(okHandler(&capturedAcctInfo, &capturedMethod))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: expiredBundle.AccessToken})
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: expiredBundle.RefreshToken})
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedAcctInfo)
		assert.Equal(t, "user-test-123", capturedAcctInfo.ID)
		assert.Equal(t, AuthMethodCookie, capturedMethod)

		var refreshedAccess bool
		for _, cookie := range rec.Result().Cookies() {
			if cookie.Name == "access_token" && cookie.Value != "" {
				refreshedAccess = true
			}
		}
		assert.True(t, refreshedAccess, "expected a refreshed access_token cookie")
	})

	t.Run("rejects request without any credentials", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		var capturedMethod AuthMethod

		guardedHandler := HybridAuthGuard(tokenMgr)(okHandler(&capturedAcctInfo, &capturedMethod))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Empty(t, rec.Header().Get("HX-Redirect"))
		assert.Contains(t, rec.Body.String(), "AUTHENTICATION_ERROR")
	})

	t.Run("rejects invalid bearer token without falling back to cookies", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		var capturedMethod AuthMethod

		guardedHandler := HybridAuthGuard(tokenMgr)(okHandler(&capturedAcctInfo, &capturedMethod))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		req.AddCookie(&http.Cookie{Name: "access_token", Value: bundle.AccessToken})
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid or expired token")
		assert.Nil(t, capturedAcctInfo)
	})

	t.Run("rejects refresh token supplied as bearer token", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		var capturedMethod AuthMethod

		guardedHandler := HybridAuthGuard(tokenMgr)(okHandler(&capturedAcctInfo, &capturedMethod))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		req.Header.Set("Authorization", "Bearer "+bundle.RefreshToken)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "access token required")
	})

	t.Run("rejects invalid cookies and clears the refresh cookie", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		var capturedMethod AuthMethod

		guardedHandler := HybridAuthGuard(tokenMgr)(okHandler(&capturedAcctInfo, &capturedMethod))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid.token.here"})
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid.token.here"})
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid or expired token")

		var cleared bool
		for _, cookie := range rec.Result().Cookies() {
			if cookie.Name == "refresh_token" && cookie.MaxAge < 0 {
				cleared = true
			}
		}
		assert.True(t, cleared, "expected the refresh_token cookie to be cleared")
	})

	t.Run("adds HX-Redirect header for htmx requests", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		var capturedMethod AuthMethod

		guardedHandler := HybridAuthGuard(tokenMgr)(okHandler(&capturedAcctInfo, &capturedMethod))

		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("HX-Redirect"))
		assert.Contains(t, rec.Body.String(), "AUTHENTICATION_ERROR")
	})

	t.Run("adds HX-Redirect header for html-preferring browser requests", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		var capturedMethod AuthMethod

		guardedHandler := HybridAuthGuard(tokenMgr)(okHandler(&capturedAcctInfo, &capturedMethod))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("HX-Redirect"))
	})

	t.Run("omits HX-Redirect header when html is excluded by quality", func(t *testing.T) {
		var capturedAcctInfo *AccountInfo
		var capturedMethod AuthMethod

		guardedHandler := HybridAuthGuard(tokenMgr)(okHandler(&capturedAcctInfo, &capturedMethod))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		req.Header.Set("Accept", "application/json, text/html;q=0")
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Empty(t, rec.Header().Get("HX-Redirect"))
	})
}
