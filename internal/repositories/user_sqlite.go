package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/truggeri/go-garage/internal/models"
)

// SQLiteUserRepository implements UserRepository using SQLite
type SQLiteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository creates a new SQLite-based user repository
func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

// Create inserts a new user into the database
func (r *SQLiteUserRepository) Create(ctx context.Context, user *models.User) error {
	if err := models.ValidateUser(user); err != nil {
		return err
	}

	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (id, username, email, password_hash, first_name, last_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "username") {
				return models.NewDuplicateError("User", "username", user.Username)
			}
			if strings.Contains(err.Error(), "email") {
				return models.NewDuplicateError("User", "email", user.Email)
			}
		}
		return models.NewDatabaseError("create user", err)
	}

	return nil
}

// FindByID retrieves a user by their ID
func (r *SQLiteUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, 
		       created_at, updated_at, last_login_at
		FROM users
		WHERE id = ?
	`

	user := &models.User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewNotFoundError("User", id)
		}
		return nil, models.NewDatabaseError("find user by ID", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// FindByEmail retrieves a user by their email address
func (r *SQLiteUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, 
		       created_at, updated_at, last_login_at
		FROM users
		WHERE email = ?
	`

	user := &models.User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewNotFoundError("User", email)
		}
		return nil, models.NewDatabaseError("find user by email", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// FindByUsername retrieves a user by their username
func (r *SQLiteUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, 
		       created_at, updated_at, last_login_at
		FROM users
		WHERE username = ?
	`

	user := &models.User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewNotFoundError("User", username)
		}
		return nil, models.NewDatabaseError("find user by username", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// Update modifies an existing user's information
func (r *SQLiteUserRepository) Update(ctx context.Context, user *models.User) error {
	if err := models.ValidateUser(user); err != nil {
		return err
	}

	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = ?, email = ?, password_hash = ?, 
		    first_name = ?, last_name = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "username") {
				return models.NewDuplicateError("User", "username", user.Username)
			}
			if strings.Contains(err.Error(), "email") {
				return models.NewDuplicateError("User", "email", user.Email)
			}
		}
		return models.NewDatabaseError("update user", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.NewDatabaseError("update user check", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("User", user.ID)
	}

	return nil
}

// Delete removes a user from the database
func (r *SQLiteUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return models.NewDatabaseError("delete user", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.NewDatabaseError("delete user check", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("User", id)
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *SQLiteUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET last_login_at = ? WHERE id = ?`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return models.NewDatabaseError("update last login", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.NewDatabaseError("update last login check", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("User", id)
	}

	return nil
}
