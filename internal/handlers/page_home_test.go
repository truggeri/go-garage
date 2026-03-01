package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageHandler_Home(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	t.Run("renders home page for unauthenticated users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.Home(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Welcome to Go-Garage")
	})

	t.Run("redirects to dashboard when access_token cookie is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: "some-token"})
		rec := httptest.NewRecorder()

		handler.Home(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/dashboard", rec.Header().Get("Location"))
	})
}
