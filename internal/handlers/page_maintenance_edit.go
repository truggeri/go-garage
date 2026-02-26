package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// maintenanceEditPageData holds the data passed to the edit-maintenance template.
type maintenanceEditPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// RecordID is the ID of the maintenance record being edited.
	RecordID string
	// VehicleTitle is a short human-readable title for the vehicle (e.g. "2020 Ford Focus").
	VehicleTitle string
	// Errors holds field-level and general validation error messages.
	Errors map[string]string
	// Form field values for repopulating the form after a failed submission.
	ServiceType      string
	ServiceDate      string
	MileageAtService string
	Cost             string
	ServiceProvider  string
	Notes            string
}

// maintenanceEditPageDataFromRecord builds an edit page data struct pre-populated
// with the current values of the given maintenance record.
func maintenanceEditPageDataFromRecord(
	account *middleware.AccountInfo,
	record *models.MaintenanceRecord,
	vehicle *models.Vehicle,
) maintenanceEditPageData {
	data := maintenanceEditPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		ActiveNav:       "maintenance",
		RecordID:        record.ID,
		VehicleTitle:    fmt.Sprintf("%d %s %s", vehicle.Year, vehicle.Make, vehicle.Model),
		ServiceType:     record.ServiceType,
		ServiceDate:     record.ServiceDate.Format("2006-01-02"),
		ServiceProvider: record.ServiceProvider,
		Notes:           record.Notes,
	}
	if record.MileageAtService != nil {
		data.MileageAtService = fmt.Sprintf("%d", *record.MileageAtService)
	}
	if record.Cost != nil {
		data.Cost = fmt.Sprintf("%.2f", *record.Cost)
	}
	return data
}

// MaintenanceEdit serves the edit maintenance record form page (GET /maintenance/{id}/edit).
func (h *PageHandler) MaintenanceEdit(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	record, vehicle, err := h.getMaintenanceRecordAndVehicle(r)
	if err != nil {
		writeMaintenanceRecordError(w, err)
		return
	}

	if vehicle.UserID != account.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	data := maintenanceEditPageDataFromRecord(account, record, vehicle)

	if err := h.engine.Render(w, "maintenance/edit.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// MaintenanceUpdate handles the edit maintenance record form submission (POST /maintenance/{id}/edit).
func (h *PageHandler) MaintenanceUpdate(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	record, vehicle, err := h.getMaintenanceRecordAndVehicle(r)
	if err != nil {
		writeMaintenanceRecordError(w, err)
		return
	}

	if vehicle.UserID != account.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	serviceType := strings.TrimSpace(r.FormValue("service_type"))
	serviceDateStr := r.FormValue("service_date")
	mileageStr := r.FormValue("mileage_at_service")
	costStr := r.FormValue("cost")
	serviceProvider := strings.TrimSpace(r.FormValue("service_provider"))
	notes := r.FormValue("notes")

	vehicleTitle := fmt.Sprintf("%d %s %s", vehicle.Year, vehicle.Make, vehicle.Model)

	renderForm := func(status int, formErrors map[string]string) {
		w.WriteHeader(status)
		data := maintenanceEditPageData{
			IsAuthenticated:  true,
			UserName:         account.Name,
			ActiveNav:        "maintenance",
			RecordID:         record.ID,
			VehicleTitle:     vehicleTitle,
			Errors:           formErrors,
			ServiceType:      serviceType,
			ServiceDate:      serviceDateStr,
			MileageAtService: mileageStr,
			Cost:             costStr,
			ServiceProvider:  serviceProvider,
			Notes:            notes,
		}
		_ = h.engine.Render(w, "maintenance/edit.html", "base", data)
	}

	parseResult := parseMaintenanceNewForm(serviceDateStr, mileageStr, costStr)
	if len(parseResult.Errors) > 0 {
		renderForm(http.StatusBadRequest, parseResult.Errors)
		return
	}

	// Build a temporary record for validation.
	validationRecord := &models.MaintenanceRecord{
		VehicleID:        record.VehicleID,
		ServiceType:      serviceType,
		ServiceDate:      parseResult.ServiceDate,
		MileageAtService: parseResult.MileageAtService,
		Cost:             parseResult.Cost,
		ServiceProvider:  serviceProvider,
		Notes:            notes,
	}

	if err := models.ValidateMaintenanceRecord(validationRecord); err != nil {
		var ve *models.ValidationError
		if models.IsValidationError(err, &ve) {
			renderForm(http.StatusBadRequest, map[string]string{ve.Field: ve.Message})
			return
		}
		renderForm(http.StatusBadRequest, map[string]string{"general": "Invalid form data. Please check your input."})
		return
	}

	updates := services.MaintenanceUpdates{
		ServiceType:      &serviceType,
		ServiceDate:      &parseResult.ServiceDate,
		MileageAtService: parseResult.MileageAtService,
		Cost:             parseResult.Cost,
		ServiceProvider:  &serviceProvider,
		Notes:            &notes,
	}

	if _, err := h.maintenanceService.UpdateMaintenance(r.Context(), record.ID, updates); err != nil {
		renderForm(http.StatusInternalServerError, map[string]string{"general": "Failed to update maintenance record. Please try again."})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/maintenance/%s?updated=true", record.ID), http.StatusSeeOther)
}
