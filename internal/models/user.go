package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}
