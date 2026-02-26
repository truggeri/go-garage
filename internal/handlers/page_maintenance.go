package handlers

import (
	"net/http"
	"strconv"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// maintenanceListPageData holds the data passed to the maintenance list template.
type maintenanceListPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// Records is the slice of maintenance records to display on the current page.
	Records []*models.MaintenanceRecord
	// Vehicles is the list of user's vehicles, used to populate the vehicle filter dropdown.
	Vehicles []*models.Vehicle
	// VehicleNames maps vehicle IDs to human-readable names.
	VehicleNames map[string]string
	// TotalCount is the total number of records matching the current filters.
	TotalCount int
	// Page is the current page number (1-based).
	Page int
	// PageSize is the number of records per page.
	PageSize int
	// TotalPages is the total number of pages.
	TotalPages int
	// FilterVehicleID is the current vehicle filter value.
	FilterVehicleID string
	// FilterServiceType is the current service type filter value.
	FilterServiceType string
}

const maintenanceListPageSize = 15

// MaintenanceList serves the maintenance list page (GET /maintenance).
func (h *PageHandler) MaintenanceList(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	page := 1
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}

	filterVehicleID := r.URL.Query().Get("vehicle")
	filterServiceType := r.URL.Query().Get("service_type")

	userVehicles, err := h.vehicleService.GetUserVehicles(r.Context(), account.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if filterVehicleID != "" && !isOwnedVehicle(filterVehicleID, userVehicles) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var records []*models.MaintenanceRecord
	var totalCount int

	if filterVehicleID != "" {
		records, totalCount, page, err = fetchVehicleMaintenanceRecords(h, r, filterVehicleID, filterServiceType, page)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		records, totalCount, page = fetchAllUserMaintenanceRecords(h, r, userVehicles, filterServiceType, page)
	}

	data := maintenanceListPageData{
		IsAuthenticated:   true,
		UserName:          account.Name,
		ActiveNav:         "maintenance",
		Records:           records,
		Vehicles:          userVehicles,
		VehicleNames:      buildVehicleNameMap(userVehicles),
		TotalCount:        totalCount,
		Page:              page,
		PageSize:          maintenanceListPageSize,
		TotalPages:        calcTotalPages(totalCount, maintenanceListPageSize),
		FilterVehicleID:   filterVehicleID,
		FilterServiceType: filterServiceType,
	}

	if r.URL.Query().Get("added") == queryTrue {
		data.Flash = []flashMessage{
			{Type: "success", Message: "Maintenance record added successfully."},
		}
	}

	if err := h.engine.Render(w, "maintenance/list.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
