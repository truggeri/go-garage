package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceAuthorizationGuard(t *testing.T) {
	t.Run("allows request when ownership check passes", func(t *testing.T) {
		checker := func(accountID string, r *http.Request) (bool, error) {
			return accountID == "user-123", nil
		}

		handlerCalled := false
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := ResourceAuthorizationGuard(checker)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/resource/abc", nil)
		ctx := context.WithValue(req.Context(), AccountContextKey, &AccountInfo{ID: "user-123", Name: "testuser"})
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, handlerCalled)
	})

	t.Run("blocks request when ownership check fails", func(t *testing.T) {
		checker := func(accountID string, r *http.Request) (bool, error) {
			return false, nil
		}

		handlerCalled := false
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := ResourceAuthorizationGuard(checker)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/resource/abc", nil)
		ctx := context.WithValue(req.Context(), AccountContextKey, &AccountInfo{ID: "user-456", Name: "otheruser"})
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.False(t, handlerCalled)
		assert.Contains(t, rec.Body.String(), "do not have permission")
	})

	t.Run("returns error when no authentication context", func(t *testing.T) {
		checker := func(accountID string, r *http.Request) (bool, error) {
			return true, nil
		}

		handlerCalled := false
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := ResourceAuthorizationGuard(checker)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/resource/abc", nil)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.False(t, handlerCalled)
		assert.Contains(t, rec.Body.String(), "authentication required")
	})

	t.Run("returns server error when checker fails", func(t *testing.T) {
		checker := func(accountID string, r *http.Request) (bool, error) {
			return false, errors.New("database error")
		}

		handlerCalled := false
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := ResourceAuthorizationGuard(checker)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/resource/abc", nil)
		ctx := context.WithValue(req.Context(), AccountContextKey, &AccountInfo{ID: "user-123", Name: "testuser"})
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.False(t, handlerCalled)
		assert.Contains(t, rec.Body.String(), "authorization check failed")
	})
}
