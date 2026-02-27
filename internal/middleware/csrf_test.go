package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSRFProtection_SetsTokenCookie(t *testing.T) {
	handler := CSRFProtection()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/form", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == csrfCookieName {
			csrfCookie = c
			break
		}
	}
	require.NotNil(t, csrfCookie, "csrf_token cookie should be set")
	assert.NotEmpty(t, csrfCookie.Value)
	assert.True(t, csrfCookie.HttpOnly)
	assert.Equal(t, csrfCookieMaxAge, csrfCookie.MaxAge)
	assert.Equal(t, http.SameSiteStrictMode, csrfCookie.SameSite)
}

func TestCSRFProtection_TokenAvailableInContext(t *testing.T) {
	var ctxToken string
	handler := CSRFProtection()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxToken = GetCSRFToken(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/form", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.NotEmpty(t, ctxToken)
}

func TestCSRFProtection_ReusesCookieToken(t *testing.T) {
	var firstToken, secondToken string

	handler := CSRFProtection()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondToken = GetCSRFToken(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// First request sets the cookie.
	req1 := httptest.NewRequest(http.MethodGet, "/form", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	for _, c := range rec1.Result().Cookies() {
		if c.Name == csrfCookieName {
			firstToken = c.Value
		}
	}
	require.NotEmpty(t, firstToken)

	// Second request sends the cookie back.
	req2 := httptest.NewRequest(http.MethodGet, "/form", nil)
	req2.AddCookie(&http.Cookie{Name: csrfCookieName, Value: firstToken})
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	assert.Equal(t, firstToken, secondToken, "token from cookie should be reused")
}

func TestCSRFProtection_POST_ValidToken(t *testing.T) {
	token := "test-csrf-token-value"

	handler := CSRFProtection()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	form := url.Values{}
	form.Set(csrfFormField, token)

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: token})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRFProtection_POST_MissingFormToken(t *testing.T) {
	handler := CSRFProtection()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "cookie-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRFProtection_POST_MismatchedToken(t *testing.T) {
	handler := CSRFProtection()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	form := url.Values{}
	form.Set(csrfFormField, "wrong-token")

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "correct-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRFProtection_GET_NoTokenRequired(t *testing.T) {
	handler := CSRFProtection()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
