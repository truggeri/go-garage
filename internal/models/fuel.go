package models

import (
	"time"
)

// FuelType represents the type of fuel used for a fill-up.
type FuelType string

const (
	FuelTypeGasoline FuelType = "gasoline"
	FuelTypeDiesel   FuelType = "diesel"
	FuelTypeE85      FuelType = "e85"
)

// fuelTypeDisplayNames maps enum values to human-readable display names.
var fuelTypeDisplayNames = map[FuelType]string{
	FuelTypeGasoline: "Gasoline",
	FuelTypeDiesel:   "Diesel",
	FuelTypeE85:      "E85",
}

// allFuelTypes defines the canonical order for dropdowns and iteration.
var allFuelTypes = []FuelType{
	FuelTypeGasoline,
	FuelTypeDiesel,
	FuelTypeE85,
}

// IsValidFuelType returns true if the given string is a recognized fuel type.
func IsValidFuelType(s string) bool {
	_, ok := fuelTypeDisplayNames[FuelType(s)]
	return ok
}

// FuelTypeDisplayName returns the human-readable display name for a fuel type enum value.
// It returns the raw value if the fuel type is not recognized.
func FuelTypeDisplayName(s string) string {
	if name, ok := fuelTypeDisplayNames[FuelType(s)]; ok {
		return name
	}
	return s
}

// AllFuelTypes returns all valid fuel types in display order.
func AllFuelTypes() []FuelType {
	result := make([]FuelType, len(allFuelTypes))
	copy(result, allFuelTypes)
	return result
}

// FuelRecord represents a fuel fill-up record for a vehicle.
type FuelRecord struct {
	ID                    string
	VehicleID             string
	FillDate              time.Time
	Mileage               int
	Volume                float64
	FuelType              string
	PartialFill           bool
	PricePerUnit          *float64
	OctaneRating          *int
	Location              string
	Brand                 string
	Notes                 string
	CityDrivingPercentage *int
	VehicleReportedMPG    *float64
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
