package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPaging(t *testing.T) {
	t.Run("returns defaults when no query params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		page, limit := extractPaging(req)
		assert.Equal(t, 1, page)
		assert.Equal(t, 20, limit)
	})

	t.Run("extracts valid page and limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?page=3&limit=50", nil)
		page, limit := extractPaging(req)
		assert.Equal(t, 3, page)
		assert.Equal(t, 50, limit)
	})

	t.Run("ignores invalid page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?page=-1", nil)
		page, _ := extractPaging(req)
		assert.Equal(t, 1, page)
	})

	t.Run("ignores invalid limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?limit=200", nil)
		_, limit := extractPaging(req)
		assert.Equal(t, 20, limit)
	})

	t.Run("ignores non-numeric page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?page=abc", nil)
		page, _ := extractPaging(req)
		assert.Equal(t, 1, page)
	})
}
