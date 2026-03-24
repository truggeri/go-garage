package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidFuelType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"gasoline is valid", "gasoline", true},
		{"diesel is valid", "diesel", true},
		{"e85 is valid", "e85", true},
		{"empty is invalid", "", false},
		{"unknown is invalid", "propane", false},
		{"uppercase is invalid", "Gasoline", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsValidFuelType(tt.input))
		})
	}
}

func TestFuelTypeDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"gasoline display name", "gasoline", "Gasoline"},
		{"diesel display name", "diesel", "Diesel"},
		{"e85 display name", "e85", "E85"},
		{"unknown returns raw value", "propane", "propane"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, FuelTypeDisplayName(tt.input))
		})
	}
}

func TestAllFuelTypes(t *testing.T) {
	types := AllFuelTypes()

	assert.Len(t, types, 3)
	assert.Equal(t, FuelTypeGasoline, types[0])
	assert.Equal(t, FuelTypeDiesel, types[1])
	assert.Equal(t, FuelTypeE85, types[2])

	// Verify it returns a copy, not the original slice
	types[0] = "modified"
	original := AllFuelTypes()
	assert.Equal(t, FuelTypeGasoline, original[0])
}
