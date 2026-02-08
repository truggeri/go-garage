package repositories

import (
	"context"

	"github.com/truggeri/go-garage/internal/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create inserts a new user into the database
	Create(ctx context.Context, user *models.User) error

	// FindByID retrieves a user by their ID
	FindByID(ctx context.Context, id string) (*models.User, error)

	// FindByEmail retrieves a user by their email address
	FindByEmail(ctx context.Context, email string) (*models.User, error)

	// FindByUsername retrieves a user by their username
	FindByUsername(ctx context.Context, username string) (*models.User, error)

	// Update modifies an existing user's information
	Update(ctx context.Context, user *models.User) error

	// Delete removes a user from the database
	Delete(ctx context.Context, id string) error

	// UpdateLastLogin updates the last login timestamp for a user
	UpdateLastLogin(ctx context.Context, id string) error
}
