package handlers

import (
	"math"
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
	// FilterFuelType is the current fuel type filter value.
	FilterFuelType string
	// FuelTypes is the list of valid fuel type enum values for filter dropdown.
	FuelTypes []models.FuelType
}

const fuelListPageSize = 15

// FuelList serves the fuel records list page (GET /fuel).
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
	filterFuelType := r.URL.Query().Get("fuel_type")

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
		records, totalCount, page, err = h.fetchVehicleFuelRecords(r, filterVehicleID, filterFuelType, page)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		records, totalCount, page = h.fetchAllUserFuelRecords(r, userVehicles, filterFuelType, page)
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
		FilterFuelType:  filterFuelType,
		FuelTypes:       models.AllFuelTypes(),
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

// fetchVehicleFuelRecords retrieves a paginated page of fuel records for a
// specific vehicle using DB-level filtering and pagination.
func (h *PageHandler) fetchVehicleFuelRecords(
	r *http.Request,
	vehicleID, fuelType string,
	page int,
) ([]*models.FuelRecord, int, int, error) {
	filters := repositories.FuelFilters{VehicleID: &vehicleID}
	if fuelType != "" {
		filters.FuelType = &fuelType
	}

	totalCount, err := h.fuelService.CountFuel(r.Context(), filters)
	if err != nil {
		return nil, 0, page, err
	}

	totalPages := calcTotalPages(totalCount, fuelListPageSize)
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * fuelListPageSize
	records, err := h.fuelService.ListFuel(r.Context(), filters, repositories.PaginationParams{
		Limit:  fuelListPageSize,
		Offset: offset,
	})
	return records, totalCount, page, err
}

// fetchAllUserFuelRecords retrieves all fuel records across a user's vehicles,
// optionally filters by fuel type, sorts by most-recent fill date, and paginates in memory.
func (h *PageHandler) fetchAllUserFuelRecords(
	r *http.Request,
	userVehicles []*models.Vehicle,
	fuelType string,
	page int,
) ([]*models.FuelRecord, int, int) {
	var all []*models.FuelRecord
	for _, v := range userVehicles {
		recs, err := h.fuelService.GetVehicleFuel(r.Context(), v.ID)
		if err != nil {
			continue
		}
		all = append(all, recs...)
	}

	if fuelType != "" {
		filtered := make([]*models.FuelRecord, 0, len(all))
		for _, rec := range all {
			if rec.FuelType == fuelType {
				filtered = append(filtered, rec)
			}
		}
		all = filtered
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].FillDate.After(all[j].FillDate)
	})

	totalCount := len(all)
	totalPages := int(math.Ceil(float64(totalCount) / float64(fuelListPageSize)))
	if totalPages == 0 {
		totalPages = 1
	}
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
