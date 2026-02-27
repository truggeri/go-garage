package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageHandler_NotFound(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	t.Run("renders 404 page for unauthenticated user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		rec := httptest.NewRecorder()

		handler.NotFound(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "404")
		assert.Contains(t, body, "Page Not Found")
		assert.Contains(t, body, "Back to Home")
	})

	t.Run("renders 404 page with dashboard link for authenticated user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.NotFound(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "404")
		assert.Contains(t, body, "Back to Dashboard")
	})
}

func TestPageHandler_Forbidden(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	t.Run("renders 403 page for unauthenticated user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/forbidden", nil)
		rec := httptest.NewRecorder()

		handler.Forbidden(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "403")
		assert.Contains(t, body, "Forbidden")
		assert.Contains(t, body, "Back to Home")
	})

	t.Run("renders 403 page with dashboard link for authenticated user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/forbidden", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.Forbidden(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "403")
		assert.Contains(t, body, "Back to Dashboard")
	})
}

func TestPageHandler_ServerError(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	t.Run("renders 500 page for unauthenticated user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		rec := httptest.NewRecorder()

		handler.ServerError(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "500")
		assert.Contains(t, body, "Internal Server Error")
		assert.Contains(t, body, "Back to Home")
	})

	t.Run("renders 500 page with dashboard link for authenticated user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.ServerError(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "500")
		assert.Contains(t, body, "Back to Dashboard")
	})
}

func TestPageHandler_RenderError(t *testing.T) {
	handler := newTestPageHandler(t, &mockAuthService{})

	t.Run("dispatches to NotFound for 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.RenderError(rec, req, http.StatusNotFound)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "Page Not Found")
	})

	t.Run("dispatches to Forbidden for 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.RenderError(rec, req, http.StatusForbidden)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Forbidden")
	})

	t.Run("dispatches to ServerError for 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.RenderError(rec, req, http.StatusInternalServerError)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Internal Server Error")
	})

	t.Run("dispatches to ServerError for unknown codes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.RenderError(rec, req, http.StatusBadGateway)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Internal Server Error")
	})
}
