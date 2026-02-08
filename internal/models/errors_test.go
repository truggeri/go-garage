package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundError(t *testing.T) {
	err := NewNotFoundError("Vehicle", "123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Vehicle")
	assert.Contains(t, err.Error(), "123")
	assert.Contains(t, err.Error(), "not found")
}

func TestValidationError(t *testing.T) {
	t.Run("with field", func(t *testing.T) {
		err := NewValidationError("email", "invalid format")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
		assert.Contains(t, err.Error(), "invalid format")
	})

	t.Run("without field", func(t *testing.T) {
		err := NewValidationError("", "general validation error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "general validation error")
	})
}

func TestDuplicateError(t *testing.T) {
	err := NewDuplicateError("User", "email", "test@example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "User")
	assert.Contains(t, err.Error(), "email")
	assert.Contains(t, err.Error(), "test@example.com")
	assert.Contains(t, err.Error(), "already exists")
}

func TestDatabaseError(t *testing.T) {
	innerErr := errors.New("connection failed")
	err := NewDatabaseError("query", innerErr)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query")
	assert.Contains(t, err.Error(), "connection failed")
	
	// Test unwrap
	unwrappedErr := errors.Unwrap(err)
	assert.Equal(t, innerErr, unwrappedErr)
}
