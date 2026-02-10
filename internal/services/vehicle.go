package services

import (
	"context"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
)

// VehicleService defines the interface for vehicle business logic
type VehicleService interface {
	// CreateVehicle creates a new vehicle
	CreateVehicle(ctx context.Context, vehicle *models.Vehicle) error

	// GetVehicle retrieves a vehicle by its ID
	GetVehicle(ctx context.Context, id string) (*models.Vehicle, error)

	// GetUserVehicles retrieves all vehicles for a specific user
	GetUserVehicles(ctx context.Context, userID string) ([]*models.Vehicle, error)

	// UpdateVehicle updates a vehicle's information
	UpdateVehicle(ctx context.Context, id string, updates VehicleUpdates) (*models.Vehicle, error)

	// ArchiveVehicle archives a vehicle (sets status to sold or scrapped)
	ArchiveVehicle(ctx context.Context, id string, status models.VehicleStatus) (*models.Vehicle, error)

	// ListVehicles retrieves vehicles with optional filters and pagination
	ListVehicles(ctx context.Context, filters repositories.VehicleFilters, pagination repositories.PaginationParams) ([]*models.Vehicle, error)

	// VerifyOwnership verifies that a vehicle belongs to a specific user
	VerifyOwnership(ctx context.Context, vehicleID, userID string) error
}

// VehicleUpdates contains the fields that can be updated for a vehicle
type VehicleUpdates struct {
	VIN            *string
	Make           *string
	Model          *string
	Year           *int
	Color          *string
	LicensePlate   *string
	CurrentMileage *int
	Notes          *string
}

// DefaultVehicleService implements VehicleService using a VehicleRepository
type DefaultVehicleService struct {
	repo repositories.VehicleRepository
}

// NewVehicleService creates a new DefaultVehicleService
func NewVehicleService(repo repositories.VehicleRepository) *DefaultVehicleService {
	return &DefaultVehicleService{repo: repo}
}

// CreateVehicle creates a new vehicle
func (s *DefaultVehicleService) CreateVehicle(ctx context.Context, vehicle *models.Vehicle) error {
	return s.repo.Create(ctx, vehicle)
}

// GetVehicle retrieves a vehicle by its ID
func (s *DefaultVehicleService) GetVehicle(ctx context.Context, id string) (*models.Vehicle, error) {
	return s.repo.FindByID(ctx, id)
}

// GetUserVehicles retrieves all vehicles for a specific user
func (s *DefaultVehicleService) GetUserVehicles(ctx context.Context, userID string) ([]*models.Vehicle, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// UpdateVehicle updates a vehicle's information
func (s *DefaultVehicleService) UpdateVehicle(ctx context.Context, id string, updates VehicleUpdates) (*models.Vehicle, error) {
	// Get the existing vehicle
	vehicle, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.VIN != nil {
		vehicle.VIN = *updates.VIN
	}
	if updates.Make != nil {
		vehicle.Make = *updates.Make
	}
	if updates.Model != nil {
		vehicle.Model = *updates.Model
	}
	if updates.Year != nil {
		vehicle.Year = *updates.Year
	}
	if updates.Color != nil {
		vehicle.Color = *updates.Color
	}
	if updates.LicensePlate != nil {
		vehicle.LicensePlate = *updates.LicensePlate
	}
	if updates.CurrentMileage != nil {
		vehicle.CurrentMileage = updates.CurrentMileage
	}
	if updates.Notes != nil {
		vehicle.Notes = *updates.Notes
	}

	// Update the vehicle
	if err := s.repo.Update(ctx, vehicle); err != nil {
		return nil, err
	}

	return vehicle, nil
}

// ArchiveVehicle archives a vehicle (sets status to sold or scrapped)
func (s *DefaultVehicleService) ArchiveVehicle(ctx context.Context, id string, status models.VehicleStatus) (*models.Vehicle, error) {
	// Validate the status is an archive status
	if status != models.VehicleStatusSold && status != models.VehicleStatusScrapped {
		return nil, models.NewValidationError("status", "archive status must be 'sold' or 'scrapped'")
	}

	// Get the existing vehicle
	vehicle, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update the status
	vehicle.Status = status

	// Update the vehicle
	if err := s.repo.Update(ctx, vehicle); err != nil {
		return nil, err
	}

	return vehicle, nil
}

// ListVehicles retrieves vehicles with optional filters and pagination
func (s *DefaultVehicleService) ListVehicles(ctx context.Context, filters repositories.VehicleFilters, pagination repositories.PaginationParams) ([]*models.Vehicle, error) {
	return s.repo.List(ctx, filters, pagination)
}

// VerifyOwnership verifies that a vehicle belongs to a specific user
func (s *DefaultVehicleService) VerifyOwnership(ctx context.Context, vehicleID, userID string) error {
	vehicle, err := s.repo.FindByID(ctx, vehicleID)
	if err != nil {
		return err
	}

	if vehicle.UserID != userID {
		return models.NewNotFoundError("Vehicle", vehicleID)
	}

	return nil
}
