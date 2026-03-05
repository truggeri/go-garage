package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestChangePasswordPageHandler(
	t *testing.T,
	userSvc *stubUserSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, &stubVehicleSvc{}, &stubMaintenanceSvc{}, userSvc, nil)
}

func TestPageHandler_ChangePassword(t *testing.T) {
	t.Run("renders change password form", func(t *testing.T) {
		handler := newTestChangePasswordPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodGet, "/profile/password", nil)
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ChangePassword(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Change Password")
		assert.Contains(t, body, "current_password")
		assert.Contains(t, body, "new_password")
		assert.Contains(t, body, "confirm_password")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestChangePasswordPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodGet, "/profile/password", nil)
		rec := httptest.NewRecorder()

		handler.ChangePassword(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPageHandler_ChangePasswordSubmit(t *testing.T) {
	makeForm := func(current, newPass, confirm string) *strings.Reader {
		form := url.Values{}
		form.Set("current_password", current)
		form.Set("new_password", newPass)
		form.Set("confirm_password", confirm)
		return strings.NewReader(form.Encode())
	}

	t.Run("redirects to profile on success", func(t *testing.T) {
		handler := newTestChangePasswordPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodPost, "/profile/password", makeForm("OldPass1", "NewPass1", "NewPass1"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ChangePasswordSubmit(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Contains(t, rec.Header().Get("Location"), "/profile?password_changed=true")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestChangePasswordPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodPost, "/profile/password", makeForm("OldPass1", "NewPass1", "NewPass1"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.ChangePasswordSubmit(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 400 for empty current password", func(t *testing.T) {
		handler := newTestChangePasswordPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodPost, "/profile/password", makeForm("", "NewPass1", "NewPass1"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ChangePasswordSubmit(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Current password is required")
	})

	t.Run("returns 400 for empty new password", func(t *testing.T) {
		handler := newTestChangePasswordPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodPost, "/profile/password", makeForm("OldPass1", "", ""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ChangePasswordSubmit(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "New password is required")
	})

	t.Run("returns 400 when passwords do not match", func(t *testing.T) {
		handler := newTestChangePasswordPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodPost, "/profile/password", makeForm("OldPass1", "NewPass1", "Mismatch1"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ChangePasswordSubmit(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Passwords do not match")
	})

	t.Run("returns 400 for incorrect current password", func(t *testing.T) {
		userStub := &stubUserSvc{
			changePassErr: models.NewValidationError("current_password", "current password is incorrect"),
		}
		handler := newTestChangePasswordPageHandler(t, userStub)

		req := httptest.NewRequest(http.MethodPost, "/profile/password", makeForm("WrongPass1", "NewPass1", "NewPass1"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ChangePasswordSubmit(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "current password is incorrect")
	})

	t.Run("returns 400 for weak new password", func(t *testing.T) {
		userStub := &stubUserSvc{
			changePassErr: models.NewValidationError("password", "password must be at least 8 characters long"),
		}
		handler := newTestChangePasswordPageHandler(t, userStub)

		req := httptest.NewRequest(http.MethodPost, "/profile/password", makeForm("OldPass1", "weak", "weak"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ChangePasswordSubmit(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "password must be at least 8 characters long")
	})

	t.Run("returns 500 when service fails unexpectedly", func(t *testing.T) {
		userStub := &stubUserSvc{
			changePassErr: models.NewDatabaseError("change password", assert.AnError),
		}
		handler := newTestChangePasswordPageHandler(t, userStub)

		req := httptest.NewRequest(http.MethodPost, "/profile/password", makeForm("OldPass1", "NewPass1", "NewPass1"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ChangePasswordSubmit(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to change password")
	})
}
