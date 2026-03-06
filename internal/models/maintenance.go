package models

import (
	"time"
)

// ServiceType represents the type of maintenance service performed
type ServiceType string

const (
	// ServiceTypeOilChange indicates an oil change service
	ServiceTypeOilChange ServiceType = "oil_change"
	// ServiceTypeTireRotation indicates a tire rotation service
	ServiceTypeTireRotation ServiceType = "tire_rotation"
	// ServiceTypeAirFilter indicates an air filter replacement
	ServiceTypeAirFilter ServiceType = "air_filter"
	// ServiceTypeCabinAirFilter indicates a cabin air filter replacement
	ServiceTypeCabinAirFilter ServiceType = "cabin_air_filter"
	// ServiceTypeFuelAdditive indicates a fuel additive service
	ServiceTypeFuelAdditive ServiceType = "fuel_additive"
	// ServiceTypeBattery indicates a battery service
	ServiceTypeBattery ServiceType = "battery"
	// ServiceTypeBrakes indicates a brake service
	ServiceTypeBrakes ServiceType = "brakes"
	// ServiceTypeBrakeFluid indicates a brake fluid service
	ServiceTypeBrakeFluid ServiceType = "brake_fluid"
	// ServiceTypeRadiatorFluid indicates a radiator fluid service
	ServiceTypeRadiatorFluid ServiceType = "radiator_fluid"
	// ServiceTypeTires indicates a tire replacement service
	ServiceTypeTires ServiceType = "tires"
	// ServiceTypeGlass indicates a glass repair or replacement service
	ServiceTypeGlass ServiceType = "glass"
	// ServiceTypeBodyWork indicates a body work service
	ServiceTypeBodyWork ServiceType = "body_work"
	// ServiceTypeInterior indicates an interior service
	ServiceTypeInterior ServiceType = "interior"
	// ServiceTypeOther indicates a custom service type
	ServiceTypeOther ServiceType = "other"
)

// serviceTypeDisplayNames maps service type enum values to human-readable display names.
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

// allServiceTypes is the ordered list of all valid service types.
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

// IsValidServiceType returns true if the given service type is a valid enum value.
func IsValidServiceType(st ServiceType) bool {
	_, ok := serviceTypeDisplayNames[st]
	return ok
}

// ServiceTypeDisplayName returns the human-readable display name for a service type.
func ServiceTypeDisplayName(st ServiceType) string {
	if name, ok := serviceTypeDisplayNames[st]; ok {
		return name
	}
	return string(st)
}

// AllServiceTypes returns an ordered slice of all valid service types.
func AllServiceTypes() []ServiceType {
	result := make([]ServiceType, len(allServiceTypes))
	copy(result, allServiceTypes)
	return result
}

// MaintenanceRecord represents a maintenance or service record for a vehicle
type MaintenanceRecord struct {
	ID                string
	VehicleID         string
	ServiceType       ServiceType
	CustomServiceType string
	ServiceDate       time.Time
	MileageAtService  *int
	Cost              *float64
	ServiceProvider   string
	Notes             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// DisplayServiceType returns the display-friendly service type string.
// For "Other" service types, it returns the custom service type value.
func (m *MaintenanceRecord) DisplayServiceType() string {
	if m.ServiceType == ServiceTypeOther && m.CustomServiceType != "" {
		return m.CustomServiceType
	}
	return ServiceTypeDisplayName(m.ServiceType)
}
