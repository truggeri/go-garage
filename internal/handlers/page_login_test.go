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
)

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
		assert.Equal(t, "/dashboard", rec.Header().Get("Location"))

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
