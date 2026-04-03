package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

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
	// VehicleID is the selected vehicle.
	VehicleID string
	// FuelTypes is the list of valid fuel type enum values for dropdowns.
	FuelTypes []models.FuelType
	// Form field values for repopulating the form after a failed submission.
	FillDate              string
	Mileage               string
	Volume                string
	FuelType              string
	PartialFill           bool
	PricePerUnit          string
	OctaneRating          string
	Location              string
	Brand                 string
	Notes                 string
	CityDrivingPercentage string
	VehicleReportedMPG    string
}

// fuelFormParseResult holds the parsed results of the fuel form.
type fuelFormParseResult struct {
	FillDate              time.Time
	Mileage               int
	Volume                float64
	PricePerUnit          *float64
	OctaneRating          *int
	CityDrivingPercentage *int
	VehicleReportedMPG    *float64
	Errors                map[string]string
}

// parseFuelForm parses the raw string values submitted via the fuel form
// into their typed equivalents. Business validation is handled by
// models.ValidateFuelRecord.
func parseFuelForm(fillDateStr, mileageStr, volumeStr, pricePerUnitStr, octaneRatingStr, cityDrivingStr, reportedMPGStr string) fuelFormParseResult {
	result := fuelFormParseResult{Errors: make(map[string]string)}

	if fillDateStr != "" {
		t, err := time.Parse("2006-01-02", fillDateStr)
		if err != nil {
			result.Errors["fill_date"] = "Invalid date format"
		} else {
			result.FillDate = t
		}
	}

	if mileageStr != "" {
		m, err := strconv.Atoi(mileageStr)
		if err != nil {
			result.Errors["mileage"] = "Mileage must be a valid number"
		} else {
			result.Mileage = m
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

	if pricePerUnitStr != "" {
		p, err := strconv.ParseFloat(pricePerUnitStr, 64)
		if err != nil {
			result.Errors["price_per_unit"] = "Price must be a valid number"
		} else {
			result.PricePerUnit = &p
		}
	}

	if octaneRatingStr != "" {
		o, err := strconv.Atoi(octaneRatingStr)
		if err != nil {
			result.Errors["octane_rating"] = "Octane rating must be a valid number"
		} else {
			result.OctaneRating = &o
		}
	}

	if cityDrivingStr != "" {
		c, err := strconv.Atoi(cityDrivingStr)
		if err != nil {
			result.Errors["city_driving_percentage"] = "City driving percentage must be a valid number"
		} else {
			result.CityDrivingPercentage = &c
		}
	}

	if reportedMPGStr != "" {
		m, err := strconv.ParseFloat(reportedMPGStr, 64)
		if err != nil {
			result.Errors["vehicle_reported_mpg"] = "MPG must be a valid number"
		} else {
			result.VehicleReportedMPG = &m
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
		FuelTypes:       models.AllFuelTypes(),
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
	mileageStr := r.FormValue("mileage")
	volumeStr := r.FormValue("volume")
	fuelType := strings.TrimSpace(r.FormValue("fuel_type"))
	partialFill := r.FormValue("partial_fill") == "on"
	pricePerUnitStr := r.FormValue("price_per_unit")
	octaneRatingStr := r.FormValue("octane_rating")
	location := strings.TrimSpace(r.FormValue("location"))
	brand := strings.TrimSpace(r.FormValue("brand"))
	notes := r.FormValue("notes")
	cityDrivingStr := r.FormValue("city_driving_percentage")
	reportedMPGStr := r.FormValue("vehicle_reported_mpg")

	vehicles, err := h.vehicleService.GetUserVehicles(r.Context(), account.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	renderForm := func(status int, formErrors map[string]string) {
		if formErrors == nil {
			formErrors = make(map[string]string)
		}
		w.WriteHeader(status)
		data := fuelNewPageData{
			IsAuthenticated:       true,
			UserName:              account.Name,
			ActiveNav:             "fuel",
			Vehicles:              vehicles,
			VehicleNames:          buildVehicleNameMap(vehicles),
			Errors:                formErrors,
			CSRFToken:             middleware.GetCSRFToken(r.Context()),
			VehicleID:             vehicleID,
			FuelTypes:             models.AllFuelTypes(),
			FillDate:              fillDateStr,
			Mileage:               mileageStr,
			Volume:                volumeStr,
			FuelType:              fuelType,
			PartialFill:           partialFill,
			PricePerUnit:          pricePerUnitStr,
			OctaneRating:          octaneRatingStr,
			Location:              location,
			Brand:                 brand,
			Notes:                 notes,
			CityDrivingPercentage: cityDrivingStr,
			VehicleReportedMPG:    reportedMPGStr,
		}
		_ = h.engine.Render(w, "fuel/new.html", "base", data)
	}

	if !isOwnedVehicle(vehicleID, vehicles) {
		renderForm(http.StatusBadRequest, map[string]string{"vehicle_id": "Please select a valid vehicle"})
		return
	}

	parseResult := parseFuelForm(fillDateStr, mileageStr, volumeStr, pricePerUnitStr, octaneRatingStr, cityDrivingStr, reportedMPGStr)
	if len(parseResult.Errors) > 0 {
		renderForm(http.StatusBadRequest, parseResult.Errors)
		return
	}

	record := &models.FuelRecord{
		VehicleID:             vehicleID,
		FillDate:              parseResult.FillDate,
		Mileage:               parseResult.Mileage,
		Volume:                parseResult.Volume,
		FuelType:              fuelType,
		PartialFill:           partialFill,
		PricePerUnit:          parseResult.PricePerUnit,
		OctaneRating:          parseResult.OctaneRating,
		Location:              location,
		Brand:                 brand,
		Notes:                 notes,
		CityDrivingPercentage: parseResult.CityDrivingPercentage,
		VehicleReportedMPG:    parseResult.VehicleReportedMPG,
	}

	if valErr := models.ValidateFuelRecord(record); valErr != nil {
		var ve *models.ValidationError
		if models.IsValidationError(valErr, &ve) {
			renderForm(http.StatusBadRequest, map[string]string{ve.Field: ve.Message})
			return
		}
		renderForm(http.StatusBadRequest, map[string]string{"general": "Invalid form data. Please check your input."})
		return
	}

	if createErr := h.fuelService.CreateFuel(r.Context(), record); createErr != nil {
		renderForm(http.StatusInternalServerError, map[string]string{
			"general": fmt.Sprintf("Failed to add fuel record. Please try again."),
		})
		return
	}

	http.Redirect(w, r, "/fuel?added=true", http.StatusSeeOther)
}
