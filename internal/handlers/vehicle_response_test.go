package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
)

func TestVehicleToResponseMap(t *testing.T) {
	baseTime := time.Date(2020, 6, 15, 10, 30, 0, 0, time.UTC)

	t.Run("includes required fields", func(t *testing.T) {
		vehicle := &models.Vehicle{
			ID:        "v123",
			UserID:    "u456",
			VIN:       "ABC12345678901234",
			Make:      "Ford",
			Model:     "Focus",
			Year:      2020,
			Status:    models.VehicleStatusActive,
			CreatedAt: baseTime,
			UpdatedAt: baseTime,
		}
		m := vehicleToResponseMap(vehicle)
		assert.Equal(t, "v123", m["id"])
		assert.Equal(t, "u456", m["user_id"])
		assert.Equal(t, "ABC12345678901234", m["vin"])
		assert.Equal(t, "Ford", m["make"])
		assert.Equal(t, "Focus", m["model"])
		assert.Equal(t, 2020, m["year"])
		assert.Equal(t, "active", m["status"])
		assert.Contains(t, m["created_at"], "2020-06-15")
		assert.Contains(t, m["updated_at"], "2020-06-15")
	})

	t.Run("includes optional fields when present", func(t *testing.T) {
		purchaseDate := time.Date(2019, 3, 10, 0, 0, 0, 0, time.UTC)
		purchasePrice := 25000.00
		purchaseMileage := 5000
		currentMileage := 15000

		vehicle := &models.Vehicle{
			ID:              "v123",
			UserID:          "u456",
			VIN:             "ABC12345678901234",
			Make:            "Ford",
			Model:           "Focus",
			Year:            2020,
			Status:          models.VehicleStatusActive,
			DisplayName:     "My Daily Driver",
			Color:           "Blue",
			LicensePlate:    "ABC-1234",
			PurchaseDate:    &purchaseDate,
			PurchasePrice:   &purchasePrice,
			PurchaseMileage: &purchaseMileage,
			CurrentMileage:  &currentMileage,
			Notes:           "Test notes",
			CreatedAt:       baseTime,
			UpdatedAt:       baseTime,
		}
		m := vehicleToResponseMap(vehicle)
		assert.Equal(t, "My Daily Driver", m["display_name"])
		assert.Equal(t, "Blue", m["color"])
		assert.Equal(t, "ABC-1234", m["license_plate"])
		assert.Equal(t, "2019-03-10", m["purchase_date"])
		assert.Equal(t, 25000.00, m["purchase_price"])
		assert.Equal(t, 5000, m["purchase_mileage"])
		assert.Equal(t, 15000, m["current_mileage"])
		assert.Equal(t, "Test notes", m["notes"])
	})

	t.Run("excludes optional fields when empty", func(t *testing.T) {
		vehicle := &models.Vehicle{
			ID:        "v123",
			UserID:    "u456",
			VIN:       "ABC12345678901234",
			Make:      "Ford",
			Model:     "Focus",
			Year:      2020,
			Status:    models.VehicleStatusActive,
			CreatedAt: baseTime,
			UpdatedAt: baseTime,
		}
		m := vehicleToResponseMap(vehicle)
		_, hasDisplayName := m["display_name"]
		_, hasColor := m["color"]
		_, hasLicensePlate := m["license_plate"]
		_, hasPurchaseDate := m["purchase_date"]
		_, hasPurchasePrice := m["purchase_price"]
		_, hasPurchaseMileage := m["purchase_mileage"]
		_, hasCurrentMileage := m["current_mileage"]
		_, hasNotes := m["notes"]

		assert.False(t, hasDisplayName)
		assert.False(t, hasColor)
		assert.False(t, hasLicensePlate)
		assert.False(t, hasPurchaseDate)
		assert.False(t, hasPurchasePrice)
		assert.False(t, hasPurchaseMileage)
		assert.False(t, hasCurrentMileage)
		assert.False(t, hasNotes)
	})
}

