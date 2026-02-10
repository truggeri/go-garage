package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid password",
			password:    "Password123",
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: true,
			errorMsg:    "password is required",
		},
		{
			name:        "too short",
			password:    "Pass1",
			expectError: true,
			errorMsg:    "at least 8 characters",
		},
		{
			name:        "no uppercase",
			password:    "password123",
			expectError: true,
			errorMsg:    "uppercase letter",
		},
		{
			name:        "no lowercase",
			password:    "PASSWORD123",
			expectError: true,
			errorMsg:    "lowercase letter",
		},
		{
			name:        "no digit",
			password:    "PasswordABC",
			expectError: true,
			errorMsg:    "digit",
		},
		{
			name:        "special characters ok",
			password:    "Pass@word123!",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUser(t *testing.T) {
	tests := []struct {
		user        *User
		name        string
		expectError bool
		errorField  string
	}{
		{
			name: "valid user",
			user: &User{
				Username:     "johndoe",
				Email:        "john@example.com",
				PasswordHash: "hashed",
			},
			expectError: false,
		},
		{
			name: "empty username",
			user: &User{
				Username: "",
				Email:    "john@example.com",
			},
			expectError: true,
			errorField:  "username",
		},
		{
			name: "username too short",
			user: &User{
				Username: "ab",
				Email:    "john@example.com",
			},
			expectError: true,
			errorField:  "username",
		},
		{
			name: "username too long",
			user: &User{
				Username: "this_is_a_very_long_username_that_exceeds_limit",
				Email:    "john@example.com",
			},
			expectError: true,
			errorField:  "username",
		},
		{
			name: "username with invalid characters",
			user: &User{
				Username: "john.doe@",
				Email:    "john@example.com",
			},
			expectError: true,
			errorField:  "username",
		},
		{
			name: "empty email",
			user: &User{
				Username: "johndoe",
				Email:    "",
			},
			expectError: true,
			errorField:  "email",
		},
		{
			name: "invalid email format",
			user: &User{
				Username: "johndoe",
				Email:    "not-an-email",
			},
			expectError: true,
			errorField:  "email",
		},
		{
			name: "username with underscore",
			user: &User{
				Username: "john_doe",
				Email:    "john@example.com",
			},
			expectError: false,
		},
		{
			name: "username with hyphen",
			user: &User{
				Username: "john-doe",
				Email:    "john@example.com",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUser(tt.user)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
					assert.Contains(t, err.Error(), tt.errorField)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
