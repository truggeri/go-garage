package handlers

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

func TestBuildNewFuelRecord_AllOptionalFields_PopulatesRecord(t *testing.T) {
	input := map[string]interface{}{
		"fill_date":               "2024-02-05T14:30:00Z",
		"mileage":                 float64(47250),
		"volume":                  12.0,
		"fuel_type":               "gasoline",
		"partial_fill":            true,
		"price_per_unit":          4.25,
		"octane_rating":           float64(91),
		"location":                "Main St",
		"brand":                   "Shell",
		"notes":                   "road trip",
		"city_driving_percentage": float64(40),
		"vehicle_reported_mpg":    31.5,
	}

	rec, err := buildNewFuelRecord(input, "v1")

	require.NoError(t, err)
	assert.Equal(t, "v1", rec.VehicleID)
	assert.Equal(t, 47250, rec.Mileage)
	assert.Equal(t, 12.0, rec.Volume)
	assert.Equal(t, "gasoline", rec.FuelType)
	assert.True(t, rec.PartialFill)
	require.NotNil(t, rec.PricePerUnit)
	assert.Equal(t, 4.25, *rec.PricePerUnit)
	require.NotNil(t, rec.OctaneRating)
	assert.Equal(t, 91, *rec.OctaneRating)
	assert.Equal(t, "Main St", rec.Location)
	assert.Equal(t, "Shell", rec.Brand)
	assert.Equal(t, "road trip", rec.Notes)
	require.NotNil(t, rec.CityDrivingPercentage)
	assert.Equal(t, 40, *rec.CityDrivingPercentage)
	require.NotNil(t, rec.VehicleReportedMPG)
	assert.Equal(t, 31.5, *rec.VehicleReportedMPG)
}

func TestBuildNewFuelRecord_InvalidInput_ReturnsError(t *testing.T) {
	base := func() map[string]interface{} {
		return map[string]interface{}{
			"fill_date": "2024-02-05",
			"mileage":   float64(47250),
			"volume":    12.0,
			"fuel_type": "gasoline",
		}
	}

	tests := []struct {
		name      string
		mutate    func(map[string]interface{})
		wantField string
	}{
		{
			name:      "non-string fill date",
			mutate:    func(d map[string]interface{}) { d["fill_date"] = 12345 },
			wantField: "fill_date",
		},
		{
			name:      "fractional mileage",
			mutate:    func(d map[string]interface{}) { d["mileage"] = 47250.5 },
			wantField: "mileage",
		},
		{
			name:      "fractional octane rating",
			mutate:    func(d map[string]interface{}) { d["octane_rating"] = 91.5 },
			wantField: "octane_rating",
		},
		{
			name:      "fractional city driving percentage",
			mutate:    func(d map[string]interface{}) { d["city_driving_percentage"] = 40.5 },
			wantField: "city_driving_percentage",
		},
		{
			name:      "out of range mileage",
			mutate:    func(d map[string]interface{}) { d["mileage"] = math.Pow(2, 40) },
			wantField: "mileage",
		},
		{
			name: "price without usable volume",
			mutate: func(d map[string]interface{}) {
				d["volume"] = 0.0
				d["price"] = 60.0
			},
			wantField: "volume",
		},
		{
			name:      "invalid fuel type",
			mutate:    func(d map[string]interface{}) { d["fuel_type"] = "rocket" },
			wantField: "fuel_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := base()
			tt.mutate(input)

			rec, err := buildNewFuelRecord(input, "v1")

			assert.Nil(t, rec)
			require.Error(t, err)
			var valErr *models.ValidationError
			require.ErrorAs(t, err, &valErr)
			assert.Equal(t, tt.wantField, valErr.Field)
		})
	}
}

