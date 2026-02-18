package handlers

import (
	"time"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// buildNewMaintenanceRecord creates a new MaintenanceRecord model from the input data map and vehicle ID.
func buildNewMaintenanceRecord(d map[string]interface{}, vehicleID string) (*models.MaintenanceRecord, error) {
	serviceType, _ := d["service_type"].(string)
	serviceDateStr, _ := d["service_date"].(string)

	serviceDate, err := time.Parse("2006-01-02", serviceDateStr)
	if err != nil {
		return nil, models.NewValidationError("service_date", "invalid date format, expected YYYY-MM-DD")
	}

	rec := &models.MaintenanceRecord{
		VehicleID:   vehicleID,
		ServiceType: serviceType,
		ServiceDate: serviceDate,
	}

	if v, ok := d["mileage_at_service"].(float64); ok {
		i := int(v)
		rec.MileageAtService = &i
	}
	if v, ok := d["cost"].(float64); ok {
		rec.Cost = &v
	}
	if v, ok := d["service_provider"].(string); ok {
		rec.ServiceProvider = v
	}
	if v, ok := d["notes"].(string); ok {
		rec.Notes = v
	}

	return rec, nil
}

// extractMaintenanceChanges extracts maintenance update fields from the input data map.
func extractMaintenanceChanges(d map[string]interface{}) (services.MaintenanceUpdates, error) {
	var u services.MaintenanceUpdates
	if v, ok := d["service_type"].(string); ok && v != "" {
		u.ServiceType = &v
	}
	if v, ok := d["service_date"].(string); ok && v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return u, models.NewValidationError("service_date", "invalid date format, expected YYYY-MM-DD")
		}
		u.ServiceDate = &t
	}
	if v, ok := d["mileage_at_service"].(float64); ok {
		i := int(v)
		u.MileageAtService = &i
	}
	if v, ok := d["cost"].(float64); ok {
		u.Cost = &v
	}
	if v, ok := d["service_provider"].(string); ok {
		u.ServiceProvider = &v
	}
	if v, ok := d["notes"].(string); ok {
		u.Notes = &v
	}
	return u, nil
}
