package repositories

import (
	"context"

	"github.com/truggeri/go-garage/internal/models"
)

// FuelFilters contains optional filters for listing fuel records
type FuelFilters struct {
	VehicleID *string
	FuelType  *string
}

// FuelRepository defines the interface for fuel record data access
type FuelRepository interface {
	// Create inserts a new fuel record into the database
	Create(ctx context.Context, record *models.FuelRecord) error

	// FindByID retrieves a fuel record by its ID
	FindByID(ctx context.Context, id string) (*models.FuelRecord, error)

	// FindByVehicleID retrieves all fuel records for a specific vehicle
	FindByVehicleID(ctx context.Context, vehicleID string) ([]*models.FuelRecord, error)

	// Update modifies an existing fuel record
	Update(ctx context.Context, record *models.FuelRecord) error

	// Delete removes a fuel record from the database
	Delete(ctx context.Context, id string) error

	// List retrieves fuel records with optional filters and pagination
	List(ctx context.Context, filters FuelFilters, pagination PaginationParams) ([]*models.FuelRecord, error)

	// Count returns the total number of fuel records matching the filters
	Count(ctx context.Context, filters FuelFilters) (int, error)

	// SumCostByVehicleID returns the total fuel cost (price_per_unit * volume) for a specific vehicle.
	// Returns nil if there are no records with a price for the vehicle.
	SumCostByVehicleID(ctx context.Context, vehicleID string) (*float64, error)
}
