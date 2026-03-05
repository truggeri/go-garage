package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// fuelEditPageData holds the data passed to the edit-fuel template.
type fuelEditPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// RecordID is the ID of the fuel record being edited.
	RecordID string
	// VehicleTitle is a short human-readable title for the vehicle (e.g. "2020 Ford Focus").
	VehicleTitle string
	// Errors holds field-level and general validation error messages.
	Errors map[string]string
	// CSRFToken is the CSRF protection token to embed in the form.
	CSRFToken string
	// Form field values for repopulating the form after a failed submission.
	FillDate       string
	Odometer       string
	CostPerUnit    string
	Volume         string
	FuelType       string
	CityDrivingPct string
	Location       string
	Brand          string
	Notes          string
	ReportedMPG    string
	PartialFuelUp  bool
}

// fuelEditPageDataFromRecord builds an edit page data struct pre-populated
// with the current values of the given fuel record.
func fuelEditPageDataFromRecord(
	account *middleware.AccountInfo,
	record *models.FuelRecord,
	vehicle *models.Vehicle,
) fuelEditPageData {
	data := fuelEditPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		ActiveNav:       "fuel",
		RecordID:        record.ID,
		VehicleTitle:    fmt.Sprintf("%d %s %s", vehicle.Year, vehicle.Make, vehicle.Model),
		FillDate:        record.FillDate.Format("2006-01-02"),
		Odometer:        fmt.Sprintf("%d", record.Odometer),
		CostPerUnit:     fmt.Sprintf("%.3f", record.CostPerUnit),
		Volume:          fmt.Sprintf("%.3f", record.Volume),
		FuelType:        record.FuelType,
		Location:        record.Location,
		Brand:           record.Brand,
		Notes:           record.Notes,
		PartialFuelUp:   record.PartialFuelUp,
	}
	if record.CityDrivingPct != nil {
		data.CityDrivingPct = fmt.Sprintf("%d", *record.CityDrivingPct)
	}
	if record.ReportedMPG != nil {
		data.ReportedMPG = fmt.Sprintf("%.2f", *record.ReportedMPG)
	}
	return data
}

// FuelEdit serves the edit fuel record form page (GET /fuel/{id}/edit).
func (h *PageHandler) FuelEdit(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	record, vehicle, err := h.getFuelRecordAndVehicle(r)
	if err != nil {
		writeFuelRecordError(w, err)
		return
	}

	if vehicle.UserID != account.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	data := fuelEditPageDataFromRecord(account, record, vehicle)
	data.CSRFToken = middleware.GetCSRFToken(r.Context())

	if err := h.engine.Render(w, "fuel/edit.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// FuelUpdate handles the edit fuel record form submission (POST /fuel/{id}/edit).
func (h *PageHandler) FuelUpdate(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	record, vehicle, err := h.getFuelRecordAndVehicle(r)
	if err != nil {
		writeFuelRecordError(w, err)
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

	fillDateStr := r.FormValue("fill_date")
	odometerStr := r.FormValue("odometer")
	costPerUnitStr := r.FormValue("cost_per_unit")
	volumeStr := r.FormValue("volume")
	fuelType := strings.TrimSpace(r.FormValue("fuel_type"))
	cityDrivingPctStr := r.FormValue("city_driving_pct")
	location := strings.TrimSpace(r.FormValue("location"))
	brand := strings.TrimSpace(r.FormValue("brand"))
	notes := r.FormValue("notes")
	reportedMPGStr := r.FormValue("reported_mpg")
	partialFuelUp := r.FormValue("partial_fuel_up") == "on"

	vehicleTitle := fmt.Sprintf("%d %s %s", vehicle.Year, vehicle.Make, vehicle.Model)

	renderForm := func(status int, formErrors map[string]string) {
		w.WriteHeader(status)
		data := fuelEditPageData{
			IsAuthenticated: true,
			UserName:        account.Name,
			ActiveNav:       "fuel",
			RecordID:        record.ID,
			VehicleTitle:    vehicleTitle,
			Errors:          formErrors,
			CSRFToken:       middleware.GetCSRFToken(r.Context()),
			FillDate:        fillDateStr,
			Odometer:        odometerStr,
			CostPerUnit:     costPerUnitStr,
			Volume:          volumeStr,
			FuelType:        fuelType,
			CityDrivingPct:  cityDrivingPctStr,
			Location:        location,
			Brand:           brand,
			Notes:           notes,
			ReportedMPG:     reportedMPGStr,
			PartialFuelUp:   partialFuelUp,
		}
		_ = h.engine.Render(w, "fuel/edit.html", "base", data)
	}

	parseResult := parseFuelNewForm(fillDateStr, odometerStr, costPerUnitStr, volumeStr, cityDrivingPctStr, reportedMPGStr)
	if len(parseResult.Errors) > 0 {
		renderForm(http.StatusBadRequest, parseResult.Errors)
		return
	}

	// Build a temporary record for validation.
	validationRecord := &models.FuelRecord{
		VehicleID:      record.VehicleID,
		FillDate:       parseResult.FillDate,
		Odometer:       parseResult.Odometer,
		CostPerUnit:    parseResult.CostPerUnit,
		Volume:         parseResult.Volume,
		CityDrivingPct: parseResult.CityDrivingPct,
	}

	if err := models.ValidateFuelRecord(validationRecord); err != nil {
		var ve *models.ValidationError
		if models.IsValidationError(err, &ve) {
			renderForm(http.StatusBadRequest, map[string]string{ve.Field: ve.Message})
			return
		}
		renderForm(http.StatusBadRequest, map[string]string{"general": "Invalid form data. Please check your input."})
		return
	}

	updates := services.FuelUpdates{
		FillDate:       &parseResult.FillDate,
		Odometer:       &parseResult.Odometer,
		CostPerUnit:    &parseResult.CostPerUnit,
		Volume:         &parseResult.Volume,
		FuelType:       &fuelType,
		CityDrivingPct: parseResult.CityDrivingPct,
		Location:       &location,
		Brand:          &brand,
		Notes:          &notes,
		ReportedMPG:    parseResult.ReportedMPG,
		PartialFuelUp:  &partialFuelUp,
	}

	if _, err := h.fuelService.UpdateFuelRecord(r.Context(), record.ID, updates); err != nil {
		renderForm(http.StatusInternalServerError, map[string]string{"general": "Failed to update fuel record. Please try again."})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/fuel/%s?updated=true", record.ID), http.StatusSeeOther)
}
