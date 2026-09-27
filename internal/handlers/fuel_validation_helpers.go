package handlers

import (
	"time"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

func parseFuelDate(v interface{}) (time.Time, error) {
	s, _ := v.(string)
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

func extractFuelPricePerUnit(d map[string]interface{}, volume float64) *float64 {
	if v, ok := d["price_per_unit"].(float64); ok {
		return &v
	}
	if v, ok := d["price"].(float64); ok && volume > 0 {
		pricePerUnit := v / volume
		return &pricePerUnit
	}
	return nil
}

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
	return models.ValidateFuelRecord(&rec)
}

func stringValue(d map[string]interface{}, key string) string {
	v, _ := d[key].(string)
	return v
}
