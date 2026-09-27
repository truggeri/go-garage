package handlers

import (
	"math"
	"time"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// parseFuelDate parses RFC3339 or YYYY-MM-DD fuel date strings.
func parseFuelDate(v interface{}) (time.Time, error) {
	s, ok := v.(string)
	if !ok {
		return time.Time{}, models.NewValidationError("fill_date", "fill date must be a string in RFC3339 or YYYY-MM-DD format")
	}
	if s == "" {
		return time.Time{}, models.NewValidationError("fill_date", "fill date is required")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, models.NewValidationError("fill_date", "invalid date format, expected RFC3339 or YYYY-MM-DD")
	}
	return t, nil
}

// extractFuelPricePerUnit reads price_per_unit directly or derives it from total price and volume.
func extractFuelPricePerUnit(d map[string]interface{}, volume float64) (*float64, bool, error) {
	if v, ok := d["price_per_unit"].(float64); ok {
		return &v, true, nil
	}
	if v, ok := d["price"].(float64); ok {
		if volume <= 0 {
			return nil, false, models.NewValidationError("volume", "volume must be greater than zero when deriving price per unit from price")
		}
		pricePerUnit := v / volume
		return &pricePerUnit, true, nil
	}
	return nil, false, nil
}

// validateFuelChanges applies updates to a copy of the existing record and validates the result.
func validateFuelChanges(existing *models.FuelRecord, updates services.FuelUpdates) error {
	rec := *existing
	if updates.FillDate != nil {
		rec.FillDate = *updates.FillDate
	}
	if updates.Mileage != nil {
		rec.Mileage = *updates.Mileage
	}
	if updates.Volume != nil {
		rec.Volume = *updates.Volume
	}
	if updates.FuelType != nil {
		rec.FuelType = *updates.FuelType
	}
	if updates.PartialFill != nil {
		rec.PartialFill = *updates.PartialFill
	}
	if updates.PricePerUnit != nil {
		rec.PricePerUnit = updates.PricePerUnit
	}
	if updates.OctaneRating != nil {
		rec.OctaneRating = updates.OctaneRating
	}
	if updates.CityDrivingPercentage != nil {
		rec.CityDrivingPercentage = updates.CityDrivingPercentage
	}
	if updates.VehicleReportedMPG != nil {
		rec.VehicleReportedMPG = updates.VehicleReportedMPG
	}
	if updates.Location != nil {
		rec.Location = *updates.Location
	}
	if updates.Brand != nil {
		rec.Brand = *updates.Brand
	}
	if updates.Notes != nil {
		rec.Notes = *updates.Notes
	}
	return models.ValidateFuelRecord(&rec)
}

// intFromFuelInput returns an integer field value, rejecting JSON numbers with a fractional part.
func intFromFuelInput(d map[string]interface{}, key string) (*int, error) {
	v, ok := d[key].(float64)
	if !ok {
		return nil, nil
	}
	if v != math.Trunc(v) {
		return nil, models.NewValidationError(key, key+" must be a whole number")
	}
	i := int(v)
	return &i, nil
}

// stringFromFuelInput returns a string field value or an empty string when absent or not a string.
func stringFromFuelInput(d map[string]interface{}, key string) string {
	v, _ := d[key].(string)
	return v
}
