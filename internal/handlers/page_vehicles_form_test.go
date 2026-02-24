package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVehicleNewForm(t *testing.T) {
	t.Run("returns no errors for valid required fields", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "", "", "")
		assert.Empty(t, result.Errors)
		assert.Equal(t, 2021, result.Year)
	})

	t.Run("returns zero year when year is empty", func(t *testing.T) {
		result := parseVehicleNewForm("", "", "", "", "")
		assert.NotContains(t, result.Errors, "year")
		assert.Equal(t, 0, result.Year)
	})

	t.Run("returns error when year is not a number", func(t *testing.T) {
		result := parseVehicleNewForm("abc", "", "", "", "")
		assert.Contains(t, result.Errors, "year")
		assert.Equal(t, "Year must be a valid number", result.Errors["year"])
	})

	t.Run("parses valid purchase date", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "2021-06-15", "", "", "")
		assert.Empty(t, result.Errors)
		require.NotNil(t, result.PurchaseDate)
		assert.Equal(t, 2021, result.PurchaseDate.Year())
		assert.Equal(t, 6, int(result.PurchaseDate.Month()))
		assert.Equal(t, 15, result.PurchaseDate.Day())
	})

	t.Run("returns error for invalid purchase date", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "not-a-date", "", "", "")
		assert.Contains(t, result.Errors, "purchase_date")
		assert.Equal(t, "Invalid date format", result.Errors["purchase_date"])
	})

	t.Run("leaves purchase date nil when empty", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "", "", "")
		assert.Nil(t, result.PurchaseDate)
	})

	t.Run("parses valid purchase price", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "25000.50", "", "")
		assert.Empty(t, result.Errors)
		require.NotNil(t, result.PurchasePrice)
		assert.Equal(t, 25000.50, *result.PurchasePrice)
	})

	t.Run("returns error for non-numeric purchase price", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "abc", "", "")
		assert.Contains(t, result.Errors, "purchase_price")
		assert.Equal(t, "Purchase price must be a valid number", result.Errors["purchase_price"])
	})

	t.Run("parses negative purchase price without error", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "-1", "", "")
		assert.NotContains(t, result.Errors, "purchase_price")
		require.NotNil(t, result.PurchasePrice)
		assert.Equal(t, -1.0, *result.PurchasePrice)
	})

	t.Run("parses valid purchase mileage", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "", "5000", "")
		assert.Empty(t, result.Errors)
		require.NotNil(t, result.PurchaseMileage)
		assert.Equal(t, 5000, *result.PurchaseMileage)
	})

	t.Run("returns error for non-numeric purchase mileage", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "", "abc", "")
		assert.Contains(t, result.Errors, "purchase_mileage")
		assert.Equal(t, "Mileage at purchase must be a valid number", result.Errors["purchase_mileage"])
	})

	t.Run("parses valid current mileage", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "", "", "12000")
		assert.Empty(t, result.Errors)
		require.NotNil(t, result.CurrentMileage)
		assert.Equal(t, 12000, *result.CurrentMileage)
	})

	t.Run("returns error for non-numeric current mileage", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "", "", "xyz")
		assert.Contains(t, result.Errors, "current_mileage")
		assert.Equal(t, "Current mileage must be a valid number", result.Errors["current_mileage"])
	})

	t.Run("accepts zero purchase price", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "0", "", "")
		assert.NotContains(t, result.Errors, "purchase_price")
		require.NotNil(t, result.PurchasePrice)
		assert.Equal(t, 0.0, *result.PurchasePrice)
	})

	t.Run("accepts zero purchase mileage", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "", "0", "")
		assert.NotContains(t, result.Errors, "purchase_mileage")
		require.NotNil(t, result.PurchaseMileage)
		assert.Equal(t, 0, *result.PurchaseMileage)
	})

	t.Run("accepts zero current mileage", func(t *testing.T) {
		result := parseVehicleNewForm("2021", "", "", "", "0")
		assert.NotContains(t, result.Errors, "current_mileage")
		require.NotNil(t, result.CurrentMileage)
		assert.Equal(t, 0, *result.CurrentMileage)
	})
}
