package handlers

import (
	"time"

	"github.com/truggeri/go-garage/internal/models"
)

// maintenanceToResponseMap converts a MaintenanceRecord model to a response map for JSON encoding.
func maintenanceToResponseMap(m *models.MaintenanceRecord) map[string]interface{} {
	r := map[string]interface{}{
		"id":           m.ID,
		"vehicle_id":   m.VehicleID,
		"service_type": m.ServiceType,
		"service_date": m.ServiceDate.Format("2006-01-02"),
		"created_at":   m.CreatedAt.Format(time.RFC3339),
		"updated_at":   m.UpdatedAt.Format(time.RFC3339),
	}
	if m.MileageAtService != nil {
		r["mileage_at_service"] = *m.MileageAtService
	}
	if m.Cost != nil {
		r["cost"] = *m.Cost
	}
	if m.ServiceProvider != "" {
		r["service_provider"] = m.ServiceProvider
	}
	if m.Notes != "" {
		r["notes"] = m.Notes
	}
	return r
}

// buildMaintenanceListPayload creates a paginated list response payload for maintenance records.
func buildMaintenanceListPayload(recs []*models.MaintenanceRecord, pg, sz, total int) map[string]interface{} {
	items := make([]map[string]interface{}, len(recs))
	for i, m := range recs {
		items[i] = maintenanceToResponseMap(m)
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

// buildMaintenanceSinglePayload creates a single maintenance record response payload.
func buildMaintenanceSinglePayload(m *models.MaintenanceRecord, msg string) map[string]interface{} {
	p := map[string]interface{}{"success": true, "data": maintenanceToResponseMap(m)}
	if msg != "" {
		p["message"] = msg
	}
	return p
}
