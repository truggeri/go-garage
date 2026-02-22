package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestPageHandler(t *testing.T, authSvc services.AuthenticationService) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, authSvc)
}

func TestPageHandler_Home(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.Home(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Welcome to Go-Garage")
}

func TestPageHandler_RegisterForm(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()

	handler.RegisterForm(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "Create an Account")
	assert.Contains(t, body, `method="POST"`)
	assert.Contains(t, body, `action="/register"`)
}

func TestPageHandler_RegisterSubmit(t *testing.T) {
	t.Run("redirects on successful registration", func(t *testing.T) {
		mockSvc := &mockAuthService{
			registerResult: &services.AuthenticationResult{
				AccessToken:  "token",
				RefreshToken: "refresh",
				AccountID:    "user-1",
				AccountName:  "testuser",
			},
		}
		handler := newTestPageHandler(t, mockSvc)

		form := url.Values{}
		form.Set("username", "testuser")
		form.Set("email", "test@example.com")
		form.Set("password", "StrongPass1")
		form.Set("confirm_password", "StrongPass1")
		form.Set("first_name", "Test")
		form.Set("last_name", "User")

		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.RegisterSubmit(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login?registered=true", rec.Header().Get("Location"))
	})

	t.Run("shows errors for missing fields", func(t *testing.T) {
		handler := newTestPageHandler(t, &mockAuthService{})

		form := url.Values{}
		// All fields empty

		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.RegisterSubmit(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("shows error for password mismatch", func(t *testing.T) {
		handler := newTestPageHandler(t, &mockAuthService{})

		form := url.Values{}
		form.Set("username", "testuser")
		form.Set("email", "test@example.com")
		form.Set("password", "StrongPass1")
		form.Set("confirm_password", "DifferentPass1")

		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.RegisterSubmit(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("shows error for duplicate user", func(t *testing.T) {
		mockSvc := &mockAuthService{
			registerErr: models.NewDuplicateError("User", "username", "testuser"),
		}
		handler := newTestPageHandler(t, mockSvc)

		form := url.Values{}
		form.Set("username", "testuser")
		form.Set("email", "test@example.com")
		form.Set("password", "StrongPass1")
		form.Set("confirm_password", "StrongPass1")

		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.RegisterSubmit(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("shows error for validation failure", func(t *testing.T) {
		mockSvc := &mockAuthService{
			registerErr: models.NewValidationError("password", "password too weak"),
		}
		handler := newTestPageHandler(t, mockSvc)

		form := url.Values{}
		form.Set("username", "testuser")
		form.Set("email", "test@example.com")
		form.Set("password", "weak")
		form.Set("confirm_password", "weak")

		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.RegisterSubmit(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestPageHandler_LoginForm(t *testing.T) {
	t.Run("renders login form", func(t *testing.T) {
		handler := newTestPageHandler(t, &mockAuthService{})

		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		rec := httptest.NewRecorder()

		handler.LoginForm(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Welcome Back")
		assert.Contains(t, body, `method="POST"`)
		assert.Contains(t, body, `action="/login"`)
	})

	t.Run("shows success flash after registration", func(t *testing.T) {
		handler := newTestPageHandler(t, &mockAuthService{})

		req := httptest.NewRequest(http.MethodGet, "/login?registered=true", nil)
		rec := httptest.NewRecorder()

		handler.LoginForm(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Account created successfully")
	})
}

func TestPageHandler_LoginSubmit(t *testing.T) {
	t.Run("redirects on successful login", func(t *testing.T) {
		mockSvc := &mockAuthService{
			authenticateResult: &services.AuthenticationResult{
				AccessToken:      "access-token",
				RefreshToken:     "refresh-token",
				AccessExpiresAt:  time.Now().Add(15 * time.Minute).Unix(),
				RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
				AccountID:        "user-1",
				AccountName:      "testuser",
			},
		}
		handler := newTestPageHandler(t, mockSvc)

		form := url.Values{}
		form.Set("identifier", "testuser")
		form.Set("password", "StrongPass1")

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.LoginSubmit(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/", rec.Header().Get("Location"))

		cookies := rec.Result().Cookies()
		var accessCookie, refreshCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "access_token" {
				accessCookie = c
			}
			if c.Name == "refresh_token" {
				refreshCookie = c
			}
		}
		assert.NotNil(t, accessCookie)
		assert.Equal(t, "access-token", accessCookie.Value)
		assert.True(t, accessCookie.HttpOnly)
		assert.Greater(t, accessCookie.MaxAge, 0)
		assert.NotNil(t, refreshCookie)
		assert.Equal(t, "refresh-token", refreshCookie.Value)
		assert.True(t, refreshCookie.HttpOnly)
		assert.Greater(t, refreshCookie.MaxAge, 0)
	})

	t.Run("shows errors for missing fields", func(t *testing.T) {
		handler := newTestPageHandler(t, &mockAuthService{})

		form := url.Values{}
		// All fields empty

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.LoginSubmit(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("shows error for invalid credentials", func(t *testing.T) {
		mockSvc := &mockAuthService{
			authenticateErr: models.NewValidationError("credentials", "invalid email/username or password"),
		}
		handler := newTestPageHandler(t, mockSvc)

		form := url.Values{}
		form.Set("identifier", "wronguser")
		form.Set("password", "WrongPass1")

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.LoginSubmit(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid email/username or password")
	})

	t.Run("shows error for unexpected service error", func(t *testing.T) {
		mockSvc := &mockAuthService{
			authenticateErr: models.NewDatabaseError("authenticate", assert.AnError),
		}
		handler := newTestPageHandler(t, mockSvc)

		form := url.Values{}
		form.Set("identifier", "testuser")
		form.Set("password", "StrongPass1")

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.LoginSubmit(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "An unexpected error occurred")
	})
}
