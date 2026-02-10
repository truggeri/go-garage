package models

import (
	"time"
)

// MaintenanceRecord represents a maintenance or service record for a vehicle
type MaintenanceRecord struct {
	ID               string
	VehicleID        string
	ServiceType      string
	ServiceDate      time.Time
	MileageAtService *int
	Cost             *float64
	ServiceProvider  string
	Notes            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
