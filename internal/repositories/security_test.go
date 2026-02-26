package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
)

// TestSQLInjection_UserRepository verifies that parameterized queries protect
// the user repository against SQL injection attacks.
func TestSQLInjection_UserRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	// Seed a valid user so there is data in the table.
	user := &models.User{
		Username:     "safeuser",
		Email:        "safe@example.com",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	injectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE users; --",
		"' UNION SELECT * FROM users --",
		"1' OR '1'='1' --",
		"admin'--",
		"' OR 1=1 --",
		`"; DROP TABLE users; --`,
	}

	t.Run("FindByID rejects SQL injection payloads", func(t *testing.T) {
		for _, payload := range injectionPayloads {
			result, err := userRepo.FindByID(ctx, payload)
			assert.Nil(t, result, "payload %q should not return a user", payload)
			assert.Error(t, err, "payload %q should produce an error (not found)", payload)
		}
	})

	t.Run("FindByEmail rejects SQL injection payloads", func(t *testing.T) {
		for _, payload := range injectionPayloads {
			result, err := userRepo.FindByEmail(ctx, payload)
			assert.Nil(t, result, "payload %q should not return a user", payload)
			assert.Error(t, err, "payload %q should produce an error (not found)", payload)
		}
	})

	t.Run("FindByUsername rejects SQL injection payloads", func(t *testing.T) {
		for _, payload := range injectionPayloads {
			result, err := userRepo.FindByUsername(ctx, payload)
			assert.Nil(t, result, "payload %q should not return a user", payload)
			assert.Error(t, err, "payload %q should produce an error (not found)", payload)
		}
	})
}

// TestSQLInjection_VehicleRepository verifies that parameterized queries protect
// the vehicle repository against SQL injection attacks.
func TestSQLInjection_VehicleRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	// Seed a valid user and vehicle.
	user := &models.User{
		Username:     "vehicleowner",
		Email:        "vowner@example.com",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1HGBH41JXMN109186",
		Make:   "Honda",
		Model:  "Civic",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	require.NoError(t, vehicleRepo.Create(ctx, vehicle))

	injectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE vehicles; --",
		"' UNION SELECT * FROM vehicles --",
	}

	t.Run("FindByID rejects SQL injection payloads", func(t *testing.T) {
		for _, payload := range injectionPayloads {
			result, err := vehicleRepo.FindByID(ctx, payload)
			assert.Nil(t, result, "payload %q should not return a vehicle", payload)
			assert.Error(t, err, "payload %q should produce an error", payload)
		}
	})

	t.Run("FindByVIN rejects SQL injection payloads", func(t *testing.T) {
		for _, payload := range injectionPayloads {
			result, err := vehicleRepo.FindByVIN(ctx, payload)
			assert.Nil(t, result, "payload %q should not return a vehicle", payload)
			assert.Error(t, err, "payload %q should produce an error", payload)
		}
	})

	t.Run("FindByUserID rejects SQL injection payloads", func(t *testing.T) {
		for _, payload := range injectionPayloads {
			results, err := vehicleRepo.FindByUserID(ctx, payload)
			require.NoError(t, err, "parameterized query should not fail for payload %q", payload)
			assert.Empty(t, results, "payload %q should not return vehicles", payload)
		}
	})

	t.Run("List filter values reject SQL injection", func(t *testing.T) {
		for _, payload := range injectionPayloads {
			filters := VehicleFilters{UserID: &payload}
			results, err := vehicleRepo.List(ctx, filters, PaginationParams{})
			require.NoError(t, err, "parameterized query should not fail for payload %q", payload)
			assert.Empty(t, results, "payload %q should not match any vehicles", payload)
		}
	})

	t.Run("Delete with SQL injection payload does not affect data", func(t *testing.T) {
		err := vehicleRepo.Delete(ctx, "'; DROP TABLE vehicles; --")
		assert.Error(t, err, "should return not-found error")

		// Verify the original vehicle still exists.
		found, err := vehicleRepo.FindByID(ctx, vehicle.ID)
		require.NoError(t, err)
		assert.Equal(t, vehicle.ID, found.ID)
	})
}

// TestSQLInjection_MaintenanceRepository verifies that parameterized queries protect
// the maintenance repository against SQL injection attacks.
func TestSQLInjection_MaintenanceRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)
	ctx := context.Background()

	// Seed prerequisite data.
	user := &models.User{
		Username:     "maintowner",
		Email:        "mowner@example.com",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "2HGBH41JXMN109187",
		Make:   "Toyota",
		Model:  "Camry",
		Year:   2021,
		Status: models.VehicleStatusActive,
	}
	require.NoError(t, vehicleRepo.Create(ctx, vehicle))

	injectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE maintenance_records; --",
		"' UNION SELECT * FROM maintenance_records --",
	}

	t.Run("FindByID rejects SQL injection payloads", func(t *testing.T) {
		for _, payload := range injectionPayloads {
			result, err := maintenanceRepo.FindByID(ctx, payload)
			assert.Nil(t, result, "payload %q should not return a record", payload)
			assert.Error(t, err, "payload %q should produce an error", payload)
		}
	})

	t.Run("FindByVehicleID rejects SQL injection payloads", func(t *testing.T) {
		for _, payload := range injectionPayloads {
			results, err := maintenanceRepo.FindByVehicleID(ctx, payload)
			require.NoError(t, err, "parameterized query should not fail for payload %q", payload)
			assert.Empty(t, results, "payload %q should not return records", payload)
		}
	})
}
