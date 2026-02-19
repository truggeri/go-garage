package services

import (
	"context"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// UserService defines the interface for user business logic
type UserService interface {
	// CreateUser creates a new user with password hashing
	CreateUser(ctx context.Context, user *models.User, password string) error

	// GetUser retrieves a user by their ID
	GetUser(ctx context.Context, id string) (*models.User, error)

	// UpdateUser updates a user's information
	UpdateUser(ctx context.Context, id string, updates UserUpdates) (*models.User, error)

	// ChangePassword updates a user's password
	ChangePassword(ctx context.Context, id string, currentPassword, newPassword string) error

	// DeleteUser deletes a user account after verifying password
	DeleteUser(ctx context.Context, id string, password string) error
}

// UserUpdates contains the fields that can be updated for a user
type UserUpdates struct {
	Username  *string
	Email     *string
	FirstName *string
	LastName  *string
}

// DefaultUserService implements UserService using a UserRepository
type DefaultUserService struct {
	repo repositories.UserRepository
}

// NewUserService creates a new DefaultUserService
func NewUserService(repo repositories.UserRepository) *DefaultUserService {
	return &DefaultUserService{repo: repo}
}

// CreateUser creates a new user with password hashing
func (s *DefaultUserService) CreateUser(ctx context.Context, user *models.User, password string) error {
	// Validate the password before hashing
	if err := models.ValidatePassword(password); err != nil {
		return err
	}

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	user.PasswordHash = hashedPassword

	// Create the user
	return s.repo.Create(ctx, user)
}

// GetUser retrieves a user by their ID
func (s *DefaultUserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateUser updates a user's information
func (s *DefaultUserService) UpdateUser(ctx context.Context, id string, updates UserUpdates) (*models.User, error) {
	// Get the existing user
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.Username != nil {
		user.Username = *updates.Username
	}
	if updates.Email != nil {
		user.Email = *updates.Email
	}
	if updates.FirstName != nil {
		user.FirstName = *updates.FirstName
	}
	if updates.LastName != nil {
		user.LastName = *updates.LastName
	}

	// Update the user
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword updates a user's password
func (s *DefaultUserService) ChangePassword(ctx context.Context, id string, currentPassword, newPassword string) error {
	// Get the user to verify current password
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify current password
	if verifyErr := verifyPassword(user.PasswordHash, currentPassword); verifyErr != nil {
		return models.NewValidationError("current_password", "current password is incorrect")
	}

	// Validate new password
	if validateErr := models.ValidatePassword(newPassword); validateErr != nil {
		return validateErr
	}

	// Hash the new password
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hashedPassword

	// Update the user
	return s.repo.Update(ctx, user)
}

// DeleteUser deletes a user account after verifying password
func (s *DefaultUserService) DeleteUser(ctx context.Context, id string, password string) error {
	// Get the user to verify password
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify password
	if verifyErr := verifyPassword(user.PasswordHash, password); verifyErr != nil {
		return models.NewValidationError("password", "password is incorrect")
	}

	// Delete the user (cascades to vehicles and maintenance)
	return s.repo.Delete(ctx, id)
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", models.NewDatabaseError("hash password", err)
	}
	return string(bytes), nil
}

// verifyPassword verifies a password against a bcrypt hash
func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
