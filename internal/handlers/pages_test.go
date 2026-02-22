package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestPageHandler(t *testing.T, authSvc services.AuthenticationService) *PageHandler {
	t.Helper()
	dir := createTestPageTemplates(t)
	engine := templateengine.NewEngine(dir, true)
	return NewPageHandler(engine, authSvc)
}

// createTestPageTemplates sets up a temporary template directory with minimal templates.
func createTestPageTemplates(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	for _, d := range []string{"layouts", "partials", "pages", "errors"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, d), 0o755))
	}

	files := map[string]string{
		"layouts/base.html":            `{{define "base"}}<!DOCTYPE html><html><head><title>{{block "title" .}}Go-Garage{{end}}</title></head><body>{{template "flash-messages" .}}{{block "content" .}}{{end}}</body></html>{{end}}`,
		"layouts/auth.html":            `{{define "auth"}}<!DOCTYPE html><html><head><title>{{block "title" .}}Go-Garage{{end}}</title></head><body class="auth-page">{{template "flash-messages" .}}{{block "content" .}}{{end}}</body></html>{{end}}`,
		"partials/flash-messages.html": `{{define "flash-messages"}}{{end}}`,
		"partials/navigation.html":     `{{define "navigation"}}{{end}}`,
		"partials/header.html":         `{{define "header"}}{{end}}`,
		"partials/footer.html":         `{{define "footer"}}{{end}}`,
		"pages/home.html":              `{{define "title"}}Home{{end}}{{define "content"}}<h1>Welcome</h1>{{end}}`,
		"pages/register.html":          `{{define "title"}}Register{{end}}{{define "content"}}<form method="POST" action="/register">{{if .Errors.username}}<p class="form-error">{{.Errors.username}}</p>{{end}}<input name="username" value="{{.Username}}"><input name="email" value="{{.Email}}"><button type="submit">Register</button></form>{{end}}`,
		"errors/404.html":              `{{define "title"}}Not Found{{end}}{{define "content"}}<h1>404</h1>{{end}}`,
	}

	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	}

	return dir
}

func TestPageHandler_Home(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.Home(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Welcome")
}

func TestPageHandler_RegisterForm(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()

	handler.RegisterForm(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "Register")
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
