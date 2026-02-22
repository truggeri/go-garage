package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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
