package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// maintenanceRecordFormEntry holds the form values and per-record validation errors
// for a single maintenance record in the bulk creation form.
type maintenanceRecordFormEntry struct {
	ServiceType       string
	CustomServiceType string
	ServiceDate       string
	MileageAtService  string
	Cost              string
	ServiceProvider   string
	Notes             string
	Errors            map[string]string
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
	// ServiceTypes is the list of valid service type enum values for dropdowns.
	ServiceTypes []models.ServiceType
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
		ServiceTypes:    models.AllServiceTypes(),
	}

	if err := h.engine.Render(w, "maintenance/new.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
