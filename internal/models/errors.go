package models

import (
	"errors"
	"fmt"
)

// NotFoundError indicates a requested resource was not found
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// ValidationError indicates validation failure with details
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// DuplicateError indicates a unique constraint violation
type DuplicateError struct {
	Resource string
	Field    string
	Value    string
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("%s already exists with %s: %s", e.Resource, e.Field, e.Value)
}

// NewDuplicateError creates a new DuplicateError
func NewDuplicateError(resource, field, value string) *DuplicateError {
	return &DuplicateError{
		Resource: resource,
		Field:    field,
		Value:    value,
	}
}

// DatabaseError wraps database operation errors with context
type DatabaseError struct {
	Err       error
	Operation string
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// NewDatabaseError creates a new DatabaseError
func NewDatabaseError(operation string, err error) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Err:       err,
	}
}

// IsNotFoundError checks if an error is a NotFoundError and assigns it to target if so
func IsNotFoundError(err error, target **NotFoundError) bool {
	return errors.As(err, target)
}

// IsValidationError checks if an error is a ValidationError and assigns it to target if so
func IsValidationError(err error, target **ValidationError) bool {
	return errors.As(err, target)
}

// IsDuplicateError checks if an error is a DuplicateError and assigns it to target if so
func IsDuplicateError(err error, target **DuplicateError) bool {
	return errors.As(err, target)
}
