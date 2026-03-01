package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// maintenanceNewPageData holds the data passed to the add-maintenance template.
type maintenanceNewPageData struct {
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
	VehicleID        string
	ServiceType      string
	ServiceDate      string
	MileageAtService string
	Cost             string
	ServiceProvider  string
	Notes            string
}

// maintenanceNewFormResult holds the parsed results of the add-maintenance form.
type maintenanceNewFormResult struct {
	ServiceDate      time.Time
	MileageAtService *int
	Cost             *float64
	Errors           map[string]string
}

// parseMaintenanceNewForm parses the raw string values submitted via the
// add-maintenance form into their typed equivalents. It returns a struct
// containing parsed values and a map of field-level error messages for any
// inputs that cannot be converted to the expected type. Business validation
// (e.g. required fields, value constraints) is handled by
// models.ValidateMaintenanceRecord.
func parseMaintenanceNewForm(serviceDateStr, mileageStr, costStr string) maintenanceNewFormResult {
	result := maintenanceNewFormResult{Errors: make(map[string]string)}

	if serviceDateStr != "" {
		t, err := time.Parse("2006-01-02", serviceDateStr)
		if err != nil {
			result.Errors["service_date"] = "Invalid date format"
		} else {
			result.ServiceDate = t
		}
	}

	if mileageStr != "" {
		m, err := strconv.Atoi(mileageStr)
		if err != nil {
			result.Errors["mileage_at_service"] = "Mileage must be a valid number"
		} else {
			result.MileageAtService = &m
		}
	}

	if costStr != "" {
		c, err := strconv.ParseFloat(costStr, 64)
		if err != nil {
			result.Errors["cost"] = "Cost must be a valid number"
		} else {
			result.Cost = &c
		}
	}

	return result
}

// MaintenanceNew serves the add maintenance record form page (GET /maintenance/new).
func (h *PageHandler) MaintenanceNew(w http.ResponseWriter, r *http.Request) {
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

	data := maintenanceNewPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		ActiveNav:       "maintenance",
		Vehicles:        vehicles,
		VehicleNames:    buildVehicleNameMap(vehicles),
		VehicleID:       r.URL.Query().Get("vehicle"),
		CSRFToken:       middleware.GetCSRFToken(r.Context()),
	}

	if err := h.engine.Render(w, "maintenance/new.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// MaintenanceCreate handles the add maintenance record form submission (POST /maintenance/new).
func (h *PageHandler) MaintenanceCreate(w http.ResponseWriter, r *http.Request) {
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
	serviceType := strings.TrimSpace(r.FormValue("service_type"))
	serviceDateStr := r.FormValue("service_date")
	mileageStr := r.FormValue("mileage_at_service")
	costStr := r.FormValue("cost")
	serviceProvider := strings.TrimSpace(r.FormValue("service_provider"))
	notes := r.FormValue("notes")

	vehicles, err := h.vehicleService.GetUserVehicles(r.Context(), account.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	renderForm := func(status int, formErrors map[string]string) {
		w.WriteHeader(status)
		data := maintenanceNewPageData{
			IsAuthenticated:  true,
			UserName:         account.Name,
			ActiveNav:        "maintenance",
			Vehicles:         vehicles,
			VehicleNames:     buildVehicleNameMap(vehicles),
			Errors:           formErrors,
			CSRFToken:        middleware.GetCSRFToken(r.Context()),
			VehicleID:        vehicleID,
			ServiceType:      serviceType,
			ServiceDate:      serviceDateStr,
			MileageAtService: mileageStr,
			Cost:             costStr,
			ServiceProvider:  serviceProvider,
			Notes:            notes,
		}
		_ = h.engine.Render(w, "maintenance/new.html", "base", data)
	}

	if !isOwnedVehicle(vehicleID, vehicles) {
		renderForm(http.StatusBadRequest, map[string]string{"vehicle_id": "Please select a valid vehicle"})
		return
	}

	parseResult := parseMaintenanceNewForm(serviceDateStr, mileageStr, costStr)
	if len(parseResult.Errors) > 0 {
		renderForm(http.StatusBadRequest, parseResult.Errors)
		return
	}

	record := &models.MaintenanceRecord{
		VehicleID:        vehicleID,
		ServiceType:      serviceType,
		ServiceDate:      parseResult.ServiceDate,
		MileageAtService: parseResult.MileageAtService,
		Cost:             parseResult.Cost,
		ServiceProvider:  serviceProvider,
		Notes:            notes,
	}

	if err := models.ValidateMaintenanceRecord(record); err != nil {
		var ve *models.ValidationError
		if models.IsValidationError(err, &ve) {
			renderForm(http.StatusBadRequest, map[string]string{ve.Field: ve.Message})
			return
		}
		renderForm(http.StatusBadRequest, map[string]string{"general": "Invalid form data. Please check your input."})
		return
	}

	if err := h.maintenanceService.CreateMaintenance(r.Context(), record); err != nil {
		renderForm(http.StatusInternalServerError, map[string]string{"general": "Failed to add maintenance record. Please try again."})
		return
	}

	http.Redirect(w, r, "/maintenance?added=true", http.StatusSeeOther)
}
