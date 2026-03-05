package handlers

import (
	"time"

	"github.com/truggeri/go-garage/internal/models"
)

// fuelToResponseMap converts a FuelRecord model to a response map for JSON encoding.
func fuelToResponseMap(f *models.FuelRecord) map[string]interface{} {
	r := map[string]interface{}{
		"id":            f.ID,
		"vehicle_id":    f.VehicleID,
		"fill_date":     f.FillDate.Format("2006-01-02"),
		"odometer":      f.Odometer,
		"cost_per_unit": f.CostPerUnit,
		"volume":        f.Volume,
		"partial":       f.PartialFuelUp,
		"created_at":    f.CreatedAt.Format(time.RFC3339),
		"updated_at":    f.UpdatedAt.Format(time.RFC3339),
	}
	if f.FuelType != "" {
		r["fuel_type"] = f.FuelType
	}
	if f.CityDrivingPct != nil {
		r["city_driving_pct"] = *f.CityDrivingPct
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
	if f.ReportedMPG != nil {
		r["reported_mpg"] = *f.ReportedMPG
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
		"success": true, "data": items,
		"pagination": map[string]int{"page": pg, "limit": sz, "total": total, "total_pages": tp},
	}
}

// buildFuelSinglePayload creates a single fuel record response payload.
func buildFuelSinglePayload(f *models.FuelRecord, msg string) map[string]interface{} {
	p := map[string]interface{}{"success": true, "data": fuelToResponseMap(f)}
	if msg != "" {
		p["message"] = msg
	}
	return p
}
