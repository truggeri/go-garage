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

// maxBulkRecords is the maximum number of maintenance records that can be submitted at once.
const maxBulkRecords = 50

// maintenanceRecordFormEntry holds the form values and per-record validation errors
// for a single maintenance record in the bulk creation form.
type maintenanceRecordFormEntry struct {
	ServiceType      string
	ServiceDate      string
	MileageAtService string
	Cost             string
	ServiceProvider  string
	Notes            string
	Errors           map[string]string
}

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
	// Errors holds vehicle-level and general validation error messages.
	Errors    map[string]string
	CSRFToken string
	// VehicleID is the selected vehicle for all records.
	VehicleID string
	// Records holds the form values and per-record errors for each maintenance record entry.
	Records []maintenanceRecordFormEntry
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

	vehicleID := r.URL.Query().Get("vehicle")
	if vehicleID == "" {
		vehicleID = r.URL.Query().Get("vehicle_id")
	}

	data := maintenanceNewPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		ActiveNav:       "maintenance",
		Vehicles:        vehicles,
		VehicleNames:    buildVehicleNameMap(vehicles),
		VehicleID:       vehicleID,
		CSRFToken:       middleware.GetCSRFToken(r.Context()),
		Records:         []maintenanceRecordFormEntry{{Errors: make(map[string]string)}},
	}

	if err := h.engine.Render(w, "maintenance/new.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// MaintenanceCreate handles the add maintenance record form submission (POST /maintenance/new).
// It supports bulk creation of multiple records for a single vehicle.
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

	vehicles, err := h.vehicleService.GetUserVehicles(r.Context(), account.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Read array form values for multiple records.
	serviceTypes := r.Form["service_type"]
	serviceDates := r.Form["service_date"]
	mileages := r.Form["mileage_at_service"]
	costs := r.Form["cost"]
	providers := r.Form["service_provider"]
	notesArr := r.Form["notes"]

	count := len(serviceTypes)
	if count == 0 {
		count = 1
	}
	if count > maxBulkRecords {
		count = maxBulkRecords
	}

	// Build record form entries for potential re-rendering.
	records := make([]maintenanceRecordFormEntry, count)
	for i := range records {
		records[i] = maintenanceRecordFormEntry{
			ServiceType:      safeIndex(serviceTypes, i),
			ServiceDate:      safeIndex(serviceDates, i),
			MileageAtService: safeIndex(mileages, i),
			Cost:             safeIndex(costs, i),
			ServiceProvider:  safeIndex(providers, i),
			Notes:            safeIndex(notesArr, i),
			Errors:           make(map[string]string),
		}
	}

	renderForm := func(status int, generalErrors map[string]string) {
		if generalErrors == nil {
			generalErrors = make(map[string]string)
		}
		w.WriteHeader(status)
		data := maintenanceNewPageData{
			IsAuthenticated: true,
			UserName:        account.Name,
			ActiveNav:       "maintenance",
			Vehicles:        vehicles,
			VehicleNames:    buildVehicleNameMap(vehicles),
			Errors:          generalErrors,
			CSRFToken:       middleware.GetCSRFToken(r.Context()),
			VehicleID:       vehicleID,
			Records:         records,
		}
		_ = h.engine.Render(w, "maintenance/new.html", "base", data)
	}

	if !isOwnedVehicle(vehicleID, vehicles) {
		renderForm(http.StatusBadRequest, map[string]string{"vehicle_id": "Please select a valid vehicle"})
		return
	}

	// Parse and validate all records.
	hasErrors := false
	modelRecords := make([]*models.MaintenanceRecord, count)

	for i := range records {
		entry := &records[i]
		entry.ServiceType = strings.TrimSpace(entry.ServiceType)
		entry.ServiceProvider = strings.TrimSpace(entry.ServiceProvider)

		parseResult := parseMaintenanceNewForm(entry.ServiceDate, entry.MileageAtService, entry.Cost)
		if len(parseResult.Errors) > 0 {
			for k, v := range parseResult.Errors {
				entry.Errors[k] = v
			}
			hasErrors = true
			continue
		}

		record := &models.MaintenanceRecord{
			VehicleID:        vehicleID,
			ServiceType:      entry.ServiceType,
			ServiceDate:      parseResult.ServiceDate,
			MileageAtService: parseResult.MileageAtService,
			Cost:             parseResult.Cost,
			ServiceProvider:  entry.ServiceProvider,
			Notes:            entry.Notes,
		}

		if valErr := models.ValidateMaintenanceRecord(record); valErr != nil {
			var ve *models.ValidationError
			if models.IsValidationError(valErr, &ve) {
				entry.Errors[ve.Field] = ve.Message
			} else {
				entry.Errors["general"] = "Invalid form data. Please check your input."
			}
			hasErrors = true
			continue
		}

		modelRecords[i] = record
	}

	if hasErrors {
		renderForm(http.StatusBadRequest, nil)
		return
	}

	// Create all records.
	for i, record := range modelRecords {
		if createErr := h.maintenanceService.CreateMaintenance(r.Context(), record); createErr != nil {
			renderForm(http.StatusInternalServerError, map[string]string{
				"general": fmt.Sprintf("Failed to add record %d. Please try again.", i+1),
			})
			return
		}
	}

	http.Redirect(w, r, "/maintenance?added=true", http.StatusSeeOther)
}

// safeIndex returns the string at index i of the slice, or empty string if out of bounds.
func safeIndex(slice []string, i int) string {
	if i < len(slice) {
		return slice[i]
	}
	return ""
}
