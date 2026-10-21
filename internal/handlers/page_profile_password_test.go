package handlers

import (
	"net/http"
	"net/http/httptest"
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
	return NewPageHandler(engine, &mockAuthService{}, &stubVehicleSvc{}, &stubMaintenanceSvc{}, nil, userSvc, nil)
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
		assert.Contains(t, body, `hx-put="/api/v1/users/me/password"`)
		assert.Contains(t, body, `hx-ext="json-enc"`)
		assert.Contains(t, body, `hx-swap="none"`)
		assert.Contains(t, body, `hx-validate="true"`)
		assert.Contains(t, body, `data-form-errors role="alert"`)
		assert.NotContains(t, body, `action="/profile/password"`)
		assert.NotContains(t, body, `name="csrf_token"`)
		for _, field := range []string{"current_password", "new_password", "confirm_password"} {
			assert.Contains(t, body, `data-field="`+field+`"`)
		}
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestChangePasswordPageHandler(t, &stubUserSvc{})

		req := httptest.NewRequest(http.MethodGet, "/profile/password", nil)
		rec := httptest.NewRecorder()

		handler.ChangePassword(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestUserHandler_ChangePassword_HTMXErrors(t *testing.T) {
	for _, test := range []struct {
		name, body, field string
		err               error
		status            int
	}{
		{"missing current password", `{"new_password":"NewPass1","confirm_password":"NewPass1"}`, "current_password", nil, 400},
		{"missing new password", `{"current_password":"OldPass1"}`, "new_password", nil, 400},
		{"missing confirmation", `{"current_password":"OldPass1","new_password":"NewPass1"}`, "confirm_password", nil, 400},
		{"mismatched confirmation", `{"current_password":"OldPass1","new_password":"NewPass1","confirm_password":"Mismatch1"}`, "confirm_password", nil, 400},
		{"invalid confirmation type", `{"current_password":"OldPass1","new_password":"NewPass1","confirm_password":123}`, "confirm_password", nil, 400},
		{"incorrect current password", `{"current_password":"WrongPass1","new_password":"NewPass1","confirm_password":"NewPass1"}`, "current_password", models.NewValidationError("current_password", "current password is incorrect"), 400},
		{"weak new password", `{"current_password":"OldPass1","new_password":"short","confirm_password":"short"}`, "new_password", models.NewValidationError("password", "password must be at least 8 characters long"), 400},
		{"unexpected failure", `{"current_password":"OldPass1","new_password":"NewPass1","confirm_password":"NewPass1"}`, "", models.NewDatabaseError("change password", assert.AnError), 500},
		{"malformed JSON", `{"current_password":"OldPass1",`, "", nil, 400},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := MakeUserAPIHandler(&stubUserSvc{changePassErr: test.err})
			req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", strings.NewReader(test.body))
			req = addAuthContext(req, "u1", "johndoe")
			req.Header.Set("HX-Request", "true")
			rec := httptest.NewRecorder()
			handler.ChangePassword(rec, req)

			assert.Equal(t, test.status, rec.Code)
			assert.Empty(t, rec.Header().Get("HX-Redirect"))
			if test.field != "" {
				assert.Contains(t, rec.Body.String(), `"field":"`+test.field+`"`)
			}
			for _, password := range []string{"OldPass1", "NewPass1", "Mismatch1", "WrongPass1", "short"} {
				assert.NotContains(t, rec.Body.String(), password)
			}
		})
	}
}
