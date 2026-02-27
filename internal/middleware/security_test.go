package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/auth"
)

// TestAuthenticationBypass verifies that the authentication middleware cannot be
// bypassed via common attack vectors.
func TestAuthenticationBypass(t *testing.T) {
	tokenMgr := setupTokenManager(t)
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	guarded := AuthenticationGuard(tokenMgr)(innerHandler)

	t.Run("rejects token signed with different secret", func(t *testing.T) {
		otherMgr, err := auth.BuildTokenManager("completely-different-secret-key", auth.StandardTokenDurations())
		require.NoError(t, err)

		bundle, err := otherMgr.GenerateTokenBundle(auth.TokenPayload{
			AccountID:   "attacker-123",
			AccountName: "attacker",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+bundle.AccessToken)
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects token with none algorithm", func(t *testing.T) {
		// Craft a JWT with alg=none (a classic bypass attempt).
		claims := jwt.MapClaims{
			"account_id":   "attacker-123",
			"account_name": "attacker",
			"token_kind":   "access",
			"exp":          time.Now().Add(time.Hour).Unix(),
			"iat":          time.Now().Unix(),
		}
		unsignedToken := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenString, err := unsignedToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects expired access token", func(t *testing.T) {
		shortMgr, err := auth.BuildTokenManager("test-secret-key-minimum-length!", auth.TokenDurations{
			AccessValidity:  1 * time.Millisecond,
			RefreshValidity: 1 * time.Millisecond,
		})
		require.NoError(t, err)

		bundle, err := shortMgr.GenerateTokenBundle(auth.TokenPayload{
			AccountID:   "user-123",
			AccountName: "testuser",
		})
		require.NoError(t, err)

		// Wait for the token to expire.
		time.Sleep(10 * time.Millisecond)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+bundle.AccessToken)
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects completely fabricated token string", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.eyJmYWtlIjoidHJ1ZSJ9.invalidsig")
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects empty bearer token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer ")
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects refresh token used as access token", func(t *testing.T) {
		bundle, err := tokenMgr.GenerateTokenBundle(auth.TokenPayload{
			AccountID:   "user-123",
			AccountName: "testuser",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+bundle.RefreshToken)
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// TestCookieAuthBypass verifies that cookie-based authentication cannot be bypassed.
func TestCookieAuthBypass(t *testing.T) {
	tokenMgr := setupTokenManager(t)
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	guarded := CookieAuthGuard(tokenMgr)(innerHandler)

	t.Run("redirects when cookie has token signed with wrong secret", func(t *testing.T) {
		otherMgr, err := auth.BuildTokenManager("another-secret-key-entirely!!", auth.StandardTokenDurations())
		require.NoError(t, err)

		bundle, err := otherMgr.GenerateTokenBundle(auth.TokenPayload{
			AccountID:   "attacker-123",
			AccountName: "attacker",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: bundle.AccessToken})
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("Location"))
	})

	t.Run("redirects when cookie has fabricated token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: "not-a-real-jwt"})
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("Location"))
	})

	t.Run("redirects when cookie has refresh token instead of access token", func(t *testing.T) {
		bundle, err := tokenMgr.GenerateTokenBundle(auth.TokenPayload{
			AccountID:   "user-123",
			AccountName: "testuser",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: bundle.RefreshToken})
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("Location"))
	})

	t.Run("redirects when no cookie present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("Location"))
	})
}

// TestAuthorizationBypass verifies that the authorization middleware blocks
// users from accessing resources they do not own.
func TestAuthorizationBypass(t *testing.T) {
	t.Run("API guard blocks access to resource owned by another user", func(t *testing.T) {
		checker := func(accountID string, _ *http.Request) (bool, error) {
			// Only the real owner "owner-123" passes the check.
			return accountID == "owner-123", nil
		}

		handlerCalled := false
		inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		guarded := ResourceAuthorizationGuard(checker)(inner)

		req := httptest.NewRequest(http.MethodGet, "/resource/abc", nil)
		acct := &AccountInfo{ID: "attacker-456", Name: "attacker"}
		ctx := context.WithValue(req.Context(), AccountContextKey, acct)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.False(t, handlerCalled)
	})

	t.Run("page guard blocks access when owner IDs do not match", func(t *testing.T) {
		lookup := func(_ context.Context, _ *http.Request) (interface{}, string, error) {
			return "the-resource", "owner-123", nil
		}

		handlerCalled := false
		inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlerCalled = true
		})

		guarded := PageResourceOwnershipGuard(lookup)(inner)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1", nil)
		acct := &AccountInfo{ID: "attacker-456", Name: "attacker"}
		ctx := context.WithValue(req.Context(), AccountContextKey, acct)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		guarded.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.False(t, handlerCalled)
	})
}