func TestExtractAndValidateFuelChanges_AllFields_PopulatesUpdates(t *testing.T) {
	existing := testFuelRecord("f1", "v1", time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC))
	input := map[string]interface{}{
		"fill_date":               "2024-03-01",
		"mileage":                 float64(47300),
		"volume":                  10.0,
		"fuel_type":               "diesel",
		"partial_fill":            true,
		"price":                   50.0,
		"octane_rating":           float64(87),
		"location":                "Elm St",
		"brand":                   "BP",
		"notes":                   "commute",
		"city_driving_percentage": float64(20),
		"vehicle_reported_mpg":    28.0,
	}

	updates, err := extractAndValidateFuelChanges(input, existing)

	require.NoError(t, err)
	require.NotNil(t, updates.FillDate)
	assert.Equal(t, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), *updates.FillDate)
	require.NotNil(t, updates.Mileage)
	assert.Equal(t, 47300, *updates.Mileage)
	require.NotNil(t, updates.Volume)
	assert.Equal(t, 10.0, *updates.Volume)
	require.NotNil(t, updates.FuelType)
	assert.Equal(t, "diesel", *updates.FuelType)
	require.NotNil(t, updates.PartialFill)
	assert.True(t, *updates.PartialFill)
	require.NotNil(t, updates.PricePerUnit)
	assert.Equal(t, 5.0, *updates.PricePerUnit)
	require.NotNil(t, updates.OctaneRating)
	assert.Equal(t, 87, *updates.OctaneRating)
	require.NotNil(t, updates.Location)
	assert.Equal(t, "Elm St", *updates.Location)
	require.NotNil(t, updates.Brand)
	assert.Equal(t, "BP", *updates.Brand)
	require.NotNil(t, updates.Notes)
	assert.Equal(t, "commute", *updates.Notes)
	require.NotNil(t, updates.CityDrivingPercentage)
	assert.Equal(t, 20, *updates.CityDrivingPercentage)
	require.NotNil(t, updates.VehicleReportedMPG)
	assert.Equal(t, 28.0, *updates.VehicleReportedMPG)
}

func TestExtractAndValidateFuelChanges_InvalidInput_ReturnsError(t *testing.T) {
	existing := testFuelRecord("f1", "v1", time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC))

	tests := []struct {
		name      string
		input     map[string]interface{}
		wantField string
	}{
		{
			name:      "invalid fill date",
			input:     map[string]interface{}{"fill_date": "not-a-date"},
			wantField: "fill_date",
		},
		{
			name:      "fractional mileage",
			input:     map[string]interface{}{"mileage": 47300.5},
			wantField: "mileage",
		},
		{
			name:      "fractional octane rating",
			input:     map[string]interface{}{"octane_rating": 91.5},
			wantField: "octane_rating",
		},
		{
			name:      "fractional city driving percentage",
			input:     map[string]interface{}{"city_driving_percentage": 20.5},
			wantField: "city_driving_percentage",
		},
		{
			name:      "price with zero volume update",
			input:     map[string]interface{}{"volume": 0.0, "price": 50.0},
			wantField: "volume",
		},
		{
			name:      "invalid fuel type",
			input:     map[string]interface{}{"fuel_type": "rocket"},
			wantField: "fuel_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractAndValidateFuelChanges(tt.input, existing)

			require.Error(t, err)
			var valErr *models.ValidationError
			require.ErrorAs(t, err, &valErr)
			assert.Equal(t, tt.wantField, valErr.Field)
		})
	}
}

func TestParseFuelDate(t *testing.T) {
	t.Run("parses RFC3339 value", func(t *testing.T) {
		parsed, err := parseFuelDate("2024-02-05T14:30:00Z")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC), parsed)
	})

	t.Run("parses date-only value", func(t *testing.T) {
		parsed, err := parseFuelDate("2024-02-05")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC), parsed)
	})

	t.Run("rejects empty value", func(t *testing.T) {
		_, err := parseFuelDate("")
		require.Error(t, err)
	})

	t.Run("rejects non-string value", func(t *testing.T) {
		_, err := parseFuelDate(42)
		require.Error(t, err)
	})
}

func TestExtractFuelPricePerUnit(t *testing.T) {
	t.Run("uses explicit price per unit", func(t *testing.T) {
		value, present, err := extractFuelPricePerUnit(map[string]interface{}{"price_per_unit": 3.5}, 10)
		require.NoError(t, err)
		assert.True(t, present)
		require.NotNil(t, value)
		assert.Equal(t, 3.5, *value)
	})

	t.Run("derives price per unit from total price", func(t *testing.T) {
		value, present, err := extractFuelPricePerUnit(map[string]interface{}{"price": 40.0}, 10)
		require.NoError(t, err)
		assert.True(t, present)
		require.NotNil(t, value)
		assert.Equal(t, 4.0, *value)
	})

	t.Run("returns absent when no price provided", func(t *testing.T) {
		value, present, err := extractFuelPricePerUnit(map[string]interface{}{}, 10)
		require.NoError(t, err)
		assert.False(t, present)
		assert.Nil(t, value)
	})

	t.Run("rejects total price with zero volume", func(t *testing.T) {
		_, _, err := extractFuelPricePerUnit(map[string]interface{}{"price": 40.0}, 0)
		require.Error(t, err)
	})
}

