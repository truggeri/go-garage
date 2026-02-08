package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
)

func TestUserRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	t.Run("create valid user", func(t *testing.T) {
		user := &models.User{
			Username:     "johndoe",
			Email:        "john@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("create with provided ID", func(t *testing.T) {
		user := &models.User{
			ID:           "custom-id-123",
			Username:     "janedoe",
			Email:        "jane@example.com",
			PasswordHash: "hashed_password",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.Equal(t, "custom-id-123", user.ID)
	})

	t.Run("duplicate username", func(t *testing.T) {
		user1 := &models.User{
			Username:     "duplicate",
			Email:        "user1@example.com",
			PasswordHash: "hashed_password",
		}

		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := &models.User{
			Username:     "duplicate",
			Email:        "user2@example.com",
			PasswordHash: "hashed_password",
		}

		err = repo.Create(ctx, user2)
		require.Error(t, err)
		assert.IsType(t, &models.DuplicateError{}, err)
		assert.Contains(t, err.Error(), "username")
	})

	t.Run("duplicate email", func(t *testing.T) {
		user1 := &models.User{
			Username:     "user1",
			Email:        "duplicate@example.com",
			PasswordHash: "hashed_password",
		}

		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := &models.User{
			Username:     "user2",
			Email:        "duplicate@example.com",
			PasswordHash: "hashed_password",
		}

		err = repo.Create(ctx, user2)
		require.Error(t, err)
		assert.IsType(t, &models.DuplicateError{}, err)
		assert.Contains(t, err.Error(), "email")
	})

	t.Run("validation error", func(t *testing.T) {
		user := &models.User{
			Username:     "", // Invalid: empty username
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
		}

		err := repo.Create(ctx, user)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	t.Run("find existing user", func(t *testing.T) {
		user := &models.User{
			Username:     "findme",
			Email:        "findme@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "Find",
			LastName:     "Me",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Username, found.Username)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.FirstName, found.FirstName)
		assert.Equal(t, user.LastName, found.LastName)
		assert.Nil(t, found.LastLoginAt)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "non-existent-id")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	t.Run("find existing user", func(t *testing.T) {
		user := &models.User{
			Username:     "emailuser",
			Email:        "email@example.com",
			PasswordHash: "hashed_password",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := repo.FindByEmail(ctx, "nonexistent@example.com")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestUserRepository_FindByUsername(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	t.Run("find existing user", func(t *testing.T) {
		user := &models.User{
			Username:     "usernameuser",
			Email:        "usernameuser@example.com",
			PasswordHash: "hashed_password",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByUsername(ctx, user.Username)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Username, found.Username)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := repo.FindByUsername(ctx, "nonexistentuser")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestUserRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	t.Run("update existing user", func(t *testing.T) {
		user := &models.User{
			Username:     "updateuser",
			Email:        "update@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "Old",
			LastName:     "Name",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		oldUpdatedAt := user.UpdatedAt
		time.Sleep(10 * time.Millisecond) // Ensure time difference

		user.FirstName = "New"
		user.LastName = "Name"
		user.Email = "newemail@example.com"

		err = repo.Update(ctx, user)
		require.NoError(t, err)
		assert.True(t, user.UpdatedAt.After(oldUpdatedAt))

		// Verify update
		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "New", found.FirstName)
		assert.Equal(t, "Name", found.LastName)
		assert.Equal(t, "newemail@example.com", found.Email)
	})

	t.Run("update non-existent user", func(t *testing.T) {
		user := &models.User{
			ID:           "non-existent",
			Username:     "test",
			Email:        "test@example.com",
			PasswordHash: "hashed",
		}

		err := repo.Update(ctx, user)
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})

	t.Run("update with duplicate email", func(t *testing.T) {
		user1 := &models.User{
			Username:     "user1update",
			Email:        "user1@example.com",
			PasswordHash: "hashed_password",
		}
		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := &models.User{
			Username:     "user2update",
			Email:        "user2@example.com",
			PasswordHash: "hashed_password",
		}
		err = repo.Create(ctx, user2)
		require.NoError(t, err)

		// Try to update user2 with user1's email
		user2.Email = user1.Email
		err = repo.Update(ctx, user2)
		require.Error(t, err)
		assert.IsType(t, &models.DuplicateError{}, err)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	t.Run("delete existing user", func(t *testing.T) {
		user := &models.User{
			Username:     "deleteuser",
			Email:        "delete@example.com",
			PasswordHash: "hashed_password",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.FindByID(ctx, user.ID)
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})

	t.Run("delete non-existent user", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent-id")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	t.Run("update last login for existing user", func(t *testing.T) {
		user := &models.User{
			Username:     "loginuser",
			Email:        "login@example.com",
			PasswordHash: "hashed_password",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Initially, LastLoginAt should be nil
		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, found.LastLoginAt)

		// Update last login
		err = repo.UpdateLastLogin(ctx, user.ID)
		require.NoError(t, err)

		// Verify LastLoginAt is set
		found, err = repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, found.LastLoginAt)
		assert.True(t, found.LastLoginAt.Before(time.Now().Add(time.Second)))
	})

	t.Run("update last login for non-existent user", func(t *testing.T) {
		err := repo.UpdateLastLogin(ctx, "non-existent-id")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}
