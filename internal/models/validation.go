package models

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	// emailRegex is a simple regex for email validation
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	// usernameRegex validates alphanumeric usernames with underscores and hyphens
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	// vinRegex validates 17-character VINs (excludes I, O, Q)
	vinRegex = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)
)

const (
	minUsernameLength = 3
	maxUsernameLength = 30
	minPasswordLength = 8
	vinLength         = 17
	minYear           = 1900
)

// ValidateUser validates a User model
func ValidateUser(u *User) error {
	if u.Username == "" {
		return NewValidationError("username", "username is required")
	}

	if utf8.RuneCountInString(u.Username) < minUsernameLength {
		return NewValidationError("username", "username must be at least 3 characters long")
	}

	if utf8.RuneCountInString(u.Username) > maxUsernameLength {
		return NewValidationError("username", "username must be at most 30 characters long")
	}

	if !usernameRegex.MatchString(u.Username) {
		return NewValidationError("username", "username can only contain alphanumeric characters, underscores, and hyphens")
	}

	if u.Email == "" {
		return NewValidationError("email", "email is required")
	}

	if !emailRegex.MatchString(u.Email) {
		return NewValidationError("email", "invalid email format")
	}

	// PasswordHash is validated during user creation, not on the model itself
	// since we don't want to enforce hash format requirements

	return nil
}

// ValidatePassword validates a plain text password before hashing
func ValidatePassword(password string) error {
	if password == "" {
		return NewValidationError("password", "password is required")
	}

	if utf8.RuneCountInString(password) < minPasswordLength {
		return NewValidationError("password", "password must be at least 8 characters long")
	}

	// Check for at least one uppercase, one lowercase, and one digit
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return NewValidationError("password", "password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}

	return nil
}

// ValidateVehicle validates a Vehicle model
func ValidateVehicle(v *Vehicle) error {
	if v.UserID == "" {
		return NewValidationError("user_id", "user ID is required")
	}

	if v.VIN == "" {
		return NewValidationError("vin", "VIN is required")
	}

	// Normalize VIN to uppercase and remove spaces
	normalizedVIN := strings.ToUpper(strings.ReplaceAll(v.VIN, " ", ""))

	if len(normalizedVIN) != vinLength {
		return NewValidationError("vin", "VIN must be exactly 17 characters")
	}

	if !vinRegex.MatchString(normalizedVIN) {
		return NewValidationError("vin", "VIN contains invalid characters (cannot include I, O, or Q)")
	}

	if v.Make == "" {
		return NewValidationError("make", "make is required")
	}

	if v.Model == "" {
		return NewValidationError("model", "model is required")
	}

	if v.Year == 0 {
		return NewValidationError("year", "year is required")
	}

	currentYear := time.Now().Year()
	if v.Year < minYear || v.Year > currentYear+1 {
		return NewValidationError("year", "year must be between 1900 and current year + 1")
	}

	if v.Status == "" {
		return NewValidationError("status", "status is required")
	}

	if v.Status != VehicleStatusActive && v.Status != VehicleStatusSold && v.Status != VehicleStatusScrapped {
		return NewValidationError("status", "status must be 'active', 'sold', or 'scrapped'")
	}

	if v.PurchasePrice != nil && *v.PurchasePrice < 0 {
		return NewValidationError("purchase_price", "purchase price cannot be negative")
	}

	if v.PurchaseMileage != nil && *v.PurchaseMileage < 0 {
		return NewValidationError("purchase_mileage", "purchase mileage cannot be negative")
	}

	if v.CurrentMileage != nil && *v.CurrentMileage < 0 {
		return NewValidationError("current_mileage", "current mileage cannot be negative")
	}

	// Validate current mileage is not less than purchase mileage
	if v.PurchaseMileage != nil && v.CurrentMileage != nil && *v.CurrentMileage < *v.PurchaseMileage {
		return NewValidationError("current_mileage", "current mileage cannot be less than purchase mileage")
	}

	return nil
}

// ValidateMaintenanceRecord validates a MaintenanceRecord model
func ValidateMaintenanceRecord(m *MaintenanceRecord) error {
	if m.VehicleID == "" {
		return NewValidationError("vehicle_id", "vehicle ID is required")
	}

	if m.ServiceType == "" {
		return NewValidationError("service_type", "service type is required")
	}

	if m.ServiceDate.IsZero() {
		return NewValidationError("service_date", "service date is required")
	}

	// Service date cannot be in the future
	if m.ServiceDate.After(time.Now()) {
		return NewValidationError("service_date", "service date cannot be in the future")
	}

	if m.Cost != nil && *m.Cost < 0 {
		return NewValidationError("cost", "cost cannot be negative")
	}

	if m.MileageAtService != nil && *m.MileageAtService < 0 {
		return NewValidationError("mileage_at_service", "mileage at service cannot be negative")
	}

	return nil
}
