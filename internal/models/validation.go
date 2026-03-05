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

// vehicleValidationFields defines the order in which vehicle validation errors
// are returned by ValidateVehicle.
var vehicleValidationFields = []string{
	"user_id", "vin", "make", "model", "year", "status",
	"purchase_price", "purchase_mileage", "current_mileage",
}

// ValidateVehicleAll validates a Vehicle model and returns all validation errors
// as a map of field name to error message.
func ValidateVehicleAll(v *Vehicle) map[string]string {
	errs := make(map[string]string)

	if v.UserID == "" {
		errs["user_id"] = "user ID is required"
	}

	if v.VIN == "" {
		errs["vin"] = "VIN is required"
	} else {
		// Normalize VIN to uppercase and remove spaces
		normalizedVIN := strings.ToUpper(strings.ReplaceAll(v.VIN, " ", ""))

		if len(normalizedVIN) != vinLength {
			errs["vin"] = "VIN must be exactly 17 characters"
		} else if !vinRegex.MatchString(normalizedVIN) {
			errs["vin"] = "VIN contains invalid characters (cannot include I, O, or Q)"
		}
	}

	if v.Make == "" {
		errs["make"] = "make is required"
	}

	if v.Model == "" {
		errs["model"] = "model is required"
	}

	if v.Year == 0 {
		errs["year"] = "year is required"
	} else {
		currentYear := time.Now().Year()
		if v.Year < minYear || v.Year > currentYear+1 {
			errs["year"] = "year must be between 1900 and current year + 1"
		}
	}

	if v.Status == "" {
		errs["status"] = "status is required"
	} else if v.Status != VehicleStatusActive && v.Status != VehicleStatusSold && v.Status != VehicleStatusScrapped {
		errs["status"] = "status must be 'active', 'sold', or 'scrapped'"
	}

	if v.PurchasePrice != nil && *v.PurchasePrice < 0 {
		errs["purchase_price"] = "purchase price cannot be negative"
	}

	if v.PurchaseMileage != nil && *v.PurchaseMileage < 0 {
		errs["purchase_mileage"] = "purchase mileage cannot be negative"
	}

	if v.CurrentMileage != nil && *v.CurrentMileage < 0 {
		errs["current_mileage"] = "current mileage cannot be negative"
	}

	// Validate current mileage is not less than purchase mileage
	if v.PurchaseMileage != nil && v.CurrentMileage != nil && *v.CurrentMileage < *v.PurchaseMileage {
		errs["current_mileage"] = "current mileage cannot be less than purchase mileage"
	}

	return errs
}

// ValidateVehicle validates a Vehicle model and returns the first error found.
func ValidateVehicle(v *Vehicle) error {
	errs := ValidateVehicleAll(v)
	for _, field := range vehicleValidationFields {
		if msg, ok := errs[field]; ok {
			return NewValidationError(field, msg)
		}
	}
	return nil
}

// ValidateFuelRecord validates a FuelRecord model
func ValidateFuelRecord(f *FuelRecord) error {
	if f.VehicleID == "" {
		return NewValidationError("vehicle_id", "vehicle ID is required")
	}

	if f.FillDate.IsZero() {
		return NewValidationError("fill_date", "fill date is required")
	}

	if f.FillDate.After(time.Now()) {
		return NewValidationError("fill_date", "fill date cannot be in the future")
	}

	if f.Odometer < 0 {
		return NewValidationError("odometer", "odometer must be 0 or greater")
	}

	if f.CostPerUnit < 0 {
		return NewValidationError("cost_per_unit", "cost per unit cannot be negative")
	}

	if f.Volume <= 0 {
		return NewValidationError("volume", "volume must be greater than 0")
	}

	if f.CityDrivingPct != nil && (*f.CityDrivingPct < 0 || *f.CityDrivingPct > 100) {
		return NewValidationError("city_driving_pct", "city driving percentage must be between 0 and 100")
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
