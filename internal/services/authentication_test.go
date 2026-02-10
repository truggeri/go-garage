package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/auth"
	"github.com/truggeri/go-garage/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func buildTestTokenManager(t *testing.T) *auth.TokenManager {
	mgr, err := auth.BuildTokenManager("test-secret-key-for-auth-service", auth.StandardTokenDurations())
	require.NoError(t, err)
	return mgr
}

func TestGarageAuthService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("registers new user and returns tokens", func(t *testing.T) {
		repo := newMockUserRepository()
		tokenMgr := buildTestTokenManager(t)
		svc := BuildAuthService(repo, tokenMgr)

		registration := RegistrationRequest{
			Username:  "newuser",
			Email:     "new@example.com",
			Password:  "StrongPass1",
			FirstName: "New",
			LastName:  "User",
		}

		result, err := svc.Register(ctx, registration)
		require.NoError(t, err)

		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.NotEmpty(t, result.AccountID)
		assert.Equal(t, "newuser", result.AccountName)
		assert.Greater(t, result.AccessExpiresAt, time.Now().Unix())
		assert.Greater(t, result.RefreshExpiresAt, result.AccessExpiresAt)
	})

	t.Run("rejects weak password", func(t *testing.T) {
		repo := newMockUserRepository()
		tokenMgr := buildTestTokenManager(t)
		svc := BuildAuthService(repo, tokenMgr)

		registration := RegistrationRequest{
			Username: "newuser",
			Email:    "new@example.com",
			Password: "weak",
		}

		result, err := svc.Register(ctx, registration)
		assert.Error(t, err)
		assert.Nil(t, result)

		var validationErr *models.ValidationError
		assert.ErrorAs(t, err, &validationErr)
	})

	t.Run("rejects duplicate username", func(t *testing.T) {
		repo := newMockUserRepository()
		existingUser := &models.User{
			ID:           "existing-id",
			Username:     "existinguser",
			Email:        "existing@example.com",
			PasswordHash: "hash",
		}
		repo.users[existingUser.ID] = existingUser
		repo.usersByName[existingUser.Username] = existingUser
		repo.createErr = models.NewDuplicateError("User", "username", "existinguser")

		tokenMgr := buildTestTokenManager(t)
		svc := BuildAuthService(repo, tokenMgr)

		registration := RegistrationRequest{
			Username: "existinguser",
			Email:    "different@example.com",
			Password: "StrongPass1",
		}

		result, err := svc.Register(ctx, registration)
		assert.Error(t, err)
		assert.Nil(t, result)

		var duplicateErr *models.DuplicateError
		assert.ErrorAs(t, err, &duplicateErr)
	})
}

func TestGarageAuthService_Authenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("authenticates with valid email and password", func(t *testing.T) {
		repo := newMockUserRepository()
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("ValidPass1"), bcrypt.DefaultCost)
		existingUser := &models.User{
			ID:           "user-auth-123",
			Username:     "authuser",
			Email:        "auth@example.com",
			PasswordHash: string(hashedPwd),
		}
		repo.users[existingUser.ID] = existingUser
		repo.usersByName[existingUser.Username] = existingUser

		tokenMgr := buildTestTokenManager(t)
		svc := BuildAuthService(repo, tokenMgr)

		result, err := svc.Authenticate(ctx, "auth@example.com", "ValidPass1")
		require.NoError(t, err)

		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.Equal(t, "user-auth-123", result.AccountID)
		assert.Equal(t, "authuser", result.AccountName)
	})

	t.Run("authenticates with valid username and password", func(t *testing.T) {
		repo := newMockUserRepository()
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("ValidPass1"), bcrypt.DefaultCost)
		existingUser := &models.User{
			ID:           "user-auth-123",
			Username:     "authuser",
			Email:        "auth@example.com",
			PasswordHash: string(hashedPwd),
		}
		repo.users[existingUser.ID] = existingUser
		repo.usersByName[existingUser.Username] = existingUser

		tokenMgr := buildTestTokenManager(t)
		svc := BuildAuthService(repo, tokenMgr)

		result, err := svc.Authenticate(ctx, "authuser", "ValidPass1")
		require.NoError(t, err)

		assert.NotEmpty(t, result.AccessToken)
		assert.Equal(t, "user-auth-123", result.AccountID)
	})

	t.Run("rejects invalid password", func(t *testing.T) {
		repo := newMockUserRepository()
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("ValidPass1"), bcrypt.DefaultCost)
		existingUser := &models.User{
			ID:           "user-auth-123",
			Username:     "authuser",
			Email:        "auth@example.com",
			PasswordHash: string(hashedPwd),
		}
		repo.users[existingUser.ID] = existingUser
		repo.usersByName[existingUser.Username] = existingUser

		tokenMgr := buildTestTokenManager(t)
		svc := BuildAuthService(repo, tokenMgr)

		result, err := svc.Authenticate(ctx, "auth@example.com", "WrongPassword1")
		assert.Error(t, err)
		assert.Nil(t, result)

		var validationErr *models.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "credentials", validationErr.Field)
	})

	t.Run("rejects non-existent user", func(t *testing.T) {
		repo := newMockUserRepository()
		tokenMgr := buildTestTokenManager(t)
		svc := BuildAuthService(repo, tokenMgr)

		result, err := svc.Authenticate(ctx, "nonexistent@example.com", "Password1")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGarageAuthService_RefreshSession(t *testing.T) {
	ctx := context.Background()

	t.Run("generates new tokens from valid refresh token", func(t *testing.T) {
		repo := newMockUserRepository()
		tokenMgr := buildTestTokenManager(t)
		svc := BuildAuthService(repo, tokenMgr)

		// First get a valid token bundle
		bundle, err := tokenMgr.GenerateTokenBundle(auth.TokenPayload{
			AccountID:   "user-refresh-123",
			AccountName: "refreshuser",
		})
		require.NoError(t, err)

		result, err := svc.RefreshSession(ctx, bundle.RefreshToken)
		require.NoError(t, err)

		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.Equal(t, "user-refresh-123", result.AccountID)
		assert.Equal(t, "refreshuser", result.AccountName)
	})

	t.Run("rejects invalid refresh token", func(t *testing.T) {
		repo := newMockUserRepository()
		tokenMgr := buildTestTokenManager(t)
		svc := BuildAuthService(repo, tokenMgr)

		result, err := svc.RefreshSession(ctx, "invalid.token.here")
		assert.Error(t, err)
		assert.Nil(t, result)

		var validationErr *models.ValidationError
		assert.ErrorAs(t, err, &validationErr)
	})

	t.Run("rejects access token used as refresh token", func(t *testing.T) {
		repo := newMockUserRepository()
		tokenMgr := buildTestTokenManager(t)
		svc := BuildAuthService(repo, tokenMgr)

		bundle, err := tokenMgr.GenerateTokenBundle(auth.TokenPayload{
			AccountID:   "user-refresh-123",
			AccountName: "refreshuser",
		})
		require.NoError(t, err)

		result, err := svc.RefreshSession(ctx, bundle.AccessToken)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
