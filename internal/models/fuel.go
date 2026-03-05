package models

import (
	"time"
)

// FuelRecord represents a fuel fill-up record for a vehicle
type FuelRecord struct {
	ID             string
	VehicleID      string
	FillDate       time.Time
	Odometer       int
	CostPerUnit    float64
	Volume         float64
	FuelType       string
	CityDrivingPct *int
	Location       string
	Brand          string
	Notes          string
	ReportedMPG    *float64
	PartialFuelUp  bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
