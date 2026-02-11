package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
)

// mockVehicleRepository is a mock implementation of VehicleRepository for testing
type mockVehicleRepository struct {
	vehicles   map[string]*models.Vehicle
	createErr  error
	findErr    error
	updateErr  error
	deleteErr  error
	listResult []*models.Vehicle
}

func newMockVehicleRepository() *mockVehicleRepository {
	return &mockVehicleRepository{
		vehicles: make(map[string]*models.Vehicle),
	}
}

func (m *mockVehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	if m.createErr != nil {
		return m.createErr
	}
	if vehicle.ID == "" {
		vehicle.ID = "test-vehicle-id"
	}
	m.vehicles[vehicle.ID] = vehicle
	return nil
}

func (m *mockVehicleRepository) FindByID(ctx context.Context, id string) (*models.Vehicle, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	vehicle, exists := m.vehicles[id]
	if !exists {
		return nil, models.NewNotFoundError("Vehicle", id)
	}
	return vehicle, nil
}

func (m *mockVehicleRepository) FindByUserID(ctx context.Context, userID string) ([]*models.Vehicle, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []*models.Vehicle
	for _, v := range m.vehicles {
		if v.UserID == userID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockVehicleRepository) FindByVIN(ctx context.Context, vin string) (*models.Vehicle, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, v := range m.vehicles {
		if v.VIN == vin {
			return v, nil
		}
	}
	return nil, models.NewNotFoundError("Vehicle", vin)
}

func (m *mockVehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, exists := m.vehicles[vehicle.ID]; !exists {
		return models.NewNotFoundError("Vehicle", vehicle.ID)
	}
	m.vehicles[vehicle.ID] = vehicle
	return nil
}

func (m *mockVehicleRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, exists := m.vehicles[id]; !exists {
		return models.NewNotFoundError("Vehicle", id)
	}
	delete(m.vehicles, id)
	return nil
}

func (m *mockVehicleRepository) List(ctx context.Context, filters repositories.VehicleFilters, pagination repositories.PaginationParams) ([]*models.Vehicle, error) {
	if m.listResult != nil {
		return m.listResult, nil
	}
	var result []*models.Vehicle
	for _, v := range m.vehicles {
		result = append(result, v)
	}
	return result, nil
}

func (m *mockVehicleRepository) Count(ctx context.Context, filters repositories.VehicleFilters) (int, error) {
	if m.listResult != nil {
		return len(m.listResult), nil
	}
	return len(m.vehicles), nil
}

func TestVehicleService_CreateVehicle(t *testing.T) {
	ctx := context.Background()

	t.Run("creates vehicle successfully", func(t *testing.T) {
		repo := newMockVehicleRepository()
		service := NewVehicleService(repo)

		vehicle := &models.Vehicle{
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}

		err := service.CreateVehicle(ctx, vehicle)
		require.NoError(t, err)
		assert.NotEmpty(t, vehicle.ID)
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		repo := newMockVehicleRepository()
		repo.createErr = models.NewDatabaseError("create", assert.AnError)
		service := NewVehicleService(repo)

		vehicle := &models.Vehicle{
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}

		err := service.CreateVehicle(ctx, vehicle)
		assert.Error(t, err)
	})
}

func TestVehicleService_GetVehicle(t *testing.T) {
	ctx := context.Background()

	t.Run("returns vehicle by ID", func(t *testing.T) {
		repo := newMockVehicleRepository()
		existingVehicle := &models.Vehicle{
			ID:     "vehicle-123",
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		repo.vehicles[existingVehicle.ID] = existingVehicle
		service := NewVehicleService(repo)

		vehicle, err := service.GetVehicle(ctx, "vehicle-123")
		require.NoError(t, err)
		assert.Equal(t, "Honda", vehicle.Make)
		assert.Equal(t, "Civic", vehicle.Model)
	})

	t.Run("returns not found error for non-existent vehicle", func(t *testing.T) {
		repo := newMockVehicleRepository()
		service := NewVehicleService(repo)

		_, err := service.GetVehicle(ctx, "non-existent")
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestVehicleService_GetUserVehicles(t *testing.T) {
	ctx := context.Background()

	t.Run("returns all vehicles for user", func(t *testing.T) {
		repo := newMockVehicleRepository()
		vehicle1 := &models.Vehicle{
			ID:     "vehicle-1",
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		vehicle2 := &models.Vehicle{
			ID:     "vehicle-2",
			UserID: "user-123",
			VIN:    "2HGBH41JXMN109187",
			Make:   "Toyota",
			Model:  "Camry",
			Year:   2020,
			Status: models.VehicleStatusActive,
		}
		repo.vehicles[vehicle1.ID] = vehicle1
		repo.vehicles[vehicle2.ID] = vehicle2
		service := NewVehicleService(repo)

		vehicles, err := service.GetUserVehicles(ctx, "user-123")
		require.NoError(t, err)
		assert.Len(t, vehicles, 2)
	})

	t.Run("returns empty slice for user with no vehicles", func(t *testing.T) {
		repo := newMockVehicleRepository()
		service := NewVehicleService(repo)

		vehicles, err := service.GetUserVehicles(ctx, "user-with-no-vehicles")
		require.NoError(t, err)
		assert.Empty(t, vehicles)
	})
}

func TestVehicleService_UpdateVehicle(t *testing.T) {
	ctx := context.Background()

	t.Run("updates vehicle fields", func(t *testing.T) {
		repo := newMockVehicleRepository()
		existingVehicle := &models.Vehicle{
			ID:     "vehicle-123",
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		repo.vehicles[existingVehicle.ID] = existingVehicle
		service := NewVehicleService(repo)

		newColor := "Blue"
		newMileage := 15000
		updates := VehicleUpdates{
			Color:          &newColor,
			CurrentMileage: &newMileage,
		}

		updatedVehicle, err := service.UpdateVehicle(ctx, "vehicle-123", updates)
		require.NoError(t, err)
		assert.Equal(t, "Blue", updatedVehicle.Color)
		assert.Equal(t, 15000, *updatedVehicle.CurrentMileage)
		assert.Equal(t, "Honda", updatedVehicle.Make) // unchanged
	})

	t.Run("returns not found for non-existent vehicle", func(t *testing.T) {
		repo := newMockVehicleRepository()
		service := NewVehicleService(repo)

		_, err := service.UpdateVehicle(ctx, "non-existent", VehicleUpdates{})
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestVehicleService_ArchiveVehicle(t *testing.T) {
	ctx := context.Background()

	t.Run("archives vehicle as sold", func(t *testing.T) {
		repo := newMockVehicleRepository()
		existingVehicle := &models.Vehicle{
			ID:     "vehicle-123",
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		repo.vehicles[existingVehicle.ID] = existingVehicle
		service := NewVehicleService(repo)

		archivedVehicle, err := service.ArchiveVehicle(ctx, "vehicle-123", models.VehicleStatusSold)
		require.NoError(t, err)
		assert.Equal(t, models.VehicleStatusSold, archivedVehicle.Status)
	})

	t.Run("archives vehicle as scrapped", func(t *testing.T) {
		repo := newMockVehicleRepository()
		existingVehicle := &models.Vehicle{
			ID:     "vehicle-123",
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		repo.vehicles[existingVehicle.ID] = existingVehicle
		service := NewVehicleService(repo)

		archivedVehicle, err := service.ArchiveVehicle(ctx, "vehicle-123", models.VehicleStatusScrapped)
		require.NoError(t, err)
		assert.Equal(t, models.VehicleStatusScrapped, archivedVehicle.Status)
	})

	t.Run("returns validation error for active status", func(t *testing.T) {
		repo := newMockVehicleRepository()
		existingVehicle := &models.Vehicle{
			ID:     "vehicle-123",
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		repo.vehicles[existingVehicle.ID] = existingVehicle
		service := NewVehicleService(repo)

		_, err := service.ArchiveVehicle(ctx, "vehicle-123", models.VehicleStatusActive)
		assert.Error(t, err)

		var validationErr *models.ValidationError
		assert.ErrorAs(t, err, &validationErr)
	})

	t.Run("returns not found for non-existent vehicle", func(t *testing.T) {
		repo := newMockVehicleRepository()
		service := NewVehicleService(repo)

		_, err := service.ArchiveVehicle(ctx, "non-existent", models.VehicleStatusSold)
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestVehicleService_ListVehicles(t *testing.T) {
	ctx := context.Background()

	t.Run("lists vehicles with filters and pagination", func(t *testing.T) {
		repo := newMockVehicleRepository()
		vehicle1 := &models.Vehicle{
			ID:     "vehicle-1",
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		repo.listResult = []*models.Vehicle{vehicle1}
		service := NewVehicleService(repo)

		filters := repositories.VehicleFilters{}
		pagination := repositories.PaginationParams{Limit: 10, Offset: 0}

		vehicles, err := service.ListVehicles(ctx, filters, pagination)
		require.NoError(t, err)
		assert.Len(t, vehicles, 1)
	})
}

func TestVehicleService_VerifyOwnership(t *testing.T) {
	ctx := context.Background()

	t.Run("verifies ownership successfully", func(t *testing.T) {
		repo := newMockVehicleRepository()
		existingVehicle := &models.Vehicle{
			ID:     "vehicle-123",
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		repo.vehicles[existingVehicle.ID] = existingVehicle
		service := NewVehicleService(repo)

		err := service.VerifyOwnership(ctx, "vehicle-123", "user-123")
		assert.NoError(t, err)
	})

	t.Run("returns not found for wrong user", func(t *testing.T) {
		repo := newMockVehicleRepository()
		existingVehicle := &models.Vehicle{
			ID:     "vehicle-123",
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		repo.vehicles[existingVehicle.ID] = existingVehicle
		service := NewVehicleService(repo)

		err := service.VerifyOwnership(ctx, "vehicle-123", "different-user")
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})

	t.Run("returns not found for non-existent vehicle", func(t *testing.T) {
		repo := newMockVehicleRepository()
		service := NewVehicleService(repo)

		err := service.VerifyOwnership(ctx, "non-existent", "user-123")
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}