func TestBuildListPayload(t *testing.T) {
	baseTime := time.Date(2020, 6, 15, 10, 30, 0, 0, time.UTC)

	t.Run("builds payload with vehicles", func(t *testing.T) {
		vehicles := []*models.Vehicle{
			{ID: "v1", UserID: "u1", VIN: "VIN1", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive, CreatedAt: baseTime, UpdatedAt: baseTime},
			{ID: "v2", UserID: "u1", VIN: "VIN2", Make: "Honda", Model: "Civic", Year: 2021, Status: models.VehicleStatusActive, CreatedAt: baseTime, UpdatedAt: baseTime},
		}
		payload := buildListPayload(vehicles, 1, 20, 2)

		assert.True(t, payload["success"].(bool))
		data := payload["data"].([]map[string]interface{})
		assert.Len(t, data, 2)
		assert.Equal(t, "v1", data[0]["id"])
		assert.Equal(t, "v2", data[1]["id"])
	})

	t.Run("includes pagination info", func(t *testing.T) {
		vehicles := []*models.Vehicle{}
		payload := buildListPayload(vehicles, 2, 10, 25)

		pagination := payload["pagination"].(map[string]int)
		assert.Equal(t, 2, pagination["page"])
		assert.Equal(t, 10, pagination["limit"])
		assert.Equal(t, 25, pagination["total"])
		assert.Equal(t, 3, pagination["total_pages"]) // 25/10 = 2.5, rounded up to 3
	})

	t.Run("calculates total pages correctly", func(t *testing.T) {
		// Exact division
		payload := buildListPayload([]*models.Vehicle{}, 1, 10, 30)
		assert.Equal(t, 3, payload["pagination"].(map[string]int)["total_pages"])

		// With remainder
		payload = buildListPayload([]*models.Vehicle{}, 1, 10, 31)
		assert.Equal(t, 4, payload["pagination"].(map[string]int)["total_pages"])

		// Zero total
		payload = buildListPayload([]*models.Vehicle{}, 1, 10, 0)
		assert.Equal(t, 0, payload["pagination"].(map[string]int)["total_pages"])
	})
}

func TestBuildSinglePayload(t *testing.T) {
	baseTime := time.Date(2020, 6, 15, 10, 30, 0, 0, time.UTC)
	vehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", VIN: "VIN123", Make: "Ford", Model: "Focus",
		Year: 2020, Status: models.VehicleStatusActive, CreatedAt: baseTime, UpdatedAt: baseTime,
	}

	t.Run("builds payload with message", func(t *testing.T) {
		payload := buildSinglePayload(vehicle, "Vehicle created successfully")

		assert.True(t, payload["success"].(bool))
		assert.Equal(t, "Vehicle created successfully", payload["message"])
		data := payload["data"].(map[string]interface{})
		assert.Equal(t, "v1", data["id"])
	})

	t.Run("builds payload without message when empty", func(t *testing.T) {
		payload := buildSinglePayload(vehicle, "")

		assert.True(t, payload["success"].(bool))
		_, hasMessage := payload["message"]
		assert.False(t, hasMessage)
	})
}

func TestHandleDomainError(t *testing.T) {
	t.Run("handles validation error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := models.NewValidationError("vin", "Invalid VIN format")

		handleDomainError(rec, err)

		assert.Equal(t, 400, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.False(t, resp["success"].(bool))
		errObj := resp["error"].(map[string]interface{})
		assert.Equal(t, "VALIDATION_ERROR", errObj["code"])
	})

	t.Run("handles duplicate error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := models.NewDuplicateError("Vehicle", "vin", "ABC123")

		handleDomainError(rec, err)

		assert.Equal(t, 409, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		errObj := resp["error"].(map[string]interface{})
		assert.Equal(t, "DUPLICATE_ERROR", errObj["code"])
	})

	t.Run("handles not found error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := models.NewNotFoundError("Vehicle", "v999")

		handleDomainError(rec, err)

		assert.Equal(t, 404, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		errObj := resp["error"].(map[string]interface{})
		assert.Equal(t, "NOT_FOUND", errObj["code"])
	})
}
