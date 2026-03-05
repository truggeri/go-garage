package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateVehicle(t *testing.T) {
	purchasePrice := 25000.0
	negativePurchasePrice := -1000.0
	purchaseMileage := 10000
	negativeMileage := -100
	currentMileage := 15000
	lowerCurrentMileage := 5000

	tests := []struct {
		vehicle     *Vehicle
		name        string
		expectError bool
		errorField  string
	}{
		{
			name: "valid vehicle",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "1HGBH41JXMN109186",
				Make:   "Honda",
				Model:  "Civic",
				Year:   2020,
				Status: VehicleStatusActive,
			},
			expectError: false,
		},
		{
			name: "empty user ID",
			vehicle: &Vehicle{
				UserID: "",
				VIN:    "1HGBH41JXMN109186",
				Make:   "Honda",
				Model:  "Civic",
				Year:   2020,
				Status: VehicleStatusActive,
			},
			expectError: true,
			errorField:  "user_id",
		},
		{
			name: "empty VIN",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "",
				Make:   "Honda",
				Model:  "Civic",
				Year:   2020,
				Status: VehicleStatusActive,
			},
			expectError: true,
			errorField:  "vin",
		},
		{
			name: "VIN too short",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "1HGBH41JX",
				Make:   "Honda",
				Model:  "Civic",
				Year:   2020,
				Status: VehicleStatusActive,
			},
			expectError: true,
			errorField:  "vin",
		},
		{
			name: "VIN with invalid characters",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "1HGBH41JXMN10918I", // Contains 'I'
				Make:   "Honda",
				Model:  "Civic",
				Year:   2020,
				Status: VehicleStatusActive,
			},
			expectError: true,
			errorField:  "vin",
		},
		{
			name: "empty make",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "1HGBH41JXMN109186",
				Make:   "",
				Model:  "Civic",
				Year:   2020,
				Status: VehicleStatusActive,
			},
			expectError: true,
			errorField:  "make",
		},
		{
			name: "empty model",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "1HGBH41JXMN109186",
				Make:   "Honda",
				Model:  "",
				Year:   2020,
				Status: VehicleStatusActive,
			},
			expectError: true,
			errorField:  "model",
		},
		{
			name: "zero year",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "1HGBH41JXMN109186",
				Make:   "Honda",
				Model:  "Civic",
				Year:   0,
				Status: VehicleStatusActive,
			},
			expectError: true,
			errorField:  "year",
		},
		{
			name: "year too old",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "1HGBH41JXMN109186",
				Make:   "Honda",
				Model:  "Civic",
				Year:   1899,
				Status: VehicleStatusActive,
			},
			expectError: true,
			errorField:  "year",
		},
		{
			name: "year too far in future",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "1HGBH41JXMN109186",
				Make:   "Honda",
				Model:  "Civic",
				Year:   time.Now().Year() + 2,
				Status: VehicleStatusActive,
			},
			expectError: true,
			errorField:  "year",
		},
		{
			name: "empty status",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "1HGBH41JXMN109186",
				Make:   "Honda",
				Model:  "Civic",
				Year:   2020,
				Status: "",
			},
			expectError: true,
			errorField:  "status",
		},
		{
			name: "invalid status",
			vehicle: &Vehicle{
				UserID: "user123",
				VIN:    "1HGBH41JXMN109186",
				Make:   "Honda",
				Model:  "Civic",
				Year:   2020,
				Status: "unknown",
			},
			expectError: true,
			errorField:  "status",
		},
		{
			name: "negative purchase price",
			vehicle: &Vehicle{
				UserID:        "user123",
				VIN:           "1HGBH41JXMN109186",
				Make:          "Honda",
				Model:         "Civic",
				Year:          2020,
				Status:        VehicleStatusActive,
				PurchasePrice: &negativePurchasePrice,
			},
			expectError: true,
			errorField:  "purchase_price",
		},
		{
			name: "negative purchase mileage",
			vehicle: &Vehicle{
				UserID:          "user123",
				VIN:             "1HGBH41JXMN109186",
				Make:            "Honda",
				Model:           "Civic",
				Year:            2020,
				Status:          VehicleStatusActive,
				PurchaseMileage: &negativeMileage,
			},
			expectError: true,
			errorField:  "purchase_mileage",
		},
		{
			name: "negative current mileage",
			vehicle: &Vehicle{
				UserID:         "user123",
				VIN:            "1HGBH41JXMN109186",
				Make:           "Honda",
				Model:          "Civic",
				Year:           2020,
				Status:         VehicleStatusActive,
				CurrentMileage: &negativeMileage,
			},
			expectError: true,
			errorField:  "current_mileage",
		},
		{
			name: "current mileage less than purchase mileage",
			vehicle: &Vehicle{
				UserID:          "user123",
				VIN:             "1HGBH41JXMN109186",
				Make:            "Honda",
				Model:           "Civic",
				Year:            2020,
				Status:          VehicleStatusActive,
				PurchaseMileage: &purchaseMileage,
				CurrentMileage:  &lowerCurrentMileage,
			},
			expectError: true,
			errorField:  "current_mileage",
		},
		{
			name: "valid with all optional fields",
			vehicle: &Vehicle{
				UserID:          "user123",
				VIN:             "1HGBH41JXMN109186",
				Make:            "Honda",
				Model:           "Civic",
				Year:            2020,
				Status:          VehicleStatusActive,
				PurchasePrice:   &purchasePrice,
				PurchaseMileage: &purchaseMileage,
				CurrentMileage:  &currentMileage,
				Color:           "Blue",
				LicensePlate:    "ABC123",
				Notes:           "Great car",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVehicle(tt.vehicle)
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

func TestValidateVehicleAll(t *testing.T) {
	t.Run("returns empty map for valid vehicle", func(t *testing.T) {
		v := &Vehicle{
			UserID: "user123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2020,
			Status: VehicleStatusActive,
		}
		errs := ValidateVehicleAll(v)
		assert.Empty(t, errs)
	})

	t.Run("returns all errors for completely invalid vehicle", func(t *testing.T) {
		v := &Vehicle{}
		errs := ValidateVehicleAll(v)
		assert.Contains(t, errs, "user_id")
		assert.Contains(t, errs, "vin")
		assert.Contains(t, errs, "make")
		assert.Contains(t, errs, "model")
		assert.Contains(t, errs, "year")
		assert.Contains(t, errs, "status")
	})

	t.Run("returns multiple field errors simultaneously", func(t *testing.T) {
		v := &Vehicle{
			UserID: "user123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "",
			Model:  "",
			Year:   2020,
			Status: VehicleStatusActive,
		}
		errs := ValidateVehicleAll(v)
		assert.Len(t, errs, 2)
		assert.Contains(t, errs, "make")
		assert.Contains(t, errs, "model")
	})

	t.Run("returns numeric validation errors", func(t *testing.T) {
		negPrice := -100.0
		negMileage := -50
		v := &Vehicle{
			UserID:          "user123",
			VIN:             "1HGBH41JXMN109186",
			Make:            "Honda",
			Model:           "Civic",
			Year:            2020,
			Status:          VehicleStatusActive,
			PurchasePrice:   &negPrice,
			PurchaseMileage: &negMileage,
		}
		errs := ValidateVehicleAll(v)
		assert.Contains(t, errs, "purchase_price")
		assert.Contains(t, errs, "purchase_mileage")
	})
}

func TestValidateMaintenanceRecord(t *testing.T) {
	cost := 150.0
	negativeCost := -50.0
	mileage := 15000
	negativeMileage := -100

	tests := []struct {
		name        string
		record      *MaintenanceRecord
		expectError bool
		errorField  string
	}{
		{
			name: "valid record",
			record: &MaintenanceRecord{
				VehicleID:   "vehicle123",
				ServiceType: ServiceTypeOilChange,
				ServiceDate: time.Now().Add(-24 * time.Hour),
			},
			expectError: false,
		},
		{
			name: "empty vehicle ID",
			record: &MaintenanceRecord{
				VehicleID:   "",
				ServiceType: ServiceTypeOilChange,
				ServiceDate: time.Now().Add(-24 * time.Hour),
			},
			expectError: true,
			errorField:  "vehicle_id",
		},
		{
			name: "empty service type",
			record: &MaintenanceRecord{
				VehicleID:   "vehicle123",
				ServiceType: "",
				ServiceDate: time.Now().Add(-24 * time.Hour),
			},
			expectError: true,
			errorField:  "service_type",
		},
		{
			name: "invalid service type",
			record: &MaintenanceRecord{
				VehicleID:   "vehicle123",
				ServiceType: "invalid_type",
				ServiceDate: time.Now().Add(-24 * time.Hour),
			},
			expectError: true,
			errorField:  "service_type",
		},
		{
			name: "other service type without custom value",
			record: &MaintenanceRecord{
				VehicleID:   "vehicle123",
				ServiceType: ServiceTypeOther,
				ServiceDate: time.Now().Add(-24 * time.Hour),
			},
			expectError: true,
			errorField:  "custom_service_type",
		},
		{
			name: "other service type with custom value",
			record: &MaintenanceRecord{
				VehicleID:         "vehicle123",
				ServiceType:       ServiceTypeOther,
				CustomServiceType: "Custom repair",
				ServiceDate:       time.Now().Add(-24 * time.Hour),
			},
			expectError: false,
		},
		{
			name: "zero service date",
			record: &MaintenanceRecord{
				VehicleID:   "vehicle123",
				ServiceType: ServiceTypeOilChange,
				ServiceDate: time.Time{},
			},
			expectError: true,
			errorField:  "service_date",
		},
		{
			name: "future service date",
			record: &MaintenanceRecord{
				VehicleID:   "vehicle123",
				ServiceType: ServiceTypeOilChange,
				ServiceDate: time.Now().Add(24 * time.Hour),
			},
			expectError: true,
			errorField:  "service_date",
		},
		{
			name: "negative cost",
			record: &MaintenanceRecord{
				VehicleID:   "vehicle123",
				ServiceType: ServiceTypeOilChange,
				ServiceDate: time.Now().Add(-24 * time.Hour),
				Cost:        &negativeCost,
			},
			expectError: true,
			errorField:  "cost",
		},
		{
			name: "negative mileage",
			record: &MaintenanceRecord{
				VehicleID:        "vehicle123",
				ServiceType:      ServiceTypeOilChange,
				ServiceDate:      time.Now().Add(-24 * time.Hour),
				MileageAtService: &negativeMileage,
			},
			expectError: true,
			errorField:  "mileage_at_service",
		},
		{
			name: "valid with all fields",
			record: &MaintenanceRecord{
				VehicleID:        "vehicle123",
				ServiceType:      ServiceTypeOilChange,
				ServiceDate:      time.Now().Add(-24 * time.Hour),
				Cost:             &cost,
				MileageAtService: &mileage,
				ServiceProvider:  "Quick Lube",
				Notes:            "Synthetic oil",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaintenanceRecord(tt.record)
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