func TestIntFromFuelInput(t *testing.T) {
	t.Run("returns nil when key absent", func(t *testing.T) {
		value, err := intFromFuelInput(map[string]interface{}{}, "mileage")
		require.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("returns nil when value is not numeric", func(t *testing.T) {
		value, err := intFromFuelInput(map[string]interface{}{"mileage": "100"}, "mileage")
		require.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("returns integral value", func(t *testing.T) {
		value, err := intFromFuelInput(map[string]interface{}{"mileage": float64(100)}, "mileage")
		require.NoError(t, err)
		require.NotNil(t, value)
		assert.Equal(t, 100, *value)
	})

	t.Run("rejects fractional value", func(t *testing.T) {
		_, err := intFromFuelInput(map[string]interface{}{"mileage": 100.5}, "mileage")
		require.Error(t, err)
	})

	t.Run("rejects value below supported range", func(t *testing.T) {
		_, err := intFromFuelInput(map[string]interface{}{"mileage": -math.Pow(2, 40)}, "mileage")
		require.Error(t, err)
	})
}

func TestValidateFuelChanges_AppliesUpdatesBeforeValidation(t *testing.T) {
	existing := testFuelRecord("f1", "v1", time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC))
	fillDate := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	mileage := 47400
	volume := 9.0
	fuelType := "e85"
	partialFill := true
	pricePerUnit := 3.0
	octane := 87
	location := "Oak St"
	brand := "Exxon"
	notes := "topped off"
	cityPercentage := 55
	reportedMPG := 25.0

	err := validateFuelChanges(existing, services.FuelUpdates{
		FillDate: &fillDate, Mileage: &mileage, Volume: &volume, FuelType: &fuelType,
		PartialFill: &partialFill, PricePerUnit: &pricePerUnit, OctaneRating: &octane,
		Location: &location, Brand: &brand, Notes: &notes,
		CityDrivingPercentage: &cityPercentage, VehicleReportedMPG: &reportedMPG,
	})

	require.NoError(t, err)
	assert.Equal(t, 47250, existing.Mileage)
}

func TestValidateFuelChanges_InvalidUpdate_ReturnsError(t *testing.T) {
	existing := testFuelRecord("f1", "v1", time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC))
	octane := 0

	err := validateFuelChanges(existing, services.FuelUpdates{OctaneRating: &octane})

	require.Error(t, err)
}

func TestFuelToResponseMap_OptionalFields(t *testing.T) {
	fillDate := time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC)

	t.Run("omits unset optional fields", func(t *testing.T) {
		record := &models.FuelRecord{
			ID: "f1", VehicleID: "v1", FillDate: fillDate,
			Mileage: 47250, Volume: 12.0, FuelType: "gasoline",
		}

		response := fuelToResponseMap(record)

		assert.NotContains(t, response, "price")
		assert.NotContains(t, response, "price_per_unit")
		assert.NotContains(t, response, "octane_rating")
		assert.NotContains(t, response, "location")
		assert.NotContains(t, response, "brand")
		assert.NotContains(t, response, "notes")
		assert.NotContains(t, response, "city_driving_percentage")
		assert.NotContains(t, response, "vehicle_reported_mpg")
	})

	t.Run("includes set optional fields", func(t *testing.T) {
		pricePerUnit := 4.0
		octane := 91
		cityPercentage := 30
		reportedMPG := 32.5
		record := &models.FuelRecord{
			ID: "f1", VehicleID: "v1", FillDate: fillDate,
			Mileage: 47250, Volume: 12.0, FuelType: "gasoline",
			PricePerUnit: &pricePerUnit, OctaneRating: &octane,
			Location: "Main St", Brand: "Shell", Notes: "road trip",
			CityDrivingPercentage: &cityPercentage, VehicleReportedMPG: &reportedMPG,
		}

		response := fuelToResponseMap(record)

		assert.Equal(t, 4.0, response["price_per_unit"])
		assert.Equal(t, 48.0, response["price"])
		assert.Equal(t, 91, response["octane_rating"])
		assert.Equal(t, "Main St", response["location"])
		assert.Equal(t, "Shell", response["brand"])
		assert.Equal(t, "road trip", response["notes"])
		assert.Equal(t, 30, response["city_driving_percentage"])
		assert.Equal(t, 32.5, response["vehicle_reported_mpg"])
	})
}
