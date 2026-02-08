package repositories

import (
	"context"

	"github.com/truggeri/go-garage/internal/models"
)

// VehicleFilters contains optional filters for listing vehicles
type VehicleFilters struct {
	Status *models.VehicleStatus
	Make   *string
	Model  *string
	Year   *int
}

// PaginationParams contains parameters for pagination
type PaginationParams struct {
	Limit  int
	Offset int
}

// VehicleRepository defines the interface for vehicle data access
type VehicleRepository interface {
	// Create inserts a new vehicle into the database
	Create(ctx context.Context, vehicle *models.Vehicle) error

	// FindByID retrieves a vehicle by its ID
	FindByID(ctx context.Context, id string) (*models.Vehicle, error)

	// FindByUserID retrieves all vehicles for a specific user
	FindByUserID(ctx context.Context, userID string) ([]*models.Vehicle, error)

	// FindByVIN retrieves a vehicle by its VIN
	FindByVIN(ctx context.Context, vin string) (*models.Vehicle, error)

	// Update modifies an existing vehicle's information
	Update(ctx context.Context, vehicle *models.Vehicle) error

	// Delete removes a vehicle from the database
	Delete(ctx context.Context, id string) error

	// List retrieves vehicles with optional filters and pagination
	List(ctx context.Context, filters VehicleFilters, pagination PaginationParams) ([]*models.Vehicle, error)
}
