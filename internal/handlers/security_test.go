package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/services"
)

// TestXSSPrevention_LoginPage verifies that user-supplied values are HTML-escaped
// when rendered in the login page, preventing reflected XSS attacks.
func TestXSSPrevention_LoginPage(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	xssPayloads := []string{
		`<script>alert('xss')</script>`,
		`"><img src=x onerror=alert(1)>`,
		`<svg onload=alert(1)>`,
		`' onmouseover='alert(1)'`,
	}

	for _, payload := range xssPayloads {
		t.Run("escapes XSS in identifier field: "+payload[:min(len(payload), 30)], func(t *testing.T) {
			form := url.Values{}
			form.Set("identifier", payload)
			// Leave password empty to trigger validation error and re-render with user input.
			form.Set("password", "")

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			handler.LoginSubmit(rec, req)

			body := rec.Body.String()
			// The raw payload should NOT appear unescaped in the response body.
			assert.NotContains(t, body, payload, "raw XSS payload must be escaped in response")
		})
	}
}

// TestXSSPrevention_RegisterPage verifies that user-supplied values are
// HTML-escaped when rendered in the registration page.
func TestXSSPrevention_RegisterPage(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	xssPayloads := []string{
		`<script>alert('xss')</script>`,
		`"><img src=x onerror=alert(1)>`,
		`<svg onload=alert(1)>`,
		`' onmouseover='alert(1)'`,
	}

	for _, payload := range xssPayloads {
		t.Run("escapes XSS in username field: "+payload[:min(len(payload), 30)], func(t *testing.T) {
			form := url.Values{}
			form.Set("username", payload)
			form.Set("email", "")
			form.Set("password", "Test1234")
			form.Set("confirm_password", "Test1234")

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			handler.RegisterSubmit(rec, req)

			body := rec.Body.String()
			assert.NotContains(t, body, payload, "raw XSS payload must be escaped in response")
		})

		t.Run("escapes XSS in email field: "+payload[:min(len(payload), 30)], func(t *testing.T) {
			form := url.Values{}
			form.Set("username", "testuser")
			form.Set("email", payload)
			form.Set("password", "")
			form.Set("confirm_password", "")

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			handler.RegisterSubmit(rec, req)

			body := rec.Body.String()
			assert.NotContains(t, body, payload, "raw XSS payload must be escaped in response")
		})
	}
}

// TestCSRFProtection_CookieAttributes verifies that authentication cookies
// are set with security attributes that help mitigate CSRF attacks.
func TestCSRFProtection_CookieAttributes(t *testing.T) {
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

	require.Equal(t, http.StatusSeeOther, rec.Code)

	cookies := rec.Result().Cookies()
	cookieNames := []string{"access_token", "refresh_token"}

	for _, name := range cookieNames {
		var found *http.Cookie
		for _, c := range cookies {
			if c.Name == name {
				found = c
				break
			}
		}
		require.NotNil(t, found, "cookie %q must be set", name)

		t.Run(name+" is HttpOnly", func(t *testing.T) {
			assert.True(t, found.HttpOnly, "cookie %q must be HttpOnly to prevent JS access", name)
		})

		t.Run(name+" has SameSite=Strict", func(t *testing.T) {
			assert.Equal(t, http.SameSiteStrictMode, found.SameSite,
				"cookie %q must use SameSite=Strict to mitigate CSRF", name)
		})

		t.Run(name+" has a reasonable MaxAge", func(t *testing.T) {
			assert.Greater(t, found.MaxAge, 0,
				"cookie %q must have a positive MaxAge so it expires", name)
		})
	}
}

// TestAuthenticationBypass_APIHandlers verifies that API handlers reject
// requests without valid authentication context.
func TestAuthenticationBypass_APIHandlers(t *testing.T) {
	stub := &stubVehicleSvc{}
	vehicleHandler := MakeVehicleAPIHandler(stub)

	t.Run("ListAll returns 401 without auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		rec := httptest.NewRecorder()

		vehicleHandler.ListAll(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("CreateOne returns 401 without auth", func(t *testing.T) {
		body := `{"vin":"1HGBH41JXMN109186","make":"Honda","model":"Civic","year":2021}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", strings.NewReader(body))
		rec := httptest.NewRecorder()

		vehicleHandler.CreateOne(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
