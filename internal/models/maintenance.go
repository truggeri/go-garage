package models

import (
	"time"
)

// ServiceType represents the type of maintenance service performed.
type ServiceType string

const (
	ServiceTypeOilChange      ServiceType = "oil_change"
	ServiceTypeTireRotation   ServiceType = "tire_rotation"
	ServiceTypeAirFilter      ServiceType = "air_filter"
	ServiceTypeCabinAirFilter ServiceType = "cabin_air_filter"
	ServiceTypeFuelAdditive   ServiceType = "fuel_additive"
	ServiceTypeBattery        ServiceType = "battery"
	ServiceTypeBrakes         ServiceType = "brakes"
	ServiceTypeBrakeFluid     ServiceType = "brake_fluid"
	ServiceTypeRadiatorFluid  ServiceType = "radiator_fluid"
	ServiceTypeTires          ServiceType = "tires"
	ServiceTypeGlass          ServiceType = "glass"
	ServiceTypeBodyWork       ServiceType = "body_work"
	ServiceTypeInterior       ServiceType = "interior"
	ServiceTypeOther          ServiceType = "other"
)

// serviceTypeDisplayNames maps enum values to human-readable display names.
var serviceTypeDisplayNames = map[ServiceType]string{
	ServiceTypeOilChange:      "Oil Change",
	ServiceTypeTireRotation:   "Tire Rotation",
	ServiceTypeAirFilter:      "Air Filter",
	ServiceTypeCabinAirFilter: "Cabin Air Filter",
	ServiceTypeFuelAdditive:   "Fuel Additive",
	ServiceTypeBattery:        "Battery",
	ServiceTypeBrakes:         "Brakes",
	ServiceTypeBrakeFluid:     "Brake Fluid",
	ServiceTypeRadiatorFluid:  "Radiator Fluid",
	ServiceTypeTires:          "Tires",
	ServiceTypeGlass:          "Glass",
	ServiceTypeBodyWork:       "Body Work",
	ServiceTypeInterior:       "Interior",
	ServiceTypeOther:          "Other",
}

// allServiceTypes defines the canonical order for dropdowns and iteration.
var allServiceTypes = []ServiceType{
	ServiceTypeOilChange,
	ServiceTypeTireRotation,
	ServiceTypeAirFilter,
	ServiceTypeCabinAirFilter,
	ServiceTypeFuelAdditive,
	ServiceTypeBattery,
	ServiceTypeBrakes,
	ServiceTypeBrakeFluid,
	ServiceTypeRadiatorFluid,
	ServiceTypeTires,
	ServiceTypeGlass,
	ServiceTypeBodyWork,
	ServiceTypeInterior,
	ServiceTypeOther,
}

// IsValidServiceType returns true if the given string is a recognized service type.
func IsValidServiceType(s string) bool {
	_, ok := serviceTypeDisplayNames[ServiceType(s)]
	return ok
}

// ServiceTypeDisplayName returns the human-readable display name for a service type enum value.
// It returns the raw value if the service type is not recognized.
func ServiceTypeDisplayName(s string) string {
	if name, ok := serviceTypeDisplayNames[ServiceType(s)]; ok {
		return name
	}
	return s
}

// AllServiceTypes returns all valid service types in display order.
func AllServiceTypes() []ServiceType {
	result := make([]ServiceType, len(allServiceTypes))
	copy(result, allServiceTypes)
	return result
}

// MaintenanceRecord represents a maintenance or service record for a vehicle
type MaintenanceRecord struct {
	ID                string
	VehicleID         string
	ServiceType       string
	CustomServiceType string
	ServiceDate       time.Time
	MileageAtService  *int
	Cost              *float64
	ServiceProvider   string
	Notes             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
