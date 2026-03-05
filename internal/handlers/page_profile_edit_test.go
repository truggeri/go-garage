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
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestProfileEditPageHandler(
	t *testing.T,
	userSvc *stubUserSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, &stubVehicleSvc{}, &stubMaintenanceSvc{}, userSvc, nil)
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

		handler.ProfileEdit(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "johndoe")
		assert.Contains(t, body, "john@example.com")
		assert.Contains(t, body, "John")
		assert.Contains(t, body, "Doe")
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

func TestPageHandler_ProfileUpdate(t *testing.T) {
	now := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	makeForm := func(username, email, firstName, lastName string) *strings.Reader {
		form := url.Values{}
		form.Set("username", username)
		form.Set("email", email)
		form.Set("first_name", firstName)
		form.Set("last_name", lastName)
		return strings.NewReader(form.Encode())
	}

	t.Run("redirects to profile on success", func(t *testing.T) {
		userStub := &stubUserSvc{
			updateResult: &models.User{
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

		req := httptest.NewRequest(http.MethodPost, "/profile/edit", makeForm("johndoe", "john@example.com", "John", "Doe"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ProfileUpdate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Contains(t, rec.Header().Get("Location"), "/profile")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestProfileEditPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodPost, "/profile/edit", makeForm("johndoe", "john@example.com", "", ""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.ProfileUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 400 for missing username", func(t *testing.T) {
		handler := newTestProfileEditPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodPost, "/profile/edit", makeForm("", "john@example.com", "", ""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ProfileUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "username")
	})

	t.Run("returns 400 for invalid email", func(t *testing.T) {
		handler := newTestProfileEditPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodPost, "/profile/edit", makeForm("johndoe", "not-an-email", "", ""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ProfileUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "email")
	})

	t.Run("returns 409 for duplicate username", func(t *testing.T) {
		userStub := &stubUserSvc{
			updateErr: models.NewDuplicateError("User", "username", "johndoe"),
		}
		handler := newTestProfileEditPageHandler(t, userStub)

		req := httptest.NewRequest(http.MethodPost, "/profile/edit", makeForm("johndoe", "john@example.com", "", ""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ProfileUpdate(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("returns 500 when update service fails", func(t *testing.T) {
		userStub := &stubUserSvc{
			updateErr: models.NewDatabaseError("update user", assert.AnError),
		}
		handler := newTestProfileEditPageHandler(t, userStub)

		req := httptest.NewRequest(http.MethodPost, "/profile/edit", makeForm("johndoe", "john@example.com", "", ""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ProfileUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
