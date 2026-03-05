package repositories

import (
	"context"

	"github.com/truggeri/go-garage/internal/models"
)

// MetricsRepository defines the interface for vehicle metrics data access.
type MetricsRepository interface {
	// Upsert inserts or updates the metrics for a given vehicle.
	Upsert(ctx context.Context, metrics *models.VehicleMetrics) error

	// GetByVehicleID retrieves the metrics for a specific vehicle.
	// Returns nil without error if no metrics row exists.
	GetByVehicleID(ctx context.Context, vehicleID string) (*models.VehicleMetrics, error)

	// SumTotalSpentByVehicleIDs returns the sum of total_spent across the given vehicle IDs.
	SumTotalSpentByVehicleIDs(ctx context.Context, vehicleIDs []string) (float64, error)
}
