package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
)

func TestExtractPaging(t *testing.T) {
	t.Run("returns defaults when no query params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		page, limit := extractPaging(req)
		assert.Equal(t, 1, page)
		assert.Equal(t, 20, limit)
	})

	t.Run("extracts valid page and limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?page=3&limit=50", nil)
		page, limit := extractPaging(req)
		assert.Equal(t, 3, page)
		assert.Equal(t, 50, limit)
	})

	t.Run("ignores invalid page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?page=-1", nil)
		page, _ := extractPaging(req)
		assert.Equal(t, 1, page)
	})

	t.Run("ignores invalid limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?limit=200", nil)
		_, limit := extractPaging(req)
		assert.Equal(t, 20, limit)
	})

	t.Run("ignores non-numeric page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?page=abc", nil)
		page, _ := extractPaging(req)
		assert.Equal(t, 1, page)
	})
}

func TestBuildVehicleFilterSpec(t *testing.T) {
	t.Run("always includes owner ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		filters := buildVehicleFilterSpec(req, "user-123")
		require.NotNil(t, filters.UserID)
		assert.Equal(t, "user-123", *filters.UserID)
	})

	t.Run("extracts make filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?make=Ford", nil)
		filters := buildVehicleFilterSpec(req, "user-123")
		require.NotNil(t, filters.Make)
		assert.Equal(t, "Ford", *filters.Make)
	})

	t.Run("extracts model filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?model=Focus", nil)
		filters := buildVehicleFilterSpec(req, "user-123")
		require.NotNil(t, filters.Model)
		assert.Equal(t, "Focus", *filters.Model)
	})

	t.Run("extracts year filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?year=2020", nil)
		filters := buildVehicleFilterSpec(req, "user-123")
		require.NotNil(t, filters.Year)
		assert.Equal(t, 2020, *filters.Year)
	})

	t.Run("extracts status filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?status=active", nil)
		filters := buildVehicleFilterSpec(req, "user-123")
		require.NotNil(t, filters.Status)
		assert.Equal(t, models.VehicleStatus("active"), *filters.Status)
	})

	t.Run("ignores invalid year", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles?year=abc", nil)
		filters := buildVehicleFilterSpec(req, "user-123")
		assert.Nil(t, filters.Year)
	})
}

func TestParseJSONBody(t *testing.T) {
	t.Run("parses valid JSON", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{"name": "test", "value": 123}`))
		req := httptest.NewRequest(http.MethodPost, "/", body)
		data, err := parseJSONBody(req)
		require.NoError(t, err)
		assert.Equal(t, "test", data["name"])
		assert.Equal(t, float64(123), data["value"])
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{invalid json}`))
		req := httptest.NewRequest(http.MethodPost, "/", body)
		_, err := parseJSONBody(req)
		require.Error(t, err)
	})
}

func TestValidateRequiredKeys(t *testing.T) {
	t.Run("returns empty for all present keys", func(t *testing.T) {
		data := map[string]interface{}{
			"vin": "ABC123", "make": "Ford", "model": "Focus", "year": float64(2020),
		}
		errs := validateRequiredKeys(data, "vin", "make", "model", "year")
		assert.Empty(t, errs)
	})

	t.Run("returns errors for missing keys", func(t *testing.T) {
		data := map[string]interface{}{"make": "Ford"}
		errs := validateRequiredKeys(data, "vin", "make", "model")
		assert.Len(t, errs, 2)
		fields := []string{}
		for _, e := range errs {
			fields = append(fields, e.Field)
		}
		assert.Contains(t, fields, "vin")
		assert.Contains(t, fields, "model")
	})

	t.Run("returns error for empty string", func(t *testing.T) {
		data := map[string]interface{}{"vin": ""}
		errs := validateRequiredKeys(data, "vin")
		assert.Len(t, errs, 1)
		assert.Equal(t, "vin", errs[0].Field)
	})

	t.Run("returns error for zero year", func(t *testing.T) {
		data := map[string]interface{}{"year": float64(0)}
		errs := validateRequiredKeys(data, "year")
		assert.Len(t, errs, 1)
		assert.Equal(t, "year", errs[0].Field)
	})
}

