package handlers

import (
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// buildNewFuelRecord creates a new FuelRecord model from the input data map and vehicle ID.
func buildNewFuelRecord(d map[string]interface{}, vehicleID string) (*models.FuelRecord, error) {
	fillDate, err := parseFuelDate(d["fill_date"])
	if err != nil {
		return nil, err
	}

	rec := &models.FuelRecord{
		VehicleID: vehicleID,
		FillDate:  fillDate,
		FuelType:  stringFromFuelInput(d, "fuel_type"),
	}
	if v, ok := d["mileage"].(float64); ok {
		rec.Mileage = int(v)
	}
	if v, ok := d["volume"].(float64); ok {
		rec.Volume = v
	}
	if v, ok := d["partial_fill"].(bool); ok {
		rec.PartialFill = v
	}
	pricePerUnit, hasPrice, priceErr := extractFuelPricePerUnit(d, rec.Volume)
	if priceErr != nil {
		return nil, priceErr
	}
	if hasPrice {
		rec.PricePerUnit = pricePerUnit
	}
	if v, ok := d["octane_rating"].(float64); ok {
		i := int(v)
		rec.OctaneRating = &i
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
	if v, ok := d["city_driving_percentage"].(float64); ok {
		i := int(v)
		rec.CityDrivingPercentage = &i
	}
	if v, ok := d["vehicle_reported_mpg"].(float64); ok {
		rec.VehicleReportedMPG = &v
	}

	if err := models.ValidateFuelRecord(rec); err != nil {
		return nil, err
	}

	return rec, nil
}

// extractAndValidateFuelChanges extracts fuel update fields and validates the resulting record.
func extractAndValidateFuelChanges(d map[string]interface{}, existing *models.FuelRecord) (services.FuelUpdates, error) {
	var u services.FuelUpdates
	if v, ok := d["fill_date"]; ok {
		t, err := parseFuelDate(v)
		if err != nil {
			return u, err
		}
		u.FillDate = &t
	}
	if v, ok := d["mileage"].(float64); ok {
		i := int(v)
		u.Mileage = &i
	}
	if v, ok := d["volume"].(float64); ok {
		u.Volume = &v
	}
	if v, ok := d["fuel_type"].(string); ok {
		u.FuelType = &v
	}
	if v, ok := d["partial_fill"].(bool); ok {
		u.PartialFill = &v
	}
	volume := existing.Volume
	if u.Volume != nil {
		volume = *u.Volume
	}
	pricePerUnit, hasPrice, priceErr := extractFuelPricePerUnit(d, volume)
	if priceErr != nil {
		return u, priceErr
	}
	if hasPrice {
		u.PricePerUnit = pricePerUnit
	}
	if v, ok := d["octane_rating"].(float64); ok {
		i := int(v)
		u.OctaneRating = &i
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
	if v, ok := d["city_driving_percentage"].(float64); ok {
		i := int(v)
		u.CityDrivingPercentage = &i
	}
	if v, ok := d["vehicle_reported_mpg"].(float64); ok {
		u.VehicleReportedMPG = &v
	}

	return u, validateFuelChanges(existing, u)
}
