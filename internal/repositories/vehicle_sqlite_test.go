package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
)

func TestVehicleRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	// Create a user first
	user := &models.User{
		Username:     "vehicleowner",
		Email:        "owner@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("create valid vehicle", func(t *testing.T) {
		vehicle := &models.Vehicle{
			UserID: user.ID,
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2020,
			Status: models.VehicleStatusActive,
		}

		err := vehicleRepo.Create(ctx, vehicle)
		require.NoError(t, err)
		assert.NotEmpty(t, vehicle.ID)
		assert.False(t, vehicle.CreatedAt.IsZero())
		assert.False(t, vehicle.UpdatedAt.IsZero())
	})

	t.Run("create with all fields", func(t *testing.T) {
		purchasePrice := 25000.0
		purchaseMileage := 10000
		currentMileage := 15000
		purchaseDate := time.Now().Add(-365 * 24 * time.Hour)

		vehicle := &models.Vehicle{
			UserID:          user.ID,
			VIN:             "2HGBH41JXMN109187",
			Make:            "Toyota",
			Model:           "Camry",
			Year:            2019,
			Color:           "Blue",
			LicensePlate:    "ABC123",
			PurchaseDate:    &purchaseDate,
			PurchasePrice:   &purchasePrice,
			PurchaseMileage: &purchaseMileage,
			CurrentMileage:  &currentMileage,
			Status:          models.VehicleStatusActive,
			Notes:           "Great car",
		}

		err := vehicleRepo.Create(ctx, vehicle)
		require.NoError(t, err)
		assert.NotEmpty(t, vehicle.ID)
	})

	t.Run("duplicate VIN", func(t *testing.T) {
		vehicle1 := &models.Vehicle{
			UserID: user.ID,
			VIN:    "3HGBH41JXMN109188",
			Make:   "Ford",
			Model:  "Focus",
			Year:   2018,
			Status: models.VehicleStatusActive,
		}

		err := vehicleRepo.Create(ctx, vehicle1)
		require.NoError(t, err)

		vehicle2 := &models.Vehicle{
			UserID: user.ID,
			VIN:    "3HGBH41JXMN109188", // Same VIN
			Make:   "Ford",
			Model:  "Fusion",
			Year:   2019,
			Status: models.VehicleStatusActive,
		}

		err = vehicleRepo.Create(ctx, vehicle2)
		require.Error(t, err)
		assert.IsType(t, &models.DuplicateError{}, err)
		assert.Contains(t, err.Error(), "vin")
	})

	t.Run("invalid user ID", func(t *testing.T) {
		vehicle := &models.Vehicle{
			UserID: "non-existent-user",
			VIN:    "4HGBH41JXMN109189",
			Make:   "Nissan",
			Model:  "Altima",
			Year:   2020,
			Status: models.VehicleStatusActive,
		}

		err := vehicleRepo.Create(ctx, vehicle)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
		assert.Contains(t, err.Error(), "user does not exist")
	})

	t.Run("validation error", func(t *testing.T) {
		vehicle := &models.Vehicle{
			UserID: user.ID,
			VIN:    "", // Invalid: empty VIN
			Make:   "Honda",
			Model:  "Accord",
			Year:   2020,
			Status: models.VehicleStatusActive,
		}

		err := vehicleRepo.Create(ctx, vehicle)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
	})
}

func TestVehicleRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "findvehicleowner",
		Email:        "findowner@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("find existing vehicle", func(t *testing.T) {
		vehicle := &models.Vehicle{
			UserID: user.ID,
			VIN:    "5HGBH41JXMN109190",
			Make:   "Honda",
			Model:  "CR-V",
			Year:   2021,
			Color:  "Red",
			Status: models.VehicleStatusActive,
			Notes:  "Test vehicle",
		}

		err := vehicleRepo.Create(ctx, vehicle)
		require.NoError(t, err)

		found, err := vehicleRepo.FindByID(ctx, vehicle.ID)
		require.NoError(t, err)
		assert.Equal(t, vehicle.ID, found.ID)
		assert.Equal(t, vehicle.VIN, found.VIN)
		assert.Equal(t, vehicle.Make, found.Make)
		assert.Equal(t, vehicle.Model, found.Model)
		assert.Equal(t, vehicle.Color, found.Color)
		assert.Equal(t, vehicle.Notes, found.Notes)
	})

	t.Run("vehicle not found", func(t *testing.T) {
		_, err := vehicleRepo.FindByID(ctx, "non-existent-id")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestVehicleRepository_FindByVIN(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "vinowner",
		Email:        "vinowner@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("find existing vehicle", func(t *testing.T) {
		vehicle := &models.Vehicle{
			UserID: user.ID,
			VIN:    "6HGBH41JXMN109191",
			Make:   "Toyota",
			Model:  "RAV4",
			Year:   2020,
			Status: models.VehicleStatusActive,
		}

		err := vehicleRepo.Create(ctx, vehicle)
		require.NoError(t, err)

		found, err := vehicleRepo.FindByVIN(ctx, vehicle.VIN)
		require.NoError(t, err)
		assert.Equal(t, vehicle.ID, found.ID)
		assert.Equal(t, vehicle.VIN, found.VIN)
	})

	t.Run("vehicle not found", func(t *testing.T) {
		_, err := vehicleRepo.FindByVIN(ctx, "NONEXISTENT123456")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestVehicleRepository_FindByUserID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user1 := &models.User{
		Username:     "user1vehicles",
		Email:        "user1@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user1)
	require.NoError(t, err)

	user2 := &models.User{
		Username:     "user2vehicles",
		Email:        "user2@example.com",
		PasswordHash: "hashed_password",
	}
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	t.Run("find vehicles for user with vehicles", func(t *testing.T) {
		// Create vehicles for user1
		vehicle1 := &models.Vehicle{
			UserID: user1.ID,
			VIN:    "7HGBH41JXMN109192",
			Make:   "Honda",
			Model:  "Accord",
			Year:   2020,
			Status: models.VehicleStatusActive,
		}
		err := vehicleRepo.Create(ctx, vehicle1)
		require.NoError(t, err)

		vehicle2 := &models.Vehicle{
			UserID: user1.ID,
			VIN:    "8HGBH41JXMN109193",
			Make:   "Toyota",
			Model:  "Corolla",
			Year:   2019,
			Status: models.VehicleStatusActive,
		}
		err = vehicleRepo.Create(ctx, vehicle2)
		require.NoError(t, err)

		// Create vehicle for user2
		vehicle3 := &models.Vehicle{
			UserID: user2.ID,
			VIN:    "9HGBH41JXMN109194",
			Make:   "Ford",
			Model:  "Mustang",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		err = vehicleRepo.Create(ctx, vehicle3)
		require.NoError(t, err)

		// Find vehicles for user1
		vehicles, err := vehicleRepo.FindByUserID(ctx, user1.ID)
		require.NoError(t, err)
		assert.Len(t, vehicles, 2)
	})

	t.Run("find vehicles for user with no vehicles", func(t *testing.T) {
		user3 := &models.User{
			Username:     "user3novehicles",
			Email:        "user3@example.com",
			PasswordHash: "hashed_password",
		}
		err := userRepo.Create(ctx, user3)
		require.NoError(t, err)

		vehicles, err := vehicleRepo.FindByUserID(ctx, user3.ID)
		require.NoError(t, err)
		assert.Empty(t, vehicles)
	})
}

func TestVehicleRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "updatevehicleowner",
		Email:        "updateowner@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("update existing vehicle", func(t *testing.T) {
		vehicle := &models.Vehicle{
			UserID: user.ID,
			VIN:    "AHGBH41JXMN109195",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2020,
			Color:  "Blue",
			Status: models.VehicleStatusActive,
		}

		err := vehicleRepo.Create(ctx, vehicle)
		require.NoError(t, err)

		oldUpdatedAt := vehicle.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		// Update vehicle
		vehicle.Color = "Red"
		vehicle.Model = "Accord"
		newMileage := 50000
		vehicle.CurrentMileage = &newMileage

		err = vehicleRepo.Update(ctx, vehicle)
		require.NoError(t, err)
		assert.True(t, vehicle.UpdatedAt.After(oldUpdatedAt))

		// Verify update
		found, err := vehicleRepo.FindByID(ctx, vehicle.ID)
		require.NoError(t, err)
		assert.Equal(t, "Red", found.Color)
		assert.Equal(t, "Accord", found.Model)
		assert.NotNil(t, found.CurrentMileage)
		assert.Equal(t, 50000, *found.CurrentMileage)
	})

	t.Run("update non-existent vehicle", func(t *testing.T) {
		vehicle := &models.Vehicle{
			ID:     "non-existent",
			UserID: user.ID,
			VIN:    "BHGBH41JXMN109196",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2020,
			Status: models.VehicleStatusActive,
		}

		err := vehicleRepo.Update(ctx, vehicle)
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestVehicleRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "deletevehicleowner",
		Email:        "deleteowner@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("delete existing vehicle", func(t *testing.T) {
		vehicle := &models.Vehicle{
			UserID: user.ID,
			VIN:    "CHGBH41JXMN109197",
			Make:   "Toyota",
			Model:  "Camry",
			Year:   2019,
			Status: models.VehicleStatusActive,
		}

		err := vehicleRepo.Create(ctx, vehicle)
		require.NoError(t, err)

		err = vehicleRepo.Delete(ctx, vehicle.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = vehicleRepo.FindByID(ctx, vehicle.ID)
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})

	t.Run("delete non-existent vehicle", func(t *testing.T) {
		err := vehicleRepo.Delete(ctx, "non-existent-id")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestVehicleRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "listvehicleowner",
		Email:        "listowner@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test vehicles
	vehicles := []*models.Vehicle{
		{
			UserID: user.ID,
			VIN:    "DHGBH41JXMN109198",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2020,
			Status: models.VehicleStatusActive,
		},
		{
			UserID: user.ID,
			VIN:    "EHGBH41JXMN109199",
			Make:   "Honda",
			Model:  "Accord",
			Year:   2021,
			Status: models.VehicleStatusActive,
		},
		{
			UserID: user.ID,
			VIN:    "FHGBH41JXMN109200",
			Make:   "Toyota",
			Model:  "Camry",
			Year:   2020,
			Status: models.VehicleStatusSold,
		},
	}

	for _, v := range vehicles {
		err := vehicleRepo.Create(ctx, v)
		require.NoError(t, err)
	}

	t.Run("list all vehicles", func(t *testing.T) {
		result, err := vehicleRepo.List(ctx, VehicleFilters{}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("filter by status", func(t *testing.T) {
		status := models.VehicleStatusActive
		result, err := vehicleRepo.List(ctx, VehicleFilters{Status: &status}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
		for _, v := range result {
			assert.Equal(t, models.VehicleStatusActive, v.Status)
		}
	})

	t.Run("filter by make", func(t *testing.T) {
		make := "Honda"
		result, err := vehicleRepo.List(ctx, VehicleFilters{Make: &make}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
		for _, v := range result {
			assert.Equal(t, "Honda", v.Make)
		}
	})

	t.Run("filter by year", func(t *testing.T) {
		year := 2020
		result, err := vehicleRepo.List(ctx, VehicleFilters{Year: &year}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
		for _, v := range result {
			assert.Equal(t, 2020, v.Year)
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		result, err := vehicleRepo.List(ctx, VehicleFilters{}, PaginationParams{Limit: 2})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(result), 2)
	})

	t.Run("filter by model", func(t *testing.T) {
		model := "Civic"
		result, err := vehicleRepo.List(ctx, VehicleFilters{Model: &model}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)
		for _, v := range result {
			assert.Equal(t, "Civic", v.Model)
		}
	})

	t.Run("filter by user ID", func(t *testing.T) {
		result, err := vehicleRepo.List(ctx, VehicleFilters{UserID: &user.ID}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("with offset pagination", func(t *testing.T) {
		all, err := vehicleRepo.List(ctx, VehicleFilters{}, PaginationParams{})
		require.NoError(t, err)

		result, err := vehicleRepo.List(ctx, VehicleFilters{}, PaginationParams{Limit: 10, Offset: len(all)})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestVehicleRepository_FindByID_AllFields(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "allfieldowner",
		Email:        "allfield@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("find vehicle with all nullable fields", func(t *testing.T) {
		purchasePrice := 25000.0
		purchaseMileage := 10000
		currentMileage := 15000
		purchaseDate := time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)

		vehicle := &models.Vehicle{
			UserID:          user.ID,
			VIN:             "XHGBH41JXMN109301",
			Make:            "Honda",
			Model:           "Civic",
			Year:            2020,
			Color:           "Red",
			LicensePlate:    "XYZ789",
			PurchaseDate:    &purchaseDate,
			PurchasePrice:   &purchasePrice,
			PurchaseMileage: &purchaseMileage,
			CurrentMileage:  &currentMileage,
			Status:          models.VehicleStatusActive,
			Notes:           "Test notes",
		}

		err := vehicleRepo.Create(ctx, vehicle)
		require.NoError(t, err)

		found, err := vehicleRepo.FindByID(ctx, vehicle.ID)
		require.NoError(t, err)
		assert.Equal(t, "Red", found.Color)
		assert.Equal(t, "XYZ789", found.LicensePlate)
		assert.Equal(t, "Test notes", found.Notes)
		assert.NotNil(t, found.PurchaseDate)
		assert.NotNil(t, found.PurchasePrice)
		assert.Equal(t, 25000.0, *found.PurchasePrice)
		assert.NotNil(t, found.PurchaseMileage)
		assert.Equal(t, 10000, *found.PurchaseMileage)
		assert.NotNil(t, found.CurrentMileage)
		assert.Equal(t, 15000, *found.CurrentMileage)
	})
}

func TestVehicleRepository_FindByVIN_AllFields(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "vinallfieldowner",
		Email:        "vinallfield@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("find vehicle by VIN with all nullable fields", func(t *testing.T) {
		purchasePrice := 30000.0
		purchaseMileage := 5000
		currentMileage := 20000
		purchaseDate := time.Date(2021, 3, 10, 0, 0, 0, 0, time.UTC)

		vehicle := &models.Vehicle{
			UserID:          user.ID,
			VIN:             "YHGBH41JXMN109302",
			Make:            "Toyota",
			Model:           "RAV4",
			Year:            2021,
			Color:           "Blue",
			LicensePlate:    "ABC456",
			PurchaseDate:    &purchaseDate,
			PurchasePrice:   &purchasePrice,
			PurchaseMileage: &purchaseMileage,
			CurrentMileage:  &currentMileage,
			Status:          models.VehicleStatusActive,
			Notes:           "VIN search test",
		}

		err := vehicleRepo.Create(ctx, vehicle)
		require.NoError(t, err)

		found, err := vehicleRepo.FindByVIN(ctx, vehicle.VIN)
		require.NoError(t, err)
		assert.Equal(t, "Blue", found.Color)
		assert.Equal(t, "ABC456", found.LicensePlate)
		assert.Equal(t, "VIN search test", found.Notes)
		assert.NotNil(t, found.PurchaseDate)
		assert.NotNil(t, found.PurchasePrice)
		assert.Equal(t, 30000.0, *found.PurchasePrice)
		assert.NotNil(t, found.PurchaseMileage)
		assert.Equal(t, 5000, *found.PurchaseMileage)
		assert.NotNil(t, found.CurrentMileage)
		assert.Equal(t, 20000, *found.CurrentMileage)
	})
}

func TestVehicleRepository_Update_DuplicateVIN(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "dupvinowner",
		Email:        "dupvin@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle1 := &models.Vehicle{
		UserID: user.ID,
		VIN:    "ZHGBH41JXMN109303",
		Make:   "Honda",
		Model:  "Civic",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle1)
	require.NoError(t, err)

	vehicle2 := &models.Vehicle{
		UserID: user.ID,
		VIN:    "ZHGBH41JXMN109304",
		Make:   "Toyota",
		Model:  "Camry",
		Year:   2021,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle2)
	require.NoError(t, err)

	t.Run("update to duplicate VIN returns error", func(t *testing.T) {
		vehicle2.VIN = vehicle1.VIN
		err := vehicleRepo.Update(ctx, vehicle2)
		require.Error(t, err)
		assert.IsType(t, &models.DuplicateError{}, err)
	})

	t.Run("update with validation error", func(t *testing.T) {
		vehicle := &models.Vehicle{
			ID:     vehicle1.ID,
			UserID: user.ID,
			VIN:    "",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2020,
			Status: models.VehicleStatusActive,
		}
		err := vehicleRepo.Update(ctx, vehicle)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
	})
}

func TestVehicleRepository_Count(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "countvehicleowner",
		Email:        "countowner@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicles := []*models.Vehicle{
		{
			UserID: user.ID,
			VIN:    "ZHGBH41JXMN109401",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2020,
			Status: models.VehicleStatusActive,
		},
		{
			UserID: user.ID,
			VIN:    "ZHGBH41JXMN109402",
			Make:   "Honda",
			Model:  "Accord",
			Year:   2021,
			Status: models.VehicleStatusActive,
		},
		{
			UserID: user.ID,
			VIN:    "ZHGBH41JXMN109403",
			Make:   "Toyota",
			Model:  "Camry",
			Year:   2020,
			Status: models.VehicleStatusSold,
		},
	}

	for _, v := range vehicles {
		err := vehicleRepo.Create(ctx, v)
		require.NoError(t, err)
	}

	t.Run("count all vehicles", func(t *testing.T) {
		count, err := vehicleRepo.Count(ctx, VehicleFilters{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3)
	})

	t.Run("count with user filter", func(t *testing.T) {
		count, err := vehicleRepo.Count(ctx, VehicleFilters{UserID: &user.ID})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3)
	})

	t.Run("count with status filter", func(t *testing.T) {
		status := models.VehicleStatusActive
		count, err := vehicleRepo.Count(ctx, VehicleFilters{Status: &status})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2)
	})

	t.Run("count with make filter", func(t *testing.T) {
		make := "Honda"
		count, err := vehicleRepo.Count(ctx, VehicleFilters{Make: &make})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2)
	})

	t.Run("count with model filter", func(t *testing.T) {
		model := "Civic"
		count, err := vehicleRepo.Count(ctx, VehicleFilters{Model: &model})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1)
	})

	t.Run("count with year filter", func(t *testing.T) {
		year := 2020
		count, err := vehicleRepo.Count(ctx, VehicleFilters{Year: &year})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2)
	})
}

func TestVehicleRepository_List_WithAllFields(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "listallfieldsowner",
		Email:        "listallfields@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	purchasePrice := 25000.0
	purchaseMileage := 10000
	currentMileage := 15000
	purchaseDate := time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)

	vehicle := &models.Vehicle{
		UserID:          user.ID,
		VIN:             "ZHGBH41JXMN109601",
		Make:            "Honda",
		Model:           "Civic",
		Year:            2020,
		Color:           "Red",
		LicensePlate:    "LIST123",
		PurchaseDate:    &purchaseDate,
		PurchasePrice:   &purchasePrice,
		PurchaseMileage: &purchaseMileage,
		CurrentMileage:  &currentMileage,
		Status:          models.VehicleStatusActive,
		Notes:           "List test with all fields",
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	t.Run("list returns vehicles with all nullable fields", func(t *testing.T) {
		result, err := vehicleRepo.List(ctx, VehicleFilters{UserID: &user.ID}, PaginationParams{})
		require.NoError(t, err)
		require.Len(t, result, 1)

		found := result[0]
		assert.Equal(t, "Red", found.Color)
		assert.Equal(t, "LIST123", found.LicensePlate)
		assert.Equal(t, "List test with all fields", found.Notes)
		assert.NotNil(t, found.PurchaseDate)
		assert.NotNil(t, found.PurchasePrice)
		assert.Equal(t, 25000.0, *found.PurchasePrice)
		assert.NotNil(t, found.PurchaseMileage)
		assert.Equal(t, 10000, *found.PurchaseMileage)
		assert.NotNil(t, found.CurrentMileage)
		assert.Equal(t, 15000, *found.CurrentMileage)
	})
}