func TestBuildNewVehicleRecord(t *testing.T) {
	t.Run("builds vehicle with required fields", func(t *testing.T) {
		data := map[string]interface{}{
			"vin": "abc123", "make": "Honda", "model": "Civic", "year": float64(2021),
		}
		vehicle, err := buildNewVehicleRecord(data, "owner-id")
		require.NoError(t, err)
		assert.Equal(t, "owner-id", vehicle.UserID)
		assert.Equal(t, "ABC123", vehicle.VIN) // should be uppercased
		assert.Equal(t, "Honda", vehicle.Make)
		assert.Equal(t, "Civic", vehicle.Model)
		assert.Equal(t, 2021, vehicle.Year)
		assert.Equal(t, models.VehicleStatusActive, vehicle.Status)
	})

	t.Run("builds vehicle with optional fields", func(t *testing.T) {
		data := map[string]interface{}{
			"vin": "ABC123", "make": "Ford", "model": "Focus", "year": float64(2020),
			"color": "Blue", "license_plate": "ABC-1234", "notes": "Test notes",
			"purchase_price": float64(25000.00), "purchase_mileage": float64(5000),
			"current_mileage": float64(10000),
		}
		vehicle, err := buildNewVehicleRecord(data, "owner-id")
		require.NoError(t, err)
		assert.Equal(t, "Blue", vehicle.Color)
		assert.Equal(t, "ABC-1234", vehicle.LicensePlate)
		assert.Equal(t, "Test notes", vehicle.Notes)
		require.NotNil(t, vehicle.PurchasePrice)
		assert.Equal(t, 25000.00, *vehicle.PurchasePrice)
		require.NotNil(t, vehicle.PurchaseMileage)
		assert.Equal(t, 5000, *vehicle.PurchaseMileage)
		require.NotNil(t, vehicle.CurrentMileage)
		assert.Equal(t, 10000, *vehicle.CurrentMileage)
	})

	t.Run("builds vehicle with custom status", func(t *testing.T) {
		data := map[string]interface{}{
			"vin": "ABC123", "make": "Ford", "model": "Focus", "year": float64(2020),
			"status": "sold",
		}
		vehicle, err := buildNewVehicleRecord(data, "owner-id")
		require.NoError(t, err)
		assert.Equal(t, models.VehicleStatus("sold"), vehicle.Status)
	})

	t.Run("parses valid purchase date", func(t *testing.T) {
		data := map[string]interface{}{
			"vin": "ABC123", "make": "Ford", "model": "Focus", "year": float64(2020),
			"purchase_date": "2020-06-15",
		}
		vehicle, err := buildNewVehicleRecord(data, "owner-id")
		require.NoError(t, err)
		require.NotNil(t, vehicle.PurchaseDate)
		assert.Equal(t, 2020, vehicle.PurchaseDate.Year())
		assert.Equal(t, 6, int(vehicle.PurchaseDate.Month()))
		assert.Equal(t, 15, vehicle.PurchaseDate.Day())
	})

	t.Run("returns error for invalid purchase date", func(t *testing.T) {
		data := map[string]interface{}{
			"vin": "ABC123", "make": "Ford", "model": "Focus", "year": float64(2020),
			"purchase_date": "invalid-date",
		}
		_, err := buildNewVehicleRecord(data, "owner-id")
		require.Error(t, err)
	})
}

func TestExtractVehicleChanges(t *testing.T) {
	t.Run("extracts VIN change with uppercase", func(t *testing.T) {
		data := map[string]interface{}{"vin": " abc123 "}
		changes := extractVehicleChanges(data)
		require.NotNil(t, changes.VIN)
		assert.Equal(t, "ABC123", *changes.VIN)
	})

	t.Run("extracts make change", func(t *testing.T) {
		data := map[string]interface{}{"make": "Toyota"}
		changes := extractVehicleChanges(data)
		require.NotNil(t, changes.Make)
		assert.Equal(t, "Toyota", *changes.Make)
	})

	t.Run("extracts model change", func(t *testing.T) {
		data := map[string]interface{}{"model": "Camry"}
		changes := extractVehicleChanges(data)
		require.NotNil(t, changes.Model)
		assert.Equal(t, "Camry", *changes.Model)
	})

	t.Run("extracts year change", func(t *testing.T) {
		data := map[string]interface{}{"year": float64(2022)}
		changes := extractVehicleChanges(data)
		require.NotNil(t, changes.Year)
		assert.Equal(t, 2022, *changes.Year)
	})

	t.Run("extracts color change", func(t *testing.T) {
		data := map[string]interface{}{"color": "Red"}
		changes := extractVehicleChanges(data)
		require.NotNil(t, changes.Color)
		assert.Equal(t, "Red", *changes.Color)
	})

	t.Run("extracts license plate change", func(t *testing.T) {
		data := map[string]interface{}{"license_plate": "XYZ-789"}
		changes := extractVehicleChanges(data)
		require.NotNil(t, changes.LicensePlate)
		assert.Equal(t, "XYZ-789", *changes.LicensePlate)
	})

	t.Run("extracts current mileage change", func(t *testing.T) {
		data := map[string]interface{}{"current_mileage": float64(15000)}
		changes := extractVehicleChanges(data)
		require.NotNil(t, changes.CurrentMileage)
		assert.Equal(t, 15000, *changes.CurrentMileage)
	})

	t.Run("extracts notes change", func(t *testing.T) {
		data := map[string]interface{}{"notes": "Updated notes"}
		changes := extractVehicleChanges(data)
		require.NotNil(t, changes.Notes)
		assert.Equal(t, "Updated notes", *changes.Notes)
	})

	t.Run("ignores empty strings", func(t *testing.T) {
		data := map[string]interface{}{"make": "", "model": ""}
		changes := extractVehicleChanges(data)
		assert.Nil(t, changes.Make)
		assert.Nil(t, changes.Model)
	})

	t.Run("ignores zero year", func(t *testing.T) {
		data := map[string]interface{}{"year": float64(0)}
		changes := extractVehicleChanges(data)
		assert.Nil(t, changes.Year)
	})
}
