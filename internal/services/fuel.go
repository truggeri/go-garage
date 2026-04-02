package services

import (
	"context"
	"time"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
)

// FuelService defines the interface for fuel record business logic
type FuelService interface {
	// CreateFuel creates a new fuel record
	CreateFuel(ctx context.Context, record *models.FuelRecord) error

	// GetFuel retrieves a fuel record by its ID
	GetFuel(ctx context.Context, id string) (*models.FuelRecord, error)

	// GetVehicleFuel retrieves all fuel records for a specific vehicle
	GetVehicleFuel(ctx context.Context, vehicleID string) ([]*models.FuelRecord, error)

	// ListFuel retrieves fuel records with optional filters and pagination
	ListFuel(ctx context.Context, filters repositories.FuelFilters, pagination repositories.PaginationParams) ([]*models.FuelRecord, error)

	// CountFuel returns the total number of fuel records matching the filters
	CountFuel(ctx context.Context, filters repositories.FuelFilters) (int, error)

	// UpdateFuel updates a fuel record
	UpdateFuel(ctx context.Context, id string, updates FuelUpdates) (*models.FuelRecord, error)

	// DeleteFuel deletes a fuel record
	DeleteFuel(ctx context.Context, id string) error
}

// FuelUpdates contains the fields that can be updated for a fuel record
type FuelUpdates struct {
	FillDate              *time.Time
	Mileage               *int
	Volume                *float64
	FuelType              *string
	PartialFill           *bool
	PricePerUnit          *float64
	OctaneRating          *int
	Location              *string
	Brand                 *string
	Notes                 *string
	CityDrivingPercentage *int
	VehicleReportedMPG    *float64
}

// DefaultFuelService implements FuelService using a FuelRepository
type DefaultFuelService struct {
	repo        repositories.FuelRepository
	vehicleRepo repositories.VehicleRepository
	metricsRepo repositories.MetricsRepository
}

// NewFuelService creates a new DefaultFuelService
func NewFuelService(repo repositories.FuelRepository, vehicleRepo repositories.VehicleRepository, metricsRepo repositories.MetricsRepository) *DefaultFuelService {
	return &DefaultFuelService{
		repo:        repo,
		vehicleRepo: vehicleRepo,
		metricsRepo: metricsRepo,
	}
}

// CreateFuel creates a new fuel record
func (s *DefaultFuelService) CreateFuel(ctx context.Context, record *models.FuelRecord) error {
	// Verify the vehicle exists
	_, err := s.vehicleRepo.FindByID(ctx, record.VehicleID)
	if err != nil {
		return err
	}

	if err := s.repo.Create(ctx, record); err != nil {
		return err
	}

	s.recalculateMetrics(ctx, record.VehicleID)
	return nil
}

// GetFuel retrieves a fuel record by its ID
func (s *DefaultFuelService) GetFuel(ctx context.Context, id string) (*models.FuelRecord, error) {
	return s.repo.FindByID(ctx, id)
}

// GetVehicleFuel retrieves all fuel records for a specific vehicle
func (s *DefaultFuelService) GetVehicleFuel(ctx context.Context, vehicleID string) ([]*models.FuelRecord, error) {
	return s.repo.FindByVehicleID(ctx, vehicleID)
}

// ListFuel retrieves fuel records with optional filters and pagination
func (s *DefaultFuelService) ListFuel(ctx context.Context, filters repositories.FuelFilters, pagination repositories.PaginationParams) ([]*models.FuelRecord, error) {
	return s.repo.List(ctx, filters, pagination)
}

// CountFuel returns the total number of fuel records matching the filters
func (s *DefaultFuelService) CountFuel(ctx context.Context, filters repositories.FuelFilters) (int, error) {
	return s.repo.Count(ctx, filters)
}

// UpdateFuel updates a fuel record
func (s *DefaultFuelService) UpdateFuel(ctx context.Context, id string, updates FuelUpdates) (*models.FuelRecord, error) {
	// Get the existing record
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.FillDate != nil {
		record.FillDate = *updates.FillDate
	}
	if updates.Mileage != nil {
		record.Mileage = *updates.Mileage
	}
	if updates.Volume != nil {
		record.Volume = *updates.Volume
	}
	if updates.FuelType != nil {
		record.FuelType = *updates.FuelType
	}
	if updates.PartialFill != nil {
		record.PartialFill = *updates.PartialFill
	}
	if updates.PricePerUnit != nil {
		record.PricePerUnit = updates.PricePerUnit
	}
	if updates.OctaneRating != nil {
		record.OctaneRating = updates.OctaneRating
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
	if updates.CityDrivingPercentage != nil {
		record.CityDrivingPercentage = updates.CityDrivingPercentage
	}
	if updates.VehicleReportedMPG != nil {
		record.VehicleReportedMPG = updates.VehicleReportedMPG
	}

	// Update the record
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}

	s.recalculateMetrics(ctx, record.VehicleID)
	return record, nil
}

// DeleteFuel deletes a fuel record
func (s *DefaultFuelService) DeleteFuel(ctx context.Context, id string) error {
	// Get the record first to find the vehicle ID for metrics recalculation
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.recalculateMetrics(ctx, record.VehicleID)
	return nil
}

// recalculateMetrics recalculates and upserts the total_fuel_spent metric for a vehicle.
// Errors are silently ignored since metrics are non-critical and the primary
// fuel operation has already succeeded.
func (s *DefaultFuelService) recalculateMetrics(ctx context.Context, vehicleID string) {
	if s.metricsRepo == nil {
		return
	}

	totalFuelSpent, err := s.repo.SumCostByVehicleID(ctx, vehicleID)
	if err != nil {
		return
	}

	// nil means no records with cost — treat as zero spent so the Upsert
	// overwrites any stale value instead of preserving it via COALESCE.
	if totalFuelSpent == nil {
		zero := 0.0
		totalFuelSpent = &zero
	}

	metrics := &models.VehicleMetrics{
		VehicleID:      vehicleID,
		TotalFuelSpent: totalFuelSpent,
	}

	_ = s.metricsRepo.Upsert(ctx, metrics)
}
