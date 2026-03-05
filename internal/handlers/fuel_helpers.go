package handlers

import (
	"time"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// buildNewFuelRecord creates a new FuelRecord model from the input data map and vehicle ID.
func buildNewFuelRecord(d map[string]interface{}, vehicleID string) (*models.FuelRecord, error) {
	fillDateStr, _ := d["fill_date"].(string)
	fillDate, err := time.Parse("2006-01-02", fillDateStr)
	if err != nil {
		return nil, models.NewValidationError("fill_date", "invalid date format, expected YYYY-MM-DD")
	}

	odometer := 0
	if v, ok := d["odometer"].(float64); ok {
		odometer = int(v)
	}

	costPerUnit := 0.0
	if v, ok := d["cost_per_unit"].(float64); ok {
		costPerUnit = v
	}

	volume := 0.0
	if v, ok := d["volume"].(float64); ok {
		volume = v
	}

	rec := &models.FuelRecord{
		VehicleID:   vehicleID,
		FillDate:    fillDate,
		Odometer:    odometer,
		CostPerUnit: costPerUnit,
		Volume:      volume,
	}

	if v, ok := d["fuel_type"].(string); ok {
		rec.FuelType = v
	}
	if v, ok := d["city_driving_pct"].(float64); ok {
		pct := int(v)
		rec.CityDrivingPct = &pct
	}
	if v, ok := d["location"].(string); ok {
		rec.Location = v
	}
	if v, ok := d["brand"].(string); ok {
		rec.Brand = v
	}
	if v, ok := d["notes"].(string); ok {
		rec.Notes = v
	}
	if v, ok := d["reported_mpg"].(float64); ok {
		rec.ReportedMPG = &v
	}
	if v, ok := d["partial"].(bool); ok {
		rec.PartialFuelUp = v
	}

	return rec, nil
}

// extractFuelChanges extracts fuel update fields from the input data map.
func extractFuelChanges(d map[string]interface{}) services.FuelUpdates {
	var u services.FuelUpdates

	if v, ok := d["fill_date"].(string); ok && v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err == nil {
			u.FillDate = &t
		}
	}
	if v, ok := d["odometer"].(float64); ok {
		i := int(v)
		u.Odometer = &i
	}
	if v, ok := d["cost_per_unit"].(float64); ok {
		u.CostPerUnit = &v
	}
	if v, ok := d["volume"].(float64); ok {
		u.Volume = &v
	}
	if v, ok := d["fuel_type"].(string); ok {
		u.FuelType = &v
	}
	if v, ok := d["city_driving_pct"].(float64); ok {
		pct := int(v)
		u.CityDrivingPct = &pct
	}
	if v, ok := d["location"].(string); ok {
		u.Location = &v
	}
	if v, ok := d["brand"].(string); ok {
		u.Brand = &v
	}
	if v, ok := d["notes"].(string); ok {
		u.Notes = &v
	}
	if v, ok := d["reported_mpg"].(float64); ok {
		u.ReportedMPG = &v
	}
	if v, ok := d["partial"].(bool); ok {
		u.PartialFuelUp = &v
	}

	return u
}
