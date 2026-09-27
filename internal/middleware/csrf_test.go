package middleware

import (
	"context"
	"io"
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

func TestCSRFProtection_POST_ValidHeaderToken(t *testing.T) {
	userID := "user-123"
	token := generateCSRFToken(userID, []byte(testCSRFSecret))

	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(`{"name":"value"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(csrfHeaderName, token)
	req = req.WithContext(context.WithValue(req.Context(), AccountContextKey, &AccountInfo{ID: userID, Name: "Test User"}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRFProtection_POST_InvalidHeaderToken(t *testing.T) {
	handler := CSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(csrfHeaderName, "invalid-token-value")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// newAPICSRFRequest builds a JSON API request authenticated with the given method.
func newAPICSRFRequest(t *testing.T, httpMethod string, authMethod AuthMethod, token string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(httpMethod, "/api/v1/vehicles", strings.NewReader(`{"make":"Ford"}`))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set(csrfHeaderName, token)
	}

	ctx := context.WithValue(req.Context(), AccountContextKey, &AccountInfo{ID: "user-123", Name: "Test User"})
	ctx = context.WithValue(ctx, AuthMethodContextKey, authMethod)
	return req.WithContext(ctx)
}

func TestAPICSRFProtection_CookieAuth_ValidHeaderToken(t *testing.T) {
	token := generateCSRFToken("user-123", []byte(testCSRFSecret))

	handler := APICSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, newAPICSRFRequest(t, http.MethodPost, AuthMethodCookie, token))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAPICSRFProtection_CookieAuth_MissingToken(t *testing.T) {
	handler := APICSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, httpMethod := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(httpMethod, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, newAPICSRFRequest(t, httpMethod, AuthMethodCookie, ""))

			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			assert.Contains(t, rec.Body.String(), "CSRF_ERROR")
		})
	}
}

func TestAPICSRFProtection_CookieAuth_WrongSessionToken(t *testing.T) {
	token := generateCSRFToken("other-user", []byte(testCSRFSecret))

	handler := APICSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, newAPICSRFRequest(t, http.MethodPost, AuthMethodCookie, token))

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAPICSRFProtection_BearerAuth_NoTokenRequired(t *testing.T) {
	handler := APICSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, httpMethod := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(httpMethod, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, newAPICSRFRequest(t, httpMethod, AuthMethodBearer, ""))

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestAPICSRFProtection_CookieAuth_GETNoTokenRequired(t *testing.T) {
	var ctxToken string
	handler := APICSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxToken = GetCSRFToken(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
	req = req.WithContext(context.WithValue(req.Context(), AuthMethodContextKey, AuthMethodCookie))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, ctxToken)
}

func TestAPICSRFProtection_JSONBodyPreserved(t *testing.T) {
	token := generateCSRFToken("user-123", []byte(testCSRFSecret))

	var body string
	handler := APICSRFProtection(testCSRFSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		body = string(raw)
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, newAPICSRFRequest(t, http.MethodPost, AuthMethodCookie, token))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"make":"Ford"}`, body)
}
