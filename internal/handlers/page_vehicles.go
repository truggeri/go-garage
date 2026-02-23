package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
)

// vehicleListPageData holds the data passed to the vehicle list template.
type vehicleListPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// Vehicles is the slice of vehicles to display on the current page.
	Vehicles []*models.Vehicle
	// TotalCount is the total number of vehicles matching the current filters.
	TotalCount int
	// Page is the current page number (1-based).
	Page int
	// PageSize is the number of vehicles per page.
	PageSize int
	// TotalPages is the total number of pages.
	TotalPages int
	// FilterMake is the current make filter value.
	FilterMake string
	// FilterStatus is the current status filter value.
	FilterStatus string
	// SortBy is the current sort field.
	SortBy string
}

const vehicleListPageSize = 12

// VehicleList serves the vehicle list page (GET /vehicles).
func (h *PageHandler) VehicleList(w http.ResponseWriter, r *http.Request) {
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

	filterMake := r.URL.Query().Get("make")
	filterStatus := r.URL.Query().Get("status")

	filters := repositories.VehicleFilters{UserID: &account.ID}
	if filterMake != "" {
		filters.Make = &filterMake
	}
	if filterStatus != "" {
		s := models.VehicleStatus(filterStatus)
		filters.Status = &s
	}

	totalCount, err := h.vehicleService.CountVehicles(r.Context(), filters)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(vehicleListPageSize)))
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * vehicleListPageSize
	vehicles, err := h.vehicleService.ListVehicles(r.Context(), filters, repositories.PaginationParams{
		Limit:  vehicleListPageSize,
		Offset: offset,
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := vehicleListPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		Vehicles:        vehicles,
		TotalCount:      totalCount,
		Page:            page,
		PageSize:        vehicleListPageSize,
		TotalPages:      totalPages,
		FilterMake:      filterMake,
		FilterStatus:    filterStatus,
	}

	if err := h.engine.Render(w, "vehicles/list.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
