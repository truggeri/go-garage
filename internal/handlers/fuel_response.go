package handlers

import (
	"time"

	"github.com/truggeri/go-garage/internal/models"
)

// fuelToResponseMap converts a FuelRecord model to a response map for JSON encoding.
func fuelToResponseMap(record *models.FuelRecord) map[string]interface{} {
	response := map[string]interface{}{
		"id":           record.ID,
		"vehicle_id":   record.VehicleID,
		"fill_date":    record.FillDate.Format(time.RFC3339),
		"mileage":      record.Mileage,
		"volume":       record.Volume,
		"fuel_type":    record.FuelType,
		"partial_fill": record.PartialFill,
		"created_at":   record.CreatedAt.Format(time.RFC3339),
		"updated_at":   record.UpdatedAt.Format(time.RFC3339),
	}
	if record.PricePerUnit != nil {
		response["price_per_unit"] = *record.PricePerUnit
		response["price"] = *record.PricePerUnit * record.Volume
	}
	if record.OctaneRating != nil {
		response["octane_rating"] = *record.OctaneRating
	}
	if record.Location != "" {
		response["location"] = record.Location
	}
	if record.Brand != "" {
		response["brand"] = record.Brand
	}
	if record.Notes != "" {
		response["notes"] = record.Notes
	}
	if record.CityDrivingPercentage != nil {
		response["city_driving_percentage"] = *record.CityDrivingPercentage
	}
	if record.VehicleReportedMPG != nil {
		response["vehicle_reported_mpg"] = *record.VehicleReportedMPG
	}
	return response
}

// buildFuelListPayload creates a paginated list response payload for fuel records.
func buildFuelListPayload(records []*models.FuelRecord, page, pageSize, total int) map[string]interface{} {
	items := make([]map[string]interface{}, len(records))
	for idx, record := range records {
		items[idx] = fuelToResponseMap(record)
	}
	totalPages := 0
	if total > 0 && pageSize > 0 {
		totalPages = total / pageSize
		if total%pageSize != 0 {
			totalPages++
		}
	}
	return map[string]interface{}{
		keySuccess: true, "data": items,
		"pagination": map[string]int{"page": page, "limit": pageSize, "total": total, "total_pages": totalPages},
	}
}

// buildFuelSinglePayload creates a single fuel record response payload.
func buildFuelSinglePayload(record *models.FuelRecord, message string) map[string]interface{} {
	payload := map[string]interface{}{keySuccess: true, "data": fuelToResponseMap(record)}
	if message != "" {
		payload["message"] = message
	}
	return payload
}
