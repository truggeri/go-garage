package services

import (
	"context"
	"time"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
)

// MaintenanceService defines the interface for maintenance record business logic
type MaintenanceService interface {
	// CreateMaintenance creates a new maintenance record
	CreateMaintenance(ctx context.Context, record *models.MaintenanceRecord) error

	// GetMaintenance retrieves a maintenance record by its ID
	GetMaintenance(ctx context.Context, id string) (*models.MaintenanceRecord, error)

	// GetVehicleMaintenance retrieves all maintenance records for a specific vehicle
	GetVehicleMaintenance(ctx context.Context, vehicleID string) ([]*models.MaintenanceRecord, error)

	// ListMaintenance retrieves maintenance records with optional filters and pagination
	ListMaintenance(ctx context.Context, filters repositories.MaintenanceFilters, pagination repositories.PaginationParams) ([]*models.MaintenanceRecord, error)

	// CountMaintenance returns the total number of maintenance records matching the filters
	CountMaintenance(ctx context.Context, filters repositories.MaintenanceFilters) (int, error)

	// UpdateMaintenance updates a maintenance record's information
	UpdateMaintenance(ctx context.Context, id string, updates MaintenanceUpdates) (*models.MaintenanceRecord, error)

	// DeleteMaintenance deletes a maintenance record
	DeleteMaintenance(ctx context.Context, id string) error
}

// MaintenanceUpdates contains the fields that can be updated for a maintenance record
type MaintenanceUpdates struct {
	ServiceType      *string
	CustomServiceType *string
	ServiceDate      *time.Time
	MileageAtService *int
	Cost             *float64
	ServiceProvider  *string
	Notes            *string
}

// DefaultMaintenanceService implements MaintenanceService using a MaintenanceRepository
type DefaultMaintenanceService struct {
	repo        repositories.MaintenanceRepository
	vehicleRepo repositories.VehicleRepository
}

// NewMaintenanceService creates a new DefaultMaintenanceService
func NewMaintenanceService(repo repositories.MaintenanceRepository, vehicleRepo repositories.VehicleRepository) *DefaultMaintenanceService {
	return &DefaultMaintenanceService{
		repo:        repo,
		vehicleRepo: vehicleRepo,
	}
}

// CreateMaintenance creates a new maintenance record
func (s *DefaultMaintenanceService) CreateMaintenance(ctx context.Context, record *models.MaintenanceRecord) error {
	// Verify the vehicle exists
	_, err := s.vehicleRepo.FindByID(ctx, record.VehicleID)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, record)
}

// GetMaintenance retrieves a maintenance record by its ID
func (s *DefaultMaintenanceService) GetMaintenance(ctx context.Context, id string) (*models.MaintenanceRecord, error) {
	return s.repo.FindByID(ctx, id)
}

// GetVehicleMaintenance retrieves all maintenance records for a specific vehicle
func (s *DefaultMaintenanceService) GetVehicleMaintenance(ctx context.Context, vehicleID string) ([]*models.MaintenanceRecord, error) {
	return s.repo.FindByVehicleID(ctx, vehicleID)
}

// ListMaintenance retrieves maintenance records with optional filters and pagination
func (s *DefaultMaintenanceService) ListMaintenance(ctx context.Context, filters repositories.MaintenanceFilters, pagination repositories.PaginationParams) ([]*models.MaintenanceRecord, error) {
	return s.repo.List(ctx, filters, pagination)
}

// CountMaintenance returns the total number of maintenance records matching the filters
func (s *DefaultMaintenanceService) CountMaintenance(ctx context.Context, filters repositories.MaintenanceFilters) (int, error) {
	return s.repo.Count(ctx, filters)
}

// UpdateMaintenance updates a maintenance record's information
func (s *DefaultMaintenanceService) UpdateMaintenance(ctx context.Context, id string, updates MaintenanceUpdates) (*models.MaintenanceRecord, error) {
	// Get the existing record
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.ServiceType != nil {
		record.ServiceType = *updates.ServiceType
	}
	if updates.CustomServiceType != nil {
		record.CustomServiceType = *updates.CustomServiceType
	}
	if updates.ServiceDate != nil {
		record.ServiceDate = *updates.ServiceDate
	}
	if updates.MileageAtService != nil {
		record.MileageAtService = updates.MileageAtService
	}
	if updates.Cost != nil {
		record.Cost = updates.Cost
	}
	if updates.ServiceProvider != nil {
		record.ServiceProvider = *updates.ServiceProvider
	}
	if updates.Notes != nil {
		record.Notes = *updates.Notes
	}

	// Update the record
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// DeleteMaintenance deletes a maintenance record
func (s *DefaultMaintenanceService) DeleteMaintenance(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
