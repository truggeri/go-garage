package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testUserID = "user-123"
const testResource = "my-vehicle"

func TestResourceAuthorizationGuard(t *testing.T) {
	t.Run("allows request when ownership check passes", func(t *testing.T) {
		checker := func(accountID string, r *http.Request) (bool, error) {
			return accountID == testUserID, nil
		}

		handlerCalled := false
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		guardedHandler := ResourceAuthorizationGuard(checker)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/resource/abc", nil)
		ctx := context.WithValue(req.Context(), AccountContextKey, &AccountInfo{ID: testUserID, Name: "testuser"})
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
		ctx := context.WithValue(req.Context(), AccountContextKey, &AccountInfo{ID: testUserID, Name: "testuser"})
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		guardedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.False(t, handlerCalled)
		assert.Contains(t, rec.Body.String(), "authorization check failed")
	})
}

func TestPageResourceOwnershipGuard(t *testing.T) {
	authCtx := func(r *http.Request, id, name string) *http.Request {
		ctx := context.WithValue(r.Context(), AccountContextKey, &AccountInfo{ID: id, Name: name})
		return r.WithContext(ctx)
	}

	t.Run("allows request and stores resource in context", func(t *testing.T) {
		lookup := func(_ context.Context, _ *http.Request) (interface{}, string, error) {
			return testResource, testUserID, nil
		}

		var gotResource interface{}
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotResource, _ = GetLoadedResourceFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		handler := PageResourceOwnershipGuard(lookup)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1", nil)
		req = authCtx(req, testUserID, "testuser")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, testResource, gotResource)
	})

	t.Run("returns 403 when owner does not match", func(t *testing.T) {
		lookup := func(_ context.Context, _ *http.Request) (interface{}, string, error) {
			return testResource, "other-user", nil
		}

		handlerCalled := false
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		handler := PageResourceOwnershipGuard(lookup)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1", nil)
		req = authCtx(req, testUserID, "testuser")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.False(t, handlerCalled)
	})

	t.Run("returns 404 when resource not found", func(t *testing.T) {
		lookup := func(_ context.Context, _ *http.Request) (interface{}, string, error) {
			return nil, "", ErrResourceNotFound
		}

		handlerCalled := false
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		handler := PageResourceOwnershipGuard(lookup)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v99", nil)
		req = authCtx(req, testUserID, "testuser")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, handlerCalled)
	})

	t.Run("returns 500 when lookup returns unexpected error", func(t *testing.T) {
		lookup := func(_ context.Context, _ *http.Request) (interface{}, string, error) {
			return nil, "", errors.New("database error")
		}

		handlerCalled := false
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		handler := PageResourceOwnershipGuard(lookup)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1", nil)
		req = authCtx(req, testUserID, "testuser")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.False(t, handlerCalled)
	})

	t.Run("returns 500 when no authentication context", func(t *testing.T) {
		lookup := func(_ context.Context, _ *http.Request) (interface{}, string, error) {
			return testResource, testUserID, nil
		}

		handlerCalled := false
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		handler := PageResourceOwnershipGuard(lookup)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.False(t, handlerCalled)
	})

	t.Run("returns 404 for wrapped not-found error", func(t *testing.T) {
		lookup := func(_ context.Context, _ *http.Request) (interface{}, string, error) {
			return nil, "", fmt.Errorf("vehicle: %w", ErrResourceNotFound)
		}

		handlerCalled := false
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		handler := PageResourceOwnershipGuard(lookup)(innerHandler)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v99", nil)
		req = authCtx(req, testUserID, "testuser")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, handlerCalled)
	})
}
