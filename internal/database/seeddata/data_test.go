package seeddata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/models"
)

func TestGetSampleUsers(t *testing.T) {
	users := GetSampleUsers()

	// Verify we have at least 3 users
	assert.GreaterOrEqual(t, len(users), 3, "Should have at least 3 sample users")

	// Check each user has required fields
	for i, user := range users {
		assert.NotEmpty(t, user.ID, "User %d should have an ID", i)
		assert.NotEmpty(t, user.Username, "User %d should have a username", i)
		assert.NotEmpty(t, user.Email, "User %d should have an email", i)
		assert.NotEmpty(t, user.PasswordHash, "User %d should have a password hash", i)
		assert.NotEmpty(t, user.FirstName, "User %d should have a first name", i)
		assert.NotEmpty(t, user.LastName, "User %d should have a last name", i)
		assert.NotZero(t, user.CreatedAt, "User %d should have a created at timestamp", i)
		assert.NotZero(t, user.UpdatedAt, "User %d should have an updated at timestamp", i)
	}

	// Verify usernames are unique
	usernameMap := make(map[string]bool)
	for _, user := range users {
		assert.False(t, usernameMap[user.Username], "Username %s should be unique", user.Username)
		usernameMap[user.Username] = true
	}

	// Verify emails are unique
	emailMap := make(map[string]bool)
	for _, user := range users {
		assert.False(t, emailMap[user.Email], "Email %s should be unique", user.Email)
		emailMap[user.Email] = true
	}

	// Verify IDs are unique
	idMap := make(map[string]bool)
	for _, user := range users {
		assert.False(t, idMap[user.ID], "ID %s should be unique", user.ID)
		idMap[user.ID] = true
	}
}

func TestGetSampleVehicles(t *testing.T) {
	vehicles := GetSampleVehicles()
	users := GetSampleUsers()

	// Verify we have at least 5 vehicles
	assert.GreaterOrEqual(t, len(vehicles), 5, "Should have at least 5 sample vehicles")

	// Create a map of user IDs for validation
	userIDs := make(map[string]bool)
	for _, user := range users {
		userIDs[user.ID] = true
	}

	// Check each vehicle has required fields
	for i, vehicle := range vehicles {
		assert.NotEmpty(t, vehicle.ID, "Vehicle %d should have an ID", i)
		assert.NotEmpty(t, vehicle.UserID, "Vehicle %d should have a user ID", i)
		assert.True(t, userIDs[vehicle.UserID], "Vehicle %d should have a valid user ID", i)
		assert.NotEmpty(t, vehicle.VIN, "Vehicle %d should have a VIN", i)
		assert.Len(t, vehicle.VIN, 17, "Vehicle %d VIN should be 17 characters", i)
		assert.NotEmpty(t, vehicle.Make, "Vehicle %d should have a make", i)
		assert.NotEmpty(t, vehicle.Model, "Vehicle %d should have a model", i)
		assert.Greater(t, vehicle.Year, 1900, "Vehicle %d should have a valid year", i)
		assert.NotEmpty(t, vehicle.Color, "Vehicle %d should have a color", i)
		assert.NotZero(t, vehicle.CreatedAt, "Vehicle %d should have a created at timestamp", i)
		assert.NotZero(t, vehicle.UpdatedAt, "Vehicle %d should have an updated at timestamp", i)
	}

	// Verify VINs are unique
	vinMap := make(map[string]bool)
	for _, vehicle := range vehicles {
		assert.False(t, vinMap[vehicle.VIN], "VIN %s should be unique", vehicle.VIN)
		vinMap[vehicle.VIN] = true
	}

	// Verify IDs are unique
	idMap := make(map[string]bool)
	for _, vehicle := range vehicles {
		assert.False(t, idMap[vehicle.ID], "ID %s should be unique", vehicle.ID)
		idMap[vehicle.ID] = true
	}

	// Verify at least one of each status exists
	statusCounts := make(map[models.VehicleStatus]int)
	for _, vehicle := range vehicles {
		statusCounts[vehicle.Status]++
	}
	assert.Greater(t, statusCounts[models.VehicleStatusActive], 0, "Should have at least one active vehicle")
}

func TestGetSampleMaintenanceRecords(t *testing.T) {
	records := GetSampleMaintenanceRecords()
	vehicles := GetSampleVehicles()

	// Verify we have at least 8 maintenance records
	assert.GreaterOrEqual(t, len(records), 8, "Should have at least 8 sample maintenance records")

	// Create a map of vehicle IDs for validation
	vehicleIDs := make(map[string]bool)
	for _, vehicle := range vehicles {
		vehicleIDs[vehicle.ID] = true
	}

	// Check each maintenance record has required fields
	for i, record := range records {
		assert.NotEmpty(t, record.ID, "Record %d should have an ID", i)
		assert.NotEmpty(t, record.VehicleID, "Record %d should have a vehicle ID", i)
		assert.True(t, vehicleIDs[record.VehicleID], "Record %d should have a valid vehicle ID", i)
		assert.NotEmpty(t, record.ServiceType, "Record %d should have a service type", i)
		assert.NotZero(t, record.ServiceDate, "Record %d should have a service date", i)
		assert.NotZero(t, record.CreatedAt, "Record %d should have a created at timestamp", i)
		assert.NotZero(t, record.UpdatedAt, "Record %d should have an updated at timestamp", i)
	}

	// Verify IDs are unique
	idMap := make(map[string]bool)
	for _, record := range records {
		assert.False(t, idMap[record.ID], "ID %s should be unique", record.ID)
		idMap[record.ID] = true
	}

	// Verify there are multiple different service types
	serviceTypes := make(map[string]bool)
	for _, record := range records {
		serviceTypes[record.ServiceType] = true
	}
	assert.GreaterOrEqual(t, len(serviceTypes), 2, "Should have at least 2 different service types")
}

func TestSeedDataRelationships(t *testing.T) {
	users := GetSampleUsers()
	vehicles := GetSampleVehicles()
	records := GetSampleMaintenanceRecords()

	// Create maps for validation
	userIDs := make(map[string]bool)
	for _, user := range users {
		userIDs[user.ID] = true
	}

	vehicleIDs := make(map[string]bool)
	for _, vehicle := range vehicles {
		vehicleIDs[vehicle.ID] = true
	}

	// Verify all vehicles belong to valid users
	for _, vehicle := range vehicles {
		assert.True(t, userIDs[vehicle.UserID], "Vehicle %s should belong to a valid user", vehicle.ID)
	}

	// Verify all maintenance records belong to valid vehicles
	for _, record := range records {
		assert.True(t, vehicleIDs[record.VehicleID], "Maintenance record %s should belong to a valid vehicle", record.ID)
	}
}
