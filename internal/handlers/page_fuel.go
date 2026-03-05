package handlers

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
)

// fuelListPageData holds the data passed to the fuel list template.
type fuelListPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// Records is the slice of fuel records to display on the current page.
	Records []*models.FuelRecord
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
}

const fuelListPageSize = 15

// FuelList serves the fuel list page (GET /fuel).
func (h *PageHandler) FuelList(w http.ResponseWriter, r *http.Request) {
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

	userVehicles, err := h.vehicleService.GetUserVehicles(r.Context(), account.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if filterVehicleID != "" && !isOwnedVehicle(filterVehicleID, userVehicles) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var records []*models.FuelRecord
	var totalCount int

	if filterVehicleID != "" {
		records, totalCount, page, err = fetchVehicleFuelRecords(h, r, filterVehicleID, page)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		records, totalCount, page = fetchAllUserFuelRecords(h, r, userVehicles, page)
	}

	data := fuelListPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		ActiveNav:       "fuel",
		Records:         records,
		Vehicles:        userVehicles,
		VehicleNames:    buildVehicleNameMap(userVehicles),
		TotalCount:      totalCount,
		Page:            page,
		PageSize:        fuelListPageSize,
		TotalPages:      calcTotalPages(totalCount, fuelListPageSize),
		FilterVehicleID: filterVehicleID,
	}

	if r.URL.Query().Get("added") == queryTrue {
		data.Flash = []flashMessage{
			{Type: "success", Message: "Fuel record added successfully."},
		}
	}

	if err := h.engine.Render(w, "fuel/list.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// fetchVehicleFuelRecords retrieves a paginated page of fuel records for a specific vehicle.
func fetchVehicleFuelRecords(
	h *PageHandler, r *http.Request,
	vehicleID string,
	page int,
) ([]*models.FuelRecord, int, int, error) {
	filters := repositories.FuelFilters{VehicleID: &vehicleID}

	totalCount, err := h.fuelService.CountFuelRecords(r.Context(), filters)
	if err != nil {
		return nil, 0, page, err
	}

	totalPages := calcTotalPages(totalCount, fuelListPageSize)
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * fuelListPageSize
	records, err := h.fuelService.ListFuelRecords(r.Context(), filters, repositories.PaginationParams{
		Limit:  fuelListPageSize,
		Offset: offset,
	})
	return records, totalCount, page, err
}

// fetchAllUserFuelRecords retrieves all fuel records across a user's vehicles,
// sorts by most-recent fill date, and paginates in memory.
func fetchAllUserFuelRecords(
	h *PageHandler, r *http.Request,
	userVehicles []*models.Vehicle,
	page int,
) ([]*models.FuelRecord, int, int) {
	var all []*models.FuelRecord
	for _, v := range userVehicles {
		recs, err := h.fuelService.GetVehicleFuelRecords(r.Context(), v.ID)
		if err != nil {
			// Skip vehicles whose records cannot be loaded so that the page
			// still renders with partial data rather than failing completely.
			// This mirrors the behaviour of fetchAllUserMaintenanceRecords.
			continue
		}
		all = append(all, recs...)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].FillDate.After(all[j].FillDate)
	})

	totalCount := len(all)
	totalPages := calcTotalPages(totalCount, fuelListPageSize)
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * fuelListPageSize
	end := offset + fuelListPageSize
	if end > len(all) {
		end = len(all)
	}

	var records []*models.FuelRecord
	if offset < len(all) {
		records = all[offset:end]
	}
	return records, totalCount, page
}
