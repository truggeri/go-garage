package handlers

import (
	"time"

	"github.com/truggeri/go-garage/internal/models"
)

// fuelToResponseMap converts a FuelRecord model to a response map for JSON encoding.
func fuelToResponseMap(f *models.FuelRecord) map[string]interface{} {
	r := map[string]interface{}{
		"id":           f.ID,
		"vehicle_id":   f.VehicleID,
		"fill_date":    f.FillDate.Format(time.RFC3339),
		"mileage":      f.Mileage,
		"volume":       f.Volume,
		"fuel_type":    f.FuelType,
		"partial_fill": f.PartialFill,
		"created_at":   f.CreatedAt.Format(time.RFC3339),
		"updated_at":   f.UpdatedAt.Format(time.RFC3339),
	}
	if f.PricePerUnit != nil {
		r["price_per_unit"] = *f.PricePerUnit
		r["price"] = *f.PricePerUnit * f.Volume
	}
	if f.OctaneRating != nil {
		r["octane_rating"] = *f.OctaneRating
	}
	if f.Location != "" {
		r["location"] = f.Location
	}
	if f.Brand != "" {
		r["brand"] = f.Brand
	}
	if f.Notes != "" {
		r["notes"] = f.Notes
	}
	if f.CityDrivingPercentage != nil {
		r["city_driving_percentage"] = *f.CityDrivingPercentage
	}
	if f.VehicleReportedMPG != nil {
		r["vehicle_reported_mpg"] = *f.VehicleReportedMPG
	}
	return r
}

// buildFuelListPayload creates a paginated list response payload for fuel records.
func buildFuelListPayload(recs []*models.FuelRecord, pg, sz, total int) map[string]interface{} {
	items := make([]map[string]interface{}, len(recs))
	for i, f := range recs {
		items[i] = fuelToResponseMap(f)
	}
	tp := 0
	if total > 0 && sz > 0 {
		tp = total / sz
		if total%sz != 0 {
			tp++
		}
	}
	return map[string]interface{}{
		keySuccess: true, "data": items,
		"pagination": map[string]int{"page": pg, "limit": sz, "total": total, "total_pages": tp},
	}
}

// buildFuelSinglePayload creates a single fuel record response payload.
func buildFuelSinglePayload(f *models.FuelRecord, msg string) map[string]interface{} {
	p := map[string]interface{}{keySuccess: true, "data": fuelToResponseMap(f)}
	if msg != "" {
		p["message"] = msg
	}
	return p
}
