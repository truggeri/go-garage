package models

import "time"

// VehicleMetrics holds aggregated metrics for a vehicle.
type VehicleMetrics struct {
	// VehicleID is the primary key and foreign key to the vehicle.
	VehicleID string
	// TotalSpent is the sum of all maintenance costs for this vehicle. Nil means no costs recorded.
	TotalSpent *float64
	// CreatedAt is the timestamp when this record was created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when this record was last updated.
	UpdatedAt time.Time
}
