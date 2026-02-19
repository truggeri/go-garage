package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

type stubUserSvc struct {
	getResult     *models.User
	getErr        error
	updateResult  *models.User
	updateErr     error
	changePassErr error
	deleteErr     error
}

func (s *stubUserSvc) CreateUser(_ context.Context, _ *models.User, _ string) error {
	return nil
}

func (s *stubUserSvc) GetUser(_ context.Context, _ string) (*models.User, error) {
	return s.getResult, s.getErr
}

func (s *stubUserSvc) UpdateUser(_ context.Context, _ string, _ services.UserUpdates) (*models.User, error) {
	return s.updateResult, s.updateErr
}

func (s *stubUserSvc) ChangePassword(_ context.Context, _, _, _ string) error {
	return s.changePassErr
}

func (s *stubUserSvc) DeleteUser(_ context.Context, _, _ string) error {
	return s.deleteErr
}

func TestUserHandler_GetMe(t *testing.T) {
	t.Run("returns user profile for authenticated user", func(t *testing.T) {
		now := time.Now()
		stub := &stubUserSvc{
			getResult: &models.User{
				ID:          "user-123",
				Username:    "johndoe",
				Email:       "john@example.com",
				FirstName:   "John",
				LastName:    "Doe",
				CreatedAt:   now,
				UpdatedAt:   now,
				LastLoginAt: &now,
			},
		}
		h := MakeUserAPIHandler(stub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.GetMe(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "user-123", data["id"])
		assert.Equal(t, "johndoe", data["username"])
		assert.Equal(t, "john@example.com", data["email"])
		assert.Equal(t, "John", data["first_name"])
		assert.Equal(t, "Doe", data["last_name"])
		assert.NotNil(t, data["last_login_at"])
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		stub := &stubUserSvc{}
		h := MakeUserAPIHandler(stub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		rec := httptest.NewRecorder()

		h.GetMe(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("returns 404 when user not found", func(t *testing.T) {
		stub := &stubUserSvc{
			getErr: models.NewNotFoundError("User", "user-123"),
		}
		h := MakeUserAPIHandler(stub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.GetMe(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestUserHandler_UpdateMe(t *testing.T) {
	t.Run("updates user profile successfully", func(t *testing.T) {
		now := time.Now()
		stub := &stubUserSvc{
			updateResult: &models.User{
				ID:        "user-123",
				Username:  "johndoe2024",
				Email:     "john.new@example.com",
				FirstName: "John",
				LastName:  "Smith",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{
			"username":  "johndoe2024",
			"email":     "john.new@example.com",
			"last_name": "Smith",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.UpdateMe(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, "Profile updated successfully", resp["message"])

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "johndoe2024", data["username"])
		assert.Equal(t, "john.new@example.com", data["email"])
		assert.Equal(t, "Smith", data["last_name"])
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		stub := &stubUserSvc{}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{"username": "newname"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me", bytes.NewReader(jsonBody))
		rec := httptest.NewRecorder()

		h.UpdateMe(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		stub := &stubUserSvc{}
		h := MakeUserAPIHandler(stub)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me", bytes.NewReader([]byte("invalid json")))
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.UpdateMe(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 409 for duplicate username", func(t *testing.T) {
		stub := &stubUserSvc{
			updateErr: models.NewDuplicateError("User", "username", "johndoe2024"),
		}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{"username": "johndoe2024"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.UpdateMe(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
	})
}

func TestUserHandler_ChangePassword(t *testing.T) {
	t.Run("changes password successfully", func(t *testing.T) {
		stub := &stubUserSvc{
			changePassErr: nil,
		}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{
			"current_password": "OldPass123",
			"new_password":     "NewPass456",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.ChangePassword(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, "Password changed successfully", resp["message"])
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		stub := &stubUserSvc{}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{
			"current_password": "OldPass123",
			"new_password":     "NewPass456",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", bytes.NewReader(jsonBody))
		rec := httptest.NewRecorder()

		h.ChangePassword(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("returns 400 when missing fields", func(t *testing.T) {
		stub := &stubUserSvc{}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{
			"current_password": "OldPass123",
			// missing new_password
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.ChangePassword(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.False(t, resp["success"].(bool))
	})

	t.Run("returns 400 for incorrect current password", func(t *testing.T) {
		stub := &stubUserSvc{
			changePassErr: models.NewValidationError("current_password", "current password is incorrect"),
		}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{
			"current_password": "WrongPass",
			"new_password":     "NewPass456",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.ChangePassword(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestUserHandler_DeleteMe(t *testing.T) {
	t.Run("deletes user account successfully", func(t *testing.T) {
		stub := &stubUserSvc{
			deleteErr: nil,
		}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{
			"password": "MyPass123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/me", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.DeleteMe(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, "Account deleted successfully", resp["message"])
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		stub := &stubUserSvc{}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{"password": "MyPass123"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/me", bytes.NewReader(jsonBody))
		rec := httptest.NewRecorder()

		h.DeleteMe(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("returns 400 when missing password", func(t *testing.T) {
		stub := &stubUserSvc{}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/me", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.DeleteMe(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.False(t, resp["success"].(bool))
	})

	t.Run("returns 400 for incorrect password", func(t *testing.T) {
		stub := &stubUserSvc{
			deleteErr: models.NewValidationError("password", "password is incorrect"),
		}
		h := MakeUserAPIHandler(stub)

		body := map[string]interface{}{
			"password": "WrongPass",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/me", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "user-123", "johndoe")
		rec := httptest.NewRecorder()

		h.DeleteMe(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
