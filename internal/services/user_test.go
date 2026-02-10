package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// mockUserRepository is a mock implementation of UserRepository for testing
type mockUserRepository struct {
	users       map[string]*models.User
	usersByName map[string]*models.User
	createErr   error
	findErr     error
	updateErr   error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:       make(map[string]*models.User),
		usersByName: make(map[string]*models.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if user.ID == "" {
		user.ID = "test-user-id"
	}
	m.users[user.ID] = user
	m.usersByName[user.Username] = user
	return nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	user, exists := m.users[id]
	if !exists {
		return nil, models.NewNotFoundError("User", id)
	}
	return user, nil
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, models.NewNotFoundError("User", email)
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	user, exists := m.usersByName[username]
	if !exists {
		return nil, models.NewNotFoundError("User", username)
	}
	return user, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *models.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, exists := m.users[user.ID]; !exists {
		return models.NewNotFoundError("User", user.ID)
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	if _, exists := m.users[id]; !exists {
		return models.NewNotFoundError("User", id)
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	if _, exists := m.users[id]; !exists {
		return models.NewNotFoundError("User", id)
	}
	return nil
}

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("creates user with hashed password", func(t *testing.T) {
		repo := newMockUserRepository()
		service := NewUserService(repo)

		user := &models.User{
			Username: "testuser",
			Email:    "test@example.com",
		}

		err := service.CreateUser(ctx, user, "ValidPass1")
		require.NoError(t, err)

		// Verify password is hashed
		assert.NotEmpty(t, user.PasswordHash)
		assert.NotEqual(t, "ValidPass1", user.PasswordHash)

		// Verify hash is valid bcrypt
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("ValidPass1"))
		assert.NoError(t, err)
	})

	t.Run("returns validation error for weak password", func(t *testing.T) {
		repo := newMockUserRepository()
		service := NewUserService(repo)

		user := &models.User{
			Username: "testuser",
			Email:    "test@example.com",
		}

		err := service.CreateUser(ctx, user, "weak")
		assert.Error(t, err)

		var validationErr *models.ValidationError
		assert.ErrorAs(t, err, &validationErr)
	})

	t.Run("returns repository error on create failure", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.createErr = models.NewDatabaseError("create", assert.AnError)
		service := NewUserService(repo)

		user := &models.User{
			Username: "testuser",
			Email:    "test@example.com",
		}

		err := service.CreateUser(ctx, user, "ValidPass1")
		assert.Error(t, err)
	})
}

func TestUserService_GetUser(t *testing.T) {
	ctx := context.Background()

	t.Run("returns user by ID", func(t *testing.T) {
		repo := newMockUserRepository()
		existingUser := &models.User{
			ID:           "user-123",
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hashed",
		}
		repo.users[existingUser.ID] = existingUser
		service := NewUserService(repo)

		user, err := service.GetUser(ctx, "user-123")
		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("returns not found error for non-existent user", func(t *testing.T) {
		repo := newMockUserRepository()
		service := NewUserService(repo)

		_, err := service.GetUser(ctx, "non-existent")
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("updates user fields", func(t *testing.T) {
		repo := newMockUserRepository()
		existingUser := &models.User{
			ID:           "user-123",
			Username:     "olduser",
			Email:        "old@example.com",
			PasswordHash: "hashed",
		}
		repo.users[existingUser.ID] = existingUser
		service := NewUserService(repo)

		newUsername := "newuser"
		newEmail := "new@example.com"
		updates := UserUpdates{
			Username: &newUsername,
			Email:    &newEmail,
		}

		updatedUser, err := service.UpdateUser(ctx, "user-123", updates)
		require.NoError(t, err)
		assert.Equal(t, "newuser", updatedUser.Username)
		assert.Equal(t, "new@example.com", updatedUser.Email)
	})

	t.Run("partial update preserves unchanged fields", func(t *testing.T) {
		repo := newMockUserRepository()
		existingUser := &models.User{
			ID:           "user-123",
			Username:     "olduser",
			Email:        "old@example.com",
			FirstName:    "John",
			LastName:     "Doe",
			PasswordHash: "hashed",
		}
		repo.users[existingUser.ID] = existingUser
		service := NewUserService(repo)

		newFirstName := "Jane"
		updates := UserUpdates{
			FirstName: &newFirstName,
		}

		updatedUser, err := service.UpdateUser(ctx, "user-123", updates)
		require.NoError(t, err)
		assert.Equal(t, "olduser", updatedUser.Username)      // unchanged
		assert.Equal(t, "old@example.com", updatedUser.Email) // unchanged
		assert.Equal(t, "Jane", updatedUser.FirstName)        // updated
		assert.Equal(t, "Doe", updatedUser.LastName)          // unchanged
	})

	t.Run("returns not found for non-existent user", func(t *testing.T) {
		repo := newMockUserRepository()
		service := NewUserService(repo)

		_, err := service.UpdateUser(ctx, "non-existent", UserUpdates{})
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	ctx := context.Background()

	t.Run("changes password successfully", func(t *testing.T) {
		repo := newMockUserRepository()
		// Create a hashed password to simulate existing user
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("OldPass1"), bcrypt.DefaultCost)
		existingUser := &models.User{
			ID:           "user-123",
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
		}
		repo.users[existingUser.ID] = existingUser
		service := NewUserService(repo)

		err := service.ChangePassword(ctx, "user-123", "OldPass1", "NewPass1")
		require.NoError(t, err)

		// Verify new password hash
		updatedUser := repo.users["user-123"]
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("NewPass1"))
		assert.NoError(t, err)
	})

	t.Run("returns error for incorrect current password", func(t *testing.T) {
		repo := newMockUserRepository()
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("OldPass1"), bcrypt.DefaultCost)
		existingUser := &models.User{
			ID:           "user-123",
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
		}
		repo.users[existingUser.ID] = existingUser
		service := NewUserService(repo)

		err := service.ChangePassword(ctx, "user-123", "WrongPass1", "NewPass1")
		assert.Error(t, err)

		var validationErr *models.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "current_password", validationErr.Field)
	})

	t.Run("returns validation error for weak new password", func(t *testing.T) {
		repo := newMockUserRepository()
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("OldPass1"), bcrypt.DefaultCost)
		existingUser := &models.User{
			ID:           "user-123",
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
		}
		repo.users[existingUser.ID] = existingUser
		service := NewUserService(repo)

		err := service.ChangePassword(ctx, "user-123", "OldPass1", "weak")
		assert.Error(t, err)

		var validationErr *models.ValidationError
		assert.ErrorAs(t, err, &validationErr)
	})

	t.Run("returns not found for non-existent user", func(t *testing.T) {
		repo := newMockUserRepository()
		service := NewUserService(repo)

		err := service.ChangePassword(ctx, "non-existent", "OldPass1", "NewPass1")
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestHashPassword(t *testing.T) {
	t.Run("returns valid bcrypt hash", func(t *testing.T) {
		hash, err := hashPassword("TestPass1")
		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("TestPass1"))
		assert.NoError(t, err)
	})
}

func TestVerifyPassword(t *testing.T) {
	t.Run("returns nil for matching password", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("TestPass1"), bcrypt.DefaultCost)
		err := verifyPassword(string(hash), "TestPass1")
		assert.NoError(t, err)
	})

	t.Run("returns error for non-matching password", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("TestPass1"), bcrypt.DefaultCost)
		err := verifyPassword(string(hash), "WrongPass")
		assert.Error(t, err)
	})
}
