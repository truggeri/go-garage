package repositories

import (
	"context"

	"github.com/truggeri/go-garage/internal/models"
)

// MaintenanceFilters contains optional filters for listing maintenance records
type MaintenanceFilters struct {
	ServiceType *string
}

// MaintenanceRepository defines the interface for maintenance record data access
type MaintenanceRepository interface {
	// Create inserts a new maintenance record into the database
	Create(ctx context.Context, record *models.MaintenanceRecord) error

	// FindByID retrieves a maintenance record by its ID
	FindByID(ctx context.Context, id string) (*models.MaintenanceRecord, error)

	// FindByVehicleID retrieves all maintenance records for a specific vehicle
	FindByVehicleID(ctx context.Context, vehicleID string) ([]*models.MaintenanceRecord, error)

	// Update modifies an existing maintenance record
	Update(ctx context.Context, record *models.MaintenanceRecord) error

	// Delete removes a maintenance record from the database
	Delete(ctx context.Context, id string) error

	// List retrieves maintenance records with optional filters and pagination
	List(ctx context.Context, filters MaintenanceFilters, pagination PaginationParams) ([]*models.MaintenanceRecord, error)
}
