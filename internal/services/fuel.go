package services

import (
	"context"
	"time"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
)

// FuelService defines the interface for fuel record business logic
type FuelService interface {
	// CreateFuelRecord creates a new fuel record
	CreateFuelRecord(ctx context.Context, record *models.FuelRecord) error

	// GetFuelRecord retrieves a fuel record by its ID
	GetFuelRecord(ctx context.Context, id string) (*models.FuelRecord, error)

	// GetVehicleFuelRecords retrieves all fuel records for a specific vehicle
	GetVehicleFuelRecords(ctx context.Context, vehicleID string) ([]*models.FuelRecord, error)

	// ListFuelRecords retrieves fuel records with optional filters and pagination
	ListFuelRecords(ctx context.Context, filters repositories.FuelFilters, pagination repositories.PaginationParams) ([]*models.FuelRecord, error)

	// CountFuelRecords returns the total number of fuel records matching the filters
	CountFuelRecords(ctx context.Context, filters repositories.FuelFilters) (int, error)

	// UpdateFuelRecord updates a fuel record's information
	UpdateFuelRecord(ctx context.Context, id string, updates FuelUpdates) (*models.FuelRecord, error)

	// DeleteFuelRecord deletes a fuel record
	DeleteFuelRecord(ctx context.Context, id string) error
}

// FuelUpdates contains the fields that can be updated for a fuel record
type FuelUpdates struct {
	FillDate       *time.Time
	Odometer       *int
	CostPerUnit    *float64
	Volume         *float64
	FuelType       *string
	CityDrivingPct *int
	Location       *string
	Brand          *string
	Notes          *string
	ReportedMPG    *float64
	PartialFuelUp  *bool
}

// DefaultFuelService implements FuelService using a FuelRepository
type DefaultFuelService struct {
	repo        repositories.FuelRepository
	vehicleRepo repositories.VehicleRepository
}

// NewFuelService creates a new DefaultFuelService
func NewFuelService(repo repositories.FuelRepository, vehicleRepo repositories.VehicleRepository) *DefaultFuelService {
	return &DefaultFuelService{
		repo:        repo,
		vehicleRepo: vehicleRepo,
	}
}

// CreateFuelRecord creates a new fuel record
func (s *DefaultFuelService) CreateFuelRecord(ctx context.Context, record *models.FuelRecord) error {
	// Verify the vehicle exists
	_, err := s.vehicleRepo.FindByID(ctx, record.VehicleID)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, record)
}

// GetFuelRecord retrieves a fuel record by its ID
func (s *DefaultFuelService) GetFuelRecord(ctx context.Context, id string) (*models.FuelRecord, error) {
	return s.repo.FindByID(ctx, id)
}

// GetVehicleFuelRecords retrieves all fuel records for a specific vehicle
func (s *DefaultFuelService) GetVehicleFuelRecords(ctx context.Context, vehicleID string) ([]*models.FuelRecord, error) {
	return s.repo.FindByVehicleID(ctx, vehicleID)
}

// ListFuelRecords retrieves fuel records with optional filters and pagination
func (s *DefaultFuelService) ListFuelRecords(ctx context.Context, filters repositories.FuelFilters, pagination repositories.PaginationParams) ([]*models.FuelRecord, error) {
	return s.repo.List(ctx, filters, pagination)
}

// CountFuelRecords returns the total number of fuel records matching the filters
func (s *DefaultFuelService) CountFuelRecords(ctx context.Context, filters repositories.FuelFilters) (int, error) {
	return s.repo.Count(ctx, filters)
}

// UpdateFuelRecord updates a fuel record's information
func (s *DefaultFuelService) UpdateFuelRecord(ctx context.Context, id string, updates FuelUpdates) (*models.FuelRecord, error) {
	// Get the existing record
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.FillDate != nil {
		record.FillDate = *updates.FillDate
	}
	if updates.Odometer != nil {
		record.Odometer = *updates.Odometer
	}
	if updates.CostPerUnit != nil {
		record.CostPerUnit = *updates.CostPerUnit
	}
	if updates.Volume != nil {
		record.Volume = *updates.Volume
	}
	if updates.FuelType != nil {
		record.FuelType = *updates.FuelType
	}
	if updates.CityDrivingPct != nil {
		record.CityDrivingPct = updates.CityDrivingPct
	}
	if updates.Location != nil {
		record.Location = *updates.Location
	}
	if updates.Brand != nil {
		record.Brand = *updates.Brand
	}
	if updates.Notes != nil {
		record.Notes = *updates.Notes
	}
	if updates.ReportedMPG != nil {
		record.ReportedMPG = updates.ReportedMPG
	}
	if updates.PartialFuelUp != nil {
		record.PartialFuelUp = *updates.PartialFuelUp
	}

	// Update the record
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// DeleteFuelRecord deletes a fuel record
func (s *DefaultFuelService) DeleteFuelRecord(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
