package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestProfileEditPageHandler(
	t *testing.T,
	userSvc *stubUserSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, &stubVehicleSvc{}, &stubMaintenanceSvc{}, nil, userSvc, nil)
}

func TestPageHandler_ProfileEdit(t *testing.T) {
	now := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("renders edit form pre-populated with user data", func(t *testing.T) {
		userStub := &stubUserSvc{
			getResult: &models.User{
				ID:        "u1",
				Username:  "johndoe",
				Email:     "john@example.com",
				FirstName: "John",
				LastName:  "Doe",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
		handler := newTestProfileEditPageHandler(t, userStub)

		req := httptest.NewRequest(http.MethodGet, "/profile/edit", nil)
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		serveHTMXPage(t, handler.ProfileEdit, rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "johndoe")
		assert.Contains(t, body, "john@example.com")
		assert.Contains(t, body, "John")
		assert.Contains(t, body, "Doe")
		assert.Contains(t, body, `hx-put="/api/v1/users/me"`)
		assert.Contains(t, body, `hx-ext="json-enc"`)
		assert.Contains(t, body, `hx-swap="none"`)
		assert.Contains(t, body, `hx-validate="true"`)
		assert.Contains(t, body, `data-form-errors`)
		for _, field := range []string{"username", "email", "first_name", "last_name"} {
			assert.Contains(t, body, `data-field="`+field+`"`)
		}
		assert.NotContains(t, body, `action="/profile/edit"`)
		assert.NotContains(t, body, `name="csrf_token"`)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestProfileEditPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodGet, "/profile/edit", nil)
		rec := httptest.NewRecorder()

		handler.ProfileEdit(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when user service fails", func(t *testing.T) {
		userStub := &stubUserSvc{getErr: models.NewDatabaseError("get user", assert.AnError)}
		handler := newTestProfileEditPageHandler(t, userStub)

		req := httptest.NewRequest(http.MethodGet, "/profile/edit", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.ProfileEdit(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
