package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// errMsgInvalidDateFormat is the error message shown when a date field has an unexpected format.
const errMsgInvalidDateFormat = "Invalid date format"

// fuelNewPageData holds the data passed to the add-fuel template.
type fuelNewPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// Vehicles is the list of the user's vehicles, used to populate the vehicle dropdown.
	Vehicles []*models.Vehicle
	// VehicleNames maps vehicle IDs to human-readable names.
	VehicleNames map[string]string
	// Errors holds field-level and general validation error messages.
	Errors    map[string]string
	CSRFToken string
	// Form field values for repopulating the form after a failed submission.
	VehicleID      string
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

// fuelNewFormResult holds the parsed results of the add-fuel form.
type fuelNewFormResult struct {
	FillDate       time.Time
	Odometer       int
	CostPerUnit    float64
	Volume         float64
	CityDrivingPct *int
	ReportedMPG    *float64
	Errors         map[string]string
}

// parseFuelNewForm parses raw form values into typed equivalents.
// Business validation is handled by models.ValidateFuelRecord.
func parseFuelNewForm(fillDateStr, odometerStr, costPerUnitStr, volumeStr, cityDrivingPctStr, reportedMPGStr string) fuelNewFormResult {
	result := fuelNewFormResult{Errors: make(map[string]string)}

	if fillDateStr != "" {
		t, err := time.Parse("2006-01-02", fillDateStr)
		if err != nil {
			result.Errors["fill_date"] = errMsgInvalidDateFormat
		} else {
			result.FillDate = t
		}
	}

	if odometerStr != "" {
		o, err := strconv.Atoi(odometerStr)
		if err != nil {
			result.Errors["odometer"] = "Odometer must be a valid number"
		} else {
			result.Odometer = o
		}
	}

	if costPerUnitStr != "" {
		c, err := strconv.ParseFloat(costPerUnitStr, 64)
		if err != nil {
			result.Errors["cost_per_unit"] = "Cost per unit must be a valid number"
		} else {
			result.CostPerUnit = c
		}
	}

	if volumeStr != "" {
		v, err := strconv.ParseFloat(volumeStr, 64)
		if err != nil {
			result.Errors["volume"] = "Volume must be a valid number"
		} else {
			result.Volume = v
		}
	}

	if cityDrivingPctStr != "" {
		pct, err := strconv.Atoi(cityDrivingPctStr)
		if err != nil {
			result.Errors["city_driving_pct"] = "City driving percentage must be a valid number"
		} else {
			result.CityDrivingPct = &pct
		}
	}

	if reportedMPGStr != "" {
		mpg, err := strconv.ParseFloat(reportedMPGStr, 64)
		if err != nil {
			result.Errors["reported_mpg"] = "Reported MPG must be a valid number"
		} else {
			result.ReportedMPG = &mpg
		}
	}

	return result
}

// FuelNew serves the add fuel record form page (GET /fuel/new).
func (h *PageHandler) FuelNew(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	vehicles, err := h.vehicleService.GetUserVehicles(r.Context(), account.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	vehicleID := r.URL.Query().Get("vehicle")
	if vehicleID == "" {
		vehicleID = r.URL.Query().Get("vehicle_id")
	}

	data := fuelNewPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		ActiveNav:       "fuel",
		Vehicles:        vehicles,
		VehicleNames:    buildVehicleNameMap(vehicles),
		VehicleID:       vehicleID,
		CSRFToken:       middleware.GetCSRFToken(r.Context()),
	}

	if err := h.engine.Render(w, "fuel/new.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// FuelCreate handles the add fuel record form submission (POST /fuel/new).
func (h *PageHandler) FuelCreate(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	vehicleID := r.FormValue("vehicle_id")
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

	vehicles, err := h.vehicleService.GetUserVehicles(r.Context(), account.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	renderForm := func(status int, formErrors map[string]string) {
		w.WriteHeader(status)
		data := fuelNewPageData{
			IsAuthenticated: true,
			UserName:        account.Name,
			ActiveNav:       "fuel",
			Vehicles:        vehicles,
			VehicleNames:    buildVehicleNameMap(vehicles),
			Errors:          formErrors,
			CSRFToken:       middleware.GetCSRFToken(r.Context()),
			VehicleID:       vehicleID,
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
		_ = h.engine.Render(w, "fuel/new.html", "base", data)
	}

	if !isOwnedVehicle(vehicleID, vehicles) {
		renderForm(http.StatusBadRequest, map[string]string{"vehicle_id": "Please select a valid vehicle"})
		return
	}

	parseResult := parseFuelNewForm(fillDateStr, odometerStr, costPerUnitStr, volumeStr, cityDrivingPctStr, reportedMPGStr)
	if len(parseResult.Errors) > 0 {
		renderForm(http.StatusBadRequest, parseResult.Errors)
		return
	}

	record := &models.FuelRecord{
		VehicleID:      vehicleID,
		FillDate:       parseResult.FillDate,
		Odometer:       parseResult.Odometer,
		CostPerUnit:    parseResult.CostPerUnit,
		Volume:         parseResult.Volume,
		FuelType:       fuelType,
		CityDrivingPct: parseResult.CityDrivingPct,
		Location:       location,
		Brand:          brand,
		Notes:          notes,
		ReportedMPG:    parseResult.ReportedMPG,
		PartialFuelUp:  partialFuelUp,
	}

	if err := models.ValidateFuelRecord(record); err != nil {
		var ve *models.ValidationError
		if models.IsValidationError(err, &ve) {
			renderForm(http.StatusBadRequest, map[string]string{ve.Field: ve.Message})
			return
		}
		renderForm(http.StatusBadRequest, map[string]string{"general": "Invalid form data. Please check your input."})
		return
	}

	if err := h.fuelService.CreateFuelRecord(r.Context(), record); err != nil {
		renderForm(http.StatusInternalServerError, map[string]string{"general": "Failed to add fuel record. Please try again."})
		return
	}

	http.Redirect(w, r, "/fuel?added=true", http.StatusSeeOther)
}
