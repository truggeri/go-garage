package models

import (
	"testing"
	"time"

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

func validFuelRecord() *FuelRecord {
	return &FuelRecord{
		VehicleID: "vehicle-123",
		FillDate:  time.Now().Add(-24 * time.Hour),
		Mileage:   50000,
		Volume:    12.5,
		FuelType:  "gasoline",
	}
}

func TestValidateFuelRecord(t *testing.T) {
	negativeFloat := -1.0
	zeroFloat := 0.0
	positiveFloat := 3.50
	negativeInt := -5
	validPercentage := 50
	tooHighPercentage := 101
	validOctane := 87
	zeroOctane := 0
	validMPG := 32.5
	negativeMPG := -1.0

	tests := []struct {
		name        string
		modify      func(f *FuelRecord)
		expectError bool
		errorField  string
	}{
		{
			name:        "valid fuel record",
			modify:      func(f *FuelRecord) {},
			expectError: false,
		},
		{
			name: "valid with all optional fields",
			modify: func(f *FuelRecord) {
				f.PartialFill = true
				f.PricePerUnit = &positiveFloat
				f.OctaneRating = &validOctane
				f.Location = "Shell Station"
				f.Brand = "Shell"
				f.Notes = "Premium fuel"
				f.CityDrivingPercentage = &validPercentage
				f.VehicleReportedMPG = &validMPG
			},
			expectError: false,
		},
		{
			name:        "missing vehicle_id",
			modify:      func(f *FuelRecord) { f.VehicleID = "" },
			expectError: true,
			errorField:  "vehicle_id",
		},
		{
			name:        "missing fill_date",
			modify:      func(f *FuelRecord) { f.FillDate = time.Time{} },
			expectError: true,
			errorField:  "fill_date",
		},
		{
			name:        "future fill_date",
			modify:      func(f *FuelRecord) { f.FillDate = time.Now().Add(48 * time.Hour) },
			expectError: true,
			errorField:  "fill_date",
		},
		{
			name:        "zero mileage",
			modify:      func(f *FuelRecord) { f.Mileage = 0 },
			expectError: true,
			errorField:  "mileage",
		},
		{
			name:        "negative mileage",
			modify:      func(f *FuelRecord) { f.Mileage = -100 },
			expectError: true,
			errorField:  "mileage",
		},
		{
			name:        "zero volume",
			modify:      func(f *FuelRecord) { f.Volume = 0 },
			expectError: true,
			errorField:  "volume",
		},
		{
			name:        "negative volume",
			modify:      func(f *FuelRecord) { f.Volume = -5.0 },
			expectError: true,
			errorField:  "volume",
		},
		{
			name:        "empty fuel_type",
			modify:      func(f *FuelRecord) { f.FuelType = "" },
			expectError: true,
			errorField:  "fuel_type",
		},
		{
			name:        "invalid fuel_type",
			modify:      func(f *FuelRecord) { f.FuelType = "propane" },
			expectError: true,
			errorField:  "fuel_type",
		},
		{
			name:        "negative price_per_unit",
			modify:      func(f *FuelRecord) { f.PricePerUnit = &negativeFloat },
			expectError: true,
			errorField:  "price_per_unit",
		},
		{
			name:        "zero price_per_unit is valid",
			modify:      func(f *FuelRecord) { f.PricePerUnit = &zeroFloat },
			expectError: false,
		},
		{
			name:        "zero octane_rating",
			modify:      func(f *FuelRecord) { f.OctaneRating = &zeroOctane },
			expectError: true,
			errorField:  "octane_rating",
		},
		{
			name:        "negative city_driving_percentage",
			modify:      func(f *FuelRecord) { f.CityDrivingPercentage = &negativeInt },
			expectError: true,
			errorField:  "city_driving_percentage",
		},
		{
			name:        "city_driving_percentage over 100",
			modify:      func(f *FuelRecord) { f.CityDrivingPercentage = &tooHighPercentage },
			expectError: true,
			errorField:  "city_driving_percentage",
		},
		{
			name:        "negative vehicle_reported_mpg",
			modify:      func(f *FuelRecord) { f.VehicleReportedMPG = &negativeMPG },
			expectError: true,
			errorField:  "vehicle_reported_mpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := validFuelRecord()
			tt.modify(f)
			err := ValidateFuelRecord(f)
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
