package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// mockAuthService is a mock implementation of AuthenticationService
type mockAuthService struct {
	registerResult     *services.AuthenticationResult
	registerErr        error
	authenticateResult *services.AuthenticationResult
	authenticateErr    error
	refreshResult      *services.AuthenticationResult
	refreshErr         error
}

func (m *mockAuthService) Register(ctx context.Context, reg services.RegistrationRequest) (*services.AuthenticationResult, error) {
	if m.registerErr != nil {
		return nil, m.registerErr
	}
	return m.registerResult, nil
}

func (m *mockAuthService) Authenticate(ctx context.Context, identifier, password string) (*services.AuthenticationResult, error) {
	if m.authenticateErr != nil {
		return nil, m.authenticateErr
	}
	return m.authenticateResult, nil
}

func (m *mockAuthService) RefreshSession(ctx context.Context, refreshToken string) (*services.AuthenticationResult, error) {
	if m.refreshErr != nil {
		return nil, m.refreshErr
	}
	return m.refreshResult, nil
}

func TestAuthHandler_HandleRegister(t *testing.T) {
	t.Run("registers user successfully", func(t *testing.T) {
		mockSvc := &mockAuthService{
			registerResult: &services.AuthenticationResult{
				AccessToken:      "access.token.here",
				RefreshToken:     "refresh.token.here",
				AccessExpiresAt:  1234567890,
				RefreshExpiresAt: 1234667890,
				AccountID:        "user-123",
				AccountName:      "testuser",
			},
		}
		handler := BuildAuthHandler(mockSvc)

		reqBody := RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "StrongPass1",
			FirstName: "Test",
			LastName:  "User",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.HandleRegister(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp AuthResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.NotNil(t, resp.Data)
		assert.Equal(t, "access.token.here", resp.Data.AccessToken)
		assert.Equal(t, "user-123", resp.Data.AccountID)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		mockSvc := &mockAuthService{}
		handler := BuildAuthHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte("invalid json")))
		rec := httptest.NewRecorder()

		handler.HandleRegister(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp AuthResponse
		json.NewDecoder(rec.Body).Decode(&resp)
		assert.False(t, resp.Success)
		assert.Equal(t, "INVALID_REQUEST", resp.Error.Code)
	})

	t.Run("returns error for missing required fields", func(t *testing.T) {
		mockSvc := &mockAuthService{}
		handler := BuildAuthHandler(mockSvc)

		reqBody := RegisterRequest{
			Username: "testuser",
			// Missing email and password
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.HandleRegister(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp AuthResponse
		json.NewDecoder(rec.Body).Decode(&resp)
		assert.False(t, resp.Success)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("returns error for duplicate user", func(t *testing.T) {
		mockSvc := &mockAuthService{
			registerErr: models.NewDuplicateError("User", "username", "testuser"),
		}
		handler := BuildAuthHandler(mockSvc)

		reqBody := RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "StrongPass1",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.HandleRegister(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)

		var resp AuthResponse
		json.NewDecoder(rec.Body).Decode(&resp)
		assert.False(t, resp.Success)
		assert.Equal(t, "DUPLICATE_ERROR", resp.Error.Code)
	})

	t.Run("returns error for validation failure", func(t *testing.T) {
		mockSvc := &mockAuthService{
			registerErr: models.NewValidationError("password", "password too weak"),
		}
		handler := BuildAuthHandler(mockSvc)

		reqBody := RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "weak",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.HandleRegister(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp AuthResponse
		json.NewDecoder(rec.Body).Decode(&resp)
		assert.False(t, resp.Success)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestAuthHandler_HandleLogin(t *testing.T) {
	t.Run("logs in user successfully", func(t *testing.T) {
		mockSvc := &mockAuthService{
			authenticateResult: &services.AuthenticationResult{
				AccessToken:      "access.token.here",
				RefreshToken:     "refresh.token.here",
				AccessExpiresAt:  1234567890,
				RefreshExpiresAt: 1234667890,
				AccountID:        "user-123",
				AccountName:      "testuser",
			},
		}
		handler := BuildAuthHandler(mockSvc)

		reqBody := LoginRequest{
			Identifier: "test@example.com",
			Password:   "StrongPass1",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.HandleLogin(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp AuthResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.NotNil(t, resp.Data)
		assert.Equal(t, "access.token.here", resp.Data.AccessToken)
	})

	t.Run("returns error for invalid credentials", func(t *testing.T) {
		mockSvc := &mockAuthService{
			authenticateErr: models.NewValidationError("credentials", "invalid email/username or password"),
		}
		handler := BuildAuthHandler(mockSvc)

		reqBody := LoginRequest{
			Identifier: "test@example.com",
			Password:   "WrongPassword1",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.HandleLogin(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp AuthResponse
		json.NewDecoder(rec.Body).Decode(&resp)
		assert.False(t, resp.Success)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("returns error for missing fields", func(t *testing.T) {
		mockSvc := &mockAuthService{}
		handler := BuildAuthHandler(mockSvc)

		reqBody := LoginRequest{
			Identifier: "test@example.com",
			// Missing password
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.HandleLogin(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAuthHandler_HandleRefresh(t *testing.T) {
	t.Run("refreshes tokens successfully", func(t *testing.T) {
		mockSvc := &mockAuthService{
			refreshResult: &services.AuthenticationResult{
				AccessToken:      "new.access.token",
				RefreshToken:     "new.refresh.token",
				AccessExpiresAt:  1234567890,
				RefreshExpiresAt: 1234667890,
				AccountID:        "user-123",
				AccountName:      "testuser",
			},
		}
		handler := BuildAuthHandler(mockSvc)

		reqBody := RefreshRequest{
			RefreshToken: "old.refresh.token",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.HandleRefresh(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp AuthResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "new.access.token", resp.Data.AccessToken)
	})

	t.Run("returns error for invalid refresh token", func(t *testing.T) {
		mockSvc := &mockAuthService{
			refreshErr: models.NewValidationError("refresh_token", "invalid or expired refresh token"),
		}
		handler := BuildAuthHandler(mockSvc)

		reqBody := RefreshRequest{
			RefreshToken: "invalid.token",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.HandleRefresh(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns error for missing refresh token", func(t *testing.T) {
		mockSvc := &mockAuthService{}
		handler := BuildAuthHandler(mockSvc)

		reqBody := RefreshRequest{}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.HandleRefresh(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAuthHandler_HandleLogout(t *testing.T) {
	t.Run("returns success for logout", func(t *testing.T) {
		mockSvc := &mockAuthService{}
		handler := BuildAuthHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		rec := httptest.NewRecorder()

		handler.HandleLogout(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp AuthResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "Logout successful")
	})
}
