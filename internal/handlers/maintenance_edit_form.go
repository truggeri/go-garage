package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

func (h *maintenanceAPIHandler) replaceFromForm(w http.ResponseWriter, r *http.Request, id, vehicleID string, data map[string]interface{}) {
	updates, errs := parseMaintenanceEditForm(data, vehicleID)
	if len(errs) > 0 {
		respondWithValidationProblems(w, "Invalid maintenance fields", errs)
		return
	}
	updated, err := h.svc.UpdateMaintenance(r.Context(), id, updates)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	w.Header().Set("HX-Redirect", fmt.Sprintf("/maintenance/%s?updated=true", id))
	respondWithPayload(w, http.StatusOK, buildMaintenanceSinglePayload(updated, "Maintenance record updated successfully"))
}

func parseMaintenanceEditForm(data map[string]interface{}, vehicleID string) (services.MaintenanceUpdates, []FieldError) {
	var updates services.MaintenanceUpdates
	values := make(map[string]string)
	var errs []FieldError
	for _, field := range []string{"service_type", "custom_service_type", "service_date", "mileage_at_service", "cost", "service_provider", "notes"} {
		value, ok := data[field].(string)
		if !ok {
			errs = append(errs, FieldError{Field: field, Message: "must be a valid string"})
			continue
		}
		values[field] = value
	}
	if len(errs) > 0 {
		return updates, errs
	}
	parsed := parseMaintenanceNewForm(values["service_date"], values["mileage_at_service"], values["cost"])
	for field, message := range parsed.Errors {
		errs = append(errs, FieldError{Field: field, Message: message})
	}
	if parsed.Cost != nil && (math.IsNaN(*parsed.Cost) || math.IsInf(*parsed.Cost, 0)) {
		errs = append(errs, FieldError{Field: "cost", Message: "Cost must be a valid number"})
	}
	if len(errs) > 0 {
		return updates, errs
	}
	customType := ""
	serviceType := strings.TrimSpace(values["service_type"])
	if serviceType == string(models.ServiceTypeOther) {
		customType = strings.TrimSpace(values["custom_service_type"])
	}
	record := &models.MaintenanceRecord{
		VehicleID: vehicleID, ServiceType: serviceType,
		CustomServiceType: customType, ServiceDate: parsed.ServiceDate,
		MileageAtService: parsed.MileageAtService, Cost: parsed.Cost,
		ServiceProvider: strings.TrimSpace(values["service_provider"]), Notes: values["notes"],
	}
	if err := models.ValidateMaintenanceRecord(record); err != nil {
		var validation *models.ValidationError
		if models.IsValidationError(err, &validation) {
			return updates, []FieldError{{Field: validation.Field, Message: validation.Message}}
		}
		return updates, []FieldError{{Field: keyGeneral, Message: err.Error()}}
	}
	return services.MaintenanceUpdates{
		ServiceType: &record.ServiceType, CustomServiceType: &record.CustomServiceType,
		ServiceDate: &record.ServiceDate, MileageAtService: record.MileageAtService,
		Cost: record.Cost, ServiceProvider: &record.ServiceProvider, Notes: &record.Notes,
		ClearMileageAtService: record.MileageAtService == nil, ClearCost: record.Cost == nil,
	}, nil
}
