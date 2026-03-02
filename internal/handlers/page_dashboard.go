package handlers

import (
	"net/http"
	"sort"
	"time"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// dashboardPageData holds the data passed to the dashboard template.
type dashboardPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// VehicleCount is the total number of vehicles belonging to the user.
	VehicleCount int
	// MaintenanceCount is the total number of maintenance records across all vehicles.
	MaintenanceCount int
	// TotalSpent is the sum of all maintenance costs.
	TotalSpent float64
	// RecentMaintenance holds the five most recent maintenance records.
	RecentMaintenance []dashboardMaintenanceRow
}

// dashboardMaintenanceRow is a pre-processed row for the recent-maintenance table.
type dashboardMaintenanceRow struct {
	VehicleName string
	ServiceType string
	ServiceDate time.Time
	Cost        *float64
}

// Dashboard serves the main dashboard page (GET /dashboard).
// It expects the CookieAuthGuard middleware to have already validated the session
// and stored the account in the request context.
func (h *PageHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
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

	// Build a name lookup map and collect all maintenance records.
	vehicleNames := make(map[string]string, len(vehicles))
	for _, v := range vehicles {
		vehicleNames[v.ID] = vehicleName(v)
	}

	var allMaintenance []*models.MaintenanceRecord
	totalSpent := 0.0
	for _, v := range vehicles {
		records, svcErr := h.maintenanceService.GetVehicleMaintenance(r.Context(), v.ID)
		if svcErr != nil {
			// Best-effort: skip this vehicle's records rather than failing the whole dashboard.
			continue
		}
		for _, rec := range records {
			allMaintenance = append(allMaintenance, rec)
			if rec.Cost != nil {
				totalSpent += *rec.Cost
			}
		}
	}

	// Sort by most-recent service date first.
	sort.Slice(allMaintenance, func(i, j int) bool {
		return allMaintenance[i].ServiceDate.After(allMaintenance[j].ServiceDate)
	})

	totalCount := len(allMaintenance)

	const recentLimit = 5
	if len(allMaintenance) > recentLimit {
		allMaintenance = allMaintenance[:recentLimit]
	}

	rows := make([]dashboardMaintenanceRow, len(allMaintenance))
	for i, rec := range allMaintenance {
		rows[i] = dashboardMaintenanceRow{
			VehicleName: vehicleNames[rec.VehicleID],
			ServiceType: rec.ServiceType,
			ServiceDate: rec.ServiceDate,
			Cost:        rec.Cost,
		}
	}

	data := dashboardPageData{
		IsAuthenticated:   true,
		UserName:          account.Name,
		ActiveNav:         "dashboard",
		VehicleCount:      len(vehicles),
		MaintenanceCount:  totalCount,
		TotalSpent:        totalSpent,
		RecentMaintenance: rows,
	}

	if err := h.engine.Render(w, "dashboard.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// vehicleName returns a short human-readable name for a vehicle.
func vehicleName(v *models.Vehicle) string {
	if v == nil {
		return "Unknown"
	}
	if v.DisplayName != "" {
		return v.DisplayName
	}
	return v.Make + " " + v.Model
}
