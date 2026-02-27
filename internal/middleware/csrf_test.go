package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCSRFSecret = "test-csrf-secret-key"

func TestCSRFProtection_SetsTokenInContext(t *testing.T) {
	var ctxToken string
	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxToken = GetCSRFToken(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/form", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, ctxToken)
	assert.Contains(t, ctxToken, csrfTokenSeparator, "token should contain nonce.signature")
}

func TestCSRFProtection_NoCookieSet(t *testing.T) {
	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/form", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		assert.NotEqual(t, "csrf_token", c.Name, "HMAC CSRF should not set a csrf_token cookie")
	}
}

func TestCSRFProtection_GeneratesUniqueTokens(t *testing.T) {
	var token1, token2 string

	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token2 = GetCSRFToken(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/form", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	token1 = token2

	req2 := httptest.NewRequest(http.MethodGet, "/form", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	assert.NotEqual(t, token1, token2, "each request should produce a unique token")
}

func TestCSRFProtection_POST_ValidToken_NoSession(t *testing.T) {
	// Generate a valid token for an unauthenticated session.
	secret := []byte(testCSRFSecret)
	token := generateCSRFToken("", secret)

	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	form := url.Values{}
	form.Set(csrfFormField, token)

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRFProtection_POST_ValidToken_WithSession(t *testing.T) {
	userID := "user-123"
	secret := []byte(testCSRFSecret)
	token := generateCSRFToken(userID, secret)

	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	form := url.Values{}
	form.Set(csrfFormField, token)

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), AccountContextKey, &AccountInfo{ID: userID, Name: "Test User"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRFProtection_POST_MissingFormToken(t *testing.T) {
	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRFProtection_POST_InvalidToken(t *testing.T) {
	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	form := url.Values{}
	form.Set(csrfFormField, "invalid-token-value")

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRFProtection_POST_WrongSessionID(t *testing.T) {
	// Token generated for one session, validated with another.
	secret := []byte(testCSRFSecret)
	token := generateCSRFToken("user-123", secret)

	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	form := url.Values{}
	form.Set(csrfFormField, token)

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), AccountContextKey, &AccountInfo{ID: "user-456", Name: "Other User"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRFProtection_POST_WrongSecret(t *testing.T) {
	// Token generated with a different secret.
	token := generateCSRFToken("", []byte("different-secret"))

	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	form := url.Values{}
	form.Set(csrfFormField, token)

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRFProtection_GET_NoTokenRequired(t *testing.T) {
	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/page", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCSRFToken_EmptyContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Empty(t, GetCSRFToken(req.Context()))
}

func TestValidateCSRFToken_MalformedToken(t *testing.T) {
	secret := []byte(testCSRFSecret)

	tests := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"no separator", "abcdef1234567890"},
		{"empty nonce", ".signature"},
		{"empty signature", "nonce."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, validateCSRFToken(tt.token, "", secret))
		})
	}
}
